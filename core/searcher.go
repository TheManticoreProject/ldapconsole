package core

import (
	"fmt"
	"sort"

	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/winacl/sid"
	goldap "github.com/go-ldap/ldap/v3"
)

// ANSI color codes used to render LDAP results.
const (
	colorReset = "\x1b[0m"
	colorGreen = "\x1b[92m"
	colorCyan  = "\x1b[96m"
	colorRed   = "\x1b[91m"
)

// LDAPSearcher wraps a Manticore LDAP session and exposes the search helpers
// used by both the interactive console and the non-interactive query mode.
type LDAPSearcher struct {
	Session *ldap.Session
	Debug   bool
}

// NewLDAPSearcher builds an LDAPSearcher bound to an established LDAP session.
func NewLDAPSearcher(session *ldap.Session, debug bool) *LDAPSearcher {
	return &LDAPSearcher{Session: session, Debug: debug}
}

// Query performs an LDAP search on the given search base with the provided scope.
func (ls *LDAPSearcher) Query(searchBase string, query string, attributes []string, scope int) ([]*goldap.Entry, error) {
	return ls.Session.Query(searchBase, query, attributes, scope)
}

// QueryAllNamingContexts runs the query against every naming context of the server.
func (ls *LDAPSearcher) QueryAllNamingContexts(query string, attributes []string, scope int) ([]*goldap.Entry, error) {
	return ls.Session.QueryAllNamingContexts(query, attributes, scope)
}

// formatAttributeValues renders the values of an attribute as human readable
// strings, decoding well-known binary attributes such as objectSid.
func formatAttributeValues(attr *goldap.EntryAttribute) []string {
	switch attr.Name {
	case "objectSid":
		values := make([]string, 0, len(attr.ByteValues))
		for _, raw := range attr.ByteValues {
			s := &sid.SID{}
			if _, err := s.Unmarshal(raw); err == nil {
				values = append(values, s.ToString())
			} else {
				values = append(values, fmt.Sprintf("%x", raw))
			}
		}
		return values
	case "objectGUID":
		values := make([]string, 0, len(attr.ByteValues))
		for _, raw := range attr.ByteValues {
			if len(raw) == 16 {
				g := guid.NewGUID()
				g.FromRawBytes(raw)
				values = append(values, g.ToFormatD())
			} else {
				values = append(values, fmt.Sprintf("%x", raw))
			}
		}
		return values
	default:
		return attr.Values
	}
}

// PrintColoredResult prints a single LDAP entry in a colored, tree-like format.
func PrintColoredResult(entry *goldap.Entry) {
	fmt.Printf("│ %s\n", entry.DN)
	for _, attr := range entry.Attributes {
		values := formatAttributeValues(attr)
		switch len(values) {
		case 0:
			fmt.Printf("    \"%s%s%s\": []\n", colorGreen, attr.Name, colorReset)
		case 1:
			fmt.Printf("    \"%s%s%s\": \"%s%s%s\",\n", colorGreen, attr.Name, colorReset, colorCyan, values[0], colorReset)
		default:
			fmt.Printf("    \"%s%s%s\": [\n", colorGreen, attr.Name, colorReset)
			for _, value := range values {
				fmt.Printf("        \"%s%s%s\",\n", colorCyan, value, colorReset)
			}
			fmt.Printf("    ],\n")
		}
	}
}

// PrintResults prints a list of LDAP entries, sorted by distinguished name.
func PrintResults(entries []*goldap.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DN < entries[j].DN
	})
	for _, entry := range entries {
		PrintColoredResult(entry)
	}
}
