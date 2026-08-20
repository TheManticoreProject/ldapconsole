package core

import (
	"reflect"
	"testing"

	goldap "github.com/go-ldap/ldap/v3"
)

func TestResolveColumnsWildcardOmitsDistinguishedName(t *testing.T) {
	entries := []*goldap.Entry{
		goldap.NewEntry("CN=user1", map[string][]string{
			"cn":                {"user1"},
			"distinguishedName": {"CN=user1"},
			"sn":                {"one"},
		}),
	}

	columns := resolveColumns(entries, []string{"*"})
	want := []string{"cn", "sn"}
	if !reflect.DeepEqual(columns, want) {
		t.Errorf("resolveColumns = %#v, want %#v", columns, want)
	}
}

func TestResolveColumnsExplicitDistinguishedName(t *testing.T) {
	columns := resolveColumns(nil, []string{"distinguishedName", "cn"})
	want := []string{"cn"}
	if !reflect.DeepEqual(columns, want) {
		t.Errorf("resolveColumns = %#v, want %#v", columns, want)
	}

	columns = resolveColumns(nil, []string{"distinguishedName"})
	if len(columns) != 0 {
		t.Errorf("resolveColumns = %#v, want no additional columns", columns)
	}
}
