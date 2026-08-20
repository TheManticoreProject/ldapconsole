package core

import (
	"reflect"
	"testing"

	goldap "github.com/go-ldap/ldap/v3"
)

func TestAttributeMapIgnoresMultivaluedAttributeOrder(t *testing.T) {
	first := goldap.NewEntry("CN=user1", map[string][]string{
		"memberOf": {"CN=GroupA", "CN=GroupB"},
	})
	second := goldap.NewEntry("CN=user1", map[string][]string{
		"memberOf": {"CN=GroupB", "CN=GroupA"},
	})

	firstAttributes := attributeMap(first)
	secondAttributes := attributeMap(second)
	if !reflect.DeepEqual(firstAttributes, secondAttributes) {
		t.Errorf("attribute maps differ: %#v != %#v", firstAttributes, secondAttributes)
	}
	wantOriginalOrder := []string{"CN=GroupA", "CN=GroupB"}
	if got := first.GetAttributeValues("memberOf"); !reflect.DeepEqual(got, wantOriginalOrder) {
		t.Errorf("attributeMap mutated entry values: %#v, want %#v", got, wantOriginalOrder)
	}
}
