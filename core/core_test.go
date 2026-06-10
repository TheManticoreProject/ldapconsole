package core

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	goldap "github.com/go-ldap/ldap/v3"
)

func TestParseSelect(t *testing.T) {
	tests := []struct {
		name       string
		arguments  []string
		wantFilter string
		wantAttrs  []string
	}{
		{
			name:       "no select returns wildcard",
			arguments:  []string{"(objectClass=user)"},
			wantFilter: "(objectClass=user)",
			wantAttrs:  []string{"*"},
		},
		{
			name:       "select with space separated attributes",
			arguments:  []string{"(objectClass=user)", "select", "cn", "sAMAccountName"},
			wantFilter: "(objectClass=user)",
			wantAttrs:  []string{"cn", "sAMAccountName"},
		},
		{
			name:       "select with comma separated attributes",
			arguments:  []string{"(objectClass=user)", "select", "cn,sAMAccountName"},
			wantFilter: "(objectClass=user)",
			wantAttrs:  []string{"cn", "sAMAccountName"},
		},
		{
			name:       "select with no attributes",
			arguments:  []string{"(objectClass=user)", "select"},
			wantFilter: "(objectClass=user)",
			wantAttrs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, attrs := parseSelect(tt.arguments)
			if filter != tt.wantFilter {
				t.Errorf("filter = %q, want %q", filter, tt.wantFilter)
			}
			if !reflect.DeepEqual(attrs, tt.wantAttrs) {
				t.Errorf("attrs = %#v, want %#v", attrs, tt.wantAttrs)
			}
		})
	}
}

func TestExportToXLSX(t *testing.T) {
	entries := []*goldap.Entry{
		goldap.NewEntry("CN=user1,DC=lab,DC=local", map[string][]string{
			"sAMAccountName": {"user1"},
			"memberOf":       {"CN=Admins,DC=lab,DC=local", "CN=Users,DC=lab,DC=local"},
		}),
		goldap.NewEntry("CN=user2,DC=lab,DC=local", map[string][]string{
			"sAMAccountName": {"user2"},
		}),
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out", "results.xlsx")

	if err := ExportToXLSX(path, entries, []string{"sAMAccountName", "memberOf"}); err != nil {
		t.Fatalf("ExportToXLSX returned error: %s", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected output file at %s: %s", path, err)
	}
}

func TestFormatAttributeValuesObjectGUID(t *testing.T) {
	// Mixed-endian objectGUID for {00112233-4455-6677-8899-aabbccddeeff}:
	// first three groups are little-endian, last two are big-endian.
	raw := []byte{0x33, 0x22, 0x11, 0x00, 0x55, 0x44, 0x77, 0x66, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	attr := &goldap.EntryAttribute{Name: "objectGUID", ByteValues: [][]byte{raw}}

	got := formatAttributeValues(attr)
	want := []string{"00112233-4455-6677-8899-aabbccddeeff"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("formatAttributeValues(objectGUID) = %#v, want %#v", got, want)
	}
}

func TestResolveColumnsWildcard(t *testing.T) {
	entries := []*goldap.Entry{
		goldap.NewEntry("CN=user1", map[string][]string{"cn": {"user1"}, "sn": {"one"}}),
		goldap.NewEntry("CN=user2", map[string][]string{"cn": {"user2"}, "mail": {"u2@lab"}}),
	}

	columns := resolveColumns(entries, []string{"*"})
	want := []string{"cn", "mail", "sn"}
	if !reflect.DeepEqual(columns, want) {
		t.Errorf("resolveColumns = %#v, want %#v", columns, want)
	}
}
