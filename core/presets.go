package core

import (
	"fmt"
	"sort"

	"github.com/TheManticoreProject/Manticore/network/ldap"
)

// PresetQuery describes a builtin LDAP query exposed through the "presetquery" command.
type PresetQuery struct {
	Description string
	Filter      string
	Attributes  []string
}

// PresetQueries maps a preset name to its definition.
var PresetQueries = map[string]PresetQuery{
	"all_users": {
		Description: "Get the list of all users.",
		Filter:      "(&(objectCategory=person)(objectClass=user))",
		Attributes:  []string{"objectSid", "sAMAccountName"},
	},
	"all_descriptions": {
		Description: "Get the descriptions of all users.",
		Filter:      "(&(objectCategory=person)(objectClass=user)(description=*))",
		Attributes:  []string{"sAMAccountName", "description"},
	},
	"all_groups": {
		Description: "Get the list of all groups.",
		Filter:      "(objectClass=group)",
		Attributes:  []string{"distinguishedName"},
	},
	"all_computers": {
		Description: "Get the list of all computers.",
		Filter:      "(objectClass=computer)",
		Attributes:  []string{"distinguishedName"},
	},
	"all_organizational_units": {
		Description: "Get the list of all organizationalUnits.",
		Filter:      "(objectClass=organizationalUnit)",
		Attributes:  []string{"distinguishedName"},
	},
	"all_kerberoastables": {
		Description: "Get the list of all kerberoastable accounts.",
		Filter:      "(&(objectClass=user)(servicePrincipalName=*)(!(objectClass=computer))(!(cn=krbtgt))(!(userAccountControl:1.2.840.113556.1.4.803:=2)))",
		Attributes:  []string{"sAMAccountName", "servicePrincipalName"},
	},
}

// PresetNames returns the sorted list of available preset query names.
func PresetNames() []string {
	names := make([]string, 0, len(PresetQueries))
	for name := range PresetQueries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// PerformPreset executes the named preset query and prints the results.
func (ls *LDAPSearcher) PerformPreset(name string) {
	preset, ok := PresetQueries[name]
	if !ok {
		fmt.Printf("[!] Unknown preset query \"%s\". Here is a list of the available preset queries:\n", name)
		PrintPresetHelp()
		return
	}

	results, err := ls.QueryAllNamingContexts(preset.Filter, preset.Attributes, ldap.ScopeWholeSubtree)
	if err != nil {
		fmt.Printf("%s[!] Error performing preset query: %s%s\n", colorRed, err, colorReset)
		return
	}

	if len(results) == 0 {
		fmt.Printf("%sNo results.%s\n", colorRed, colorReset)
		return
	}

	PrintResults(results)
	fmt.Printf("└──> LDAP query returned %d results.\n", len(results))
}

// PrintPresetHelp prints the list of available preset queries and their filters.
func PrintPresetHelp() {
	for _, name := range PresetNames() {
		preset := PresetQueries[name]
		fmt.Printf(" - %-26s %s (LDAP Filter: %s)\n", name, preset.Description, preset.Filter)
	}
}
