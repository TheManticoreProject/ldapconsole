package main

import (
	"strings"
	"testing"
)

func TestRunReturnsConnectionFailure(t *testing.T) {
	originalQuiet := quiet
	originalDomain := authDomain
	originalUsername := authUsername
	originalPassword := authPassword
	originalHashes := authHashes
	originalController := domainController
	originalPort := ldapPort
	originalLDAPS := useLdaps
	originalKerberos := useKerberos
	t.Cleanup(func() {
		quiet = originalQuiet
		authDomain = originalDomain
		authUsername = originalUsername
		authPassword = originalPassword
		authHashes = originalHashes
		domainController = originalController
		ldapPort = originalPort
		useLdaps = originalLDAPS
		useKerberos = originalKerberos
	})

	quiet = true
	authDomain = "invalid.local"
	authUsername = "test"
	authPassword = "test"
	authHashes = ""
	domainController = "127.0.0.1"
	ldapPort = 1
	useLdaps = false
	useKerberos = false

	err := run()
	if err == nil {
		t.Fatal("run() returned nil for a refused connection")
	}
	if !strings.Contains(err.Error(), "Error connecting to LDAP server") {
		t.Fatalf("run() error = %q, want connection failure", err)
	}
}
