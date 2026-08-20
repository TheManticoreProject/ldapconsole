package core

import (
	"reflect"
	"testing"

	goldap "github.com/go-ldap/ldap/v3"
)

func TestColumnValuesMatchesAttributeNamesCaseInsensitively(t *testing.T) {
	entry := goldap.NewEntry("CN=Administrator", map[string][]string{
		"sAMAccountName": {"Administrator"},
	})

	got := columnValues(entry, []string{"samaccountname"})
	want := []string{"Administrator"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("columnValues = %#v, want %#v", got, want)
	}
}
