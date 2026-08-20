package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/xuri/excelize/v2"
)

// ExportToXLSX writes the LDAP search results to an XLSX file. When the
// requested attributes contain "*", the column set is computed from the union
// of all attributes present in the results.
func ExportToXLSX(path string, entries []*goldap.Entry, requestedAttributes []string) error {
	// Make sure the destination directory exists.
	basepath := filepath.Dir(path)
	if basepath != "." && basepath != "" {
		if err := os.MkdirAll(basepath, 0o755); err != nil {
			return fmt.Errorf("could not create output directory: %w", err)
		}
	}

	attributes := resolveColumns(entries, requestedAttributes)

	f := excelize.NewFile()
	defer f.Close()

	sheet := f.GetSheetName(0)

	// Header row: distinguishedName followed by the resolved attributes.
	header := append([]string{"distinguishedName"}, attributes...)
	for col, name := range header {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, name); err != nil {
			return err
		}
	}

	// Bold header style.
	headerStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err == nil {
		lastHeader, _ := excelize.CoordinatesToCellName(len(header), 1)
		_ = f.SetCellStyle(sheet, "A1", lastHeader, headerStyle)
	}

	// Data rows.
	for rowIdx, entry := range entries {
		row := rowIdx + 2
		data := append([]string{entry.DN}, columnValues(entry, attributes)...)
		for col, value := range data {
			cell, err := excelize.CoordinatesToCellName(col+1, row)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}

	// Add an autofilter over the populated range.
	lastCell, _ := excelize.CoordinatesToCellName(len(header), len(entries)+1)
	_ = f.AutoFilter(sheet, fmt.Sprintf("A1:%s", lastCell), []excelize.AutoFilterOptions{})

	return f.SaveAs(path)
}

// resolveColumns determines which attribute columns to export.
func resolveColumns(entries []*goldap.Entry, requestedAttributes []string) []string {
	hasWildcard := false
	explicit := make([]string, 0, len(requestedAttributes))
	for _, attr := range requestedAttributes {
		if attr == "*" {
			hasWildcard = true
			continue
		}
		explicit = append(explicit, attr)
	}

	if !hasWildcard && len(explicit) != 0 {
		return explicit
	}

	// Wildcard (or no attributes given): union of all attribute names present.
	seen := map[string]struct{}{}
	for _, attr := range explicit {
		seen[attr] = struct{}{}
	}
	for _, entry := range entries {
		for _, attr := range entry.Attributes {
			seen[attr.Name] = struct{}{}
		}
	}

	columns := make([]string, 0, len(seen))
	for name := range seen {
		columns = append(columns, name)
	}
	sort.Strings(columns)
	return columns
}

// columnValues extracts the values of the given attributes for an entry,
// joining multi-valued attributes with newlines.
func columnValues(entry *goldap.Entry, attributes []string) []string {
	values := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		var attribute *goldap.EntryAttribute
		for _, candidate := range entry.Attributes {
			if strings.EqualFold(candidate.Name, attr) {
				attribute = candidate
				break
			}
		}
		if attribute == nil {
			values = append(values, "")
			continue
		}
		values = append(values, strings.Join(formatAttributeValues(attribute), "\n"))
	}
	return values
}
