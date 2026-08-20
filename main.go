package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/goopts/parser"

	"github.com/TheManticoreProject/ldapconsole/core"
)

const VERSION = "2.1.0"

var (
	// Configuration
	debug bool
	quiet bool

	// Non-interactive mode
	query      string
	attributes []string
	xlsxOutput string

	// Authentication
	authDomain   string
	authUsername string
	authPassword string
	authHashes   string

	// LDAP Connection Settings
	domainController string
	ldapPort         int
	useLdaps         bool
	useKerberos      bool
)

func parseArgs() {
	ap := parser.ArgumentsParser{
		Banner: fmt.Sprintf("ldapconsole - by Remi GASCOU (Podalirius) @ TheManticoreProject - v%s", VERSION),
	}
	ap.SetOptShowBannerOnHelp(true)

	// Configuration flags
	group_config, err := ap.NewArgumentGroup("Configuration")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		group_config.NewBoolArgument(&debug, "", "--debug", false, "Debug mode.")
		group_config.NewBoolArgument(&quiet, "", "--quiet", false, "Quiet mode, do not print the banner.")
	}

	// Non-interactive query flags
	group_query, err := ap.NewArgumentGroup("Non-interactive query")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		group_query.NewStringArgument(&query, "-q", "--query", "", false, "LDAP query to perform. If set, ldapconsole runs in non-interactive mode.")
		group_query.NewListOfStringsArgument(&attributes, "-a", "--attribute", []string{}, false, "Attributes to extract. Can be specified multiple times.")
		group_query.NewStringArgument(&xlsxOutput, "-x", "--xlsx", "", false, "Output results of the query to an XLSX file.")
	}

	// LDAP Connection Settings
	group_ldapSettings, err := ap.NewArgumentGroup("LDAP Connection Settings")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		group_ldapSettings.NewStringArgument(&domainController, "-dc", "--dc-ip", "", true, "IP Address of the domain controller or KDC (Key Distribution Center) for Kerberos. If omitted, it will use the domain part (FQDN) specified in the identity parameter.")
		group_ldapSettings.NewTcpPortArgument(&ldapPort, "-lp", "--ldap-port", 389, false, "Port number to connect to LDAP server.")
		group_ldapSettings.NewBoolArgument(&useLdaps, "-L", "--use-ldaps", false, "Use LDAPS instead of LDAP.")
		group_ldapSettings.NewBoolArgument(&useKerberos, "-k", "--use-kerberos", false, "Use Kerberos instead of NTLM.")
	}

	// Authentication flags
	group_auth, err := ap.NewArgumentGroup("Authentication")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		group_auth.NewStringArgument(&authDomain, "-d", "--domain", "", true, "Active Directory domain to authenticate to.")
		group_auth.NewStringArgument(&authUsername, "-u", "--username", "", true, "User to authenticate as.")
		group_auth.NewStringArgument(&authPassword, "-p", "--password", "", false, "Password to authenticate with.")
		group_auth.NewStringArgument(&authHashes, "-H", "--hashes", "", false, "NT/LM hashes, format is LMhash:NThash.")
	}

	ap.Parse()
}

func main() {
	parseArgs()
	if err := run(); err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
	}
}

func run() error {
	if !quiet {
		fmt.Printf("ldapconsole v%s - by Remi GASCOU (Podalirius) @ TheManticoreProject\n\n", VERSION)
	}

	// Set up credentials
	creds, err := credentials.NewCredentials(authDomain, authUsername, authPassword, authHashes)
	if err != nil {
		return fmt.Errorf("Error creating credentials: %w", err)
	}

	// Use LDAPS default port if not explicitly set
	if useLdaps && ldapPort == 389 {
		ldapPort = 636
	}

	if !quiet {
		if authDomain != "" {
			fmt.Printf("[>] Trying to authenticate as \"%s\\%s\" on %s ...\n", authDomain, authUsername, domainController)
		} else {
			fmt.Printf("[>] Trying to authenticate as \"%s\" on %s ...\n", authUsername, domainController)
		}
	}

	// Init the LDAP session
	ldapSession, err := ldap.NewSession(domainController, ldapPort, creds, useLdaps, useKerberos)
	if err != nil {
		return fmt.Errorf("Error creating LDAP session: %w", err)
	}

	success, err := ldapSession.Connect()
	if !success {
		return fmt.Errorf("Error connecting to LDAP server: %w", err)
	}
	if !quiet {
		fmt.Printf("[+] Authentication successful!\n\n")
	}

	// Resolve the default search base from the RootDSE
	rootDSE, err := ldapSession.GetRootDSE()
	if err != nil {
		return fmt.Errorf("Error fetching RootDSE: %w", err)
	}
	searchBase := rootDSE.GetAttributeValue("defaultNamingContext")

	searcher := core.NewLDAPSearcher(ldapSession, debug)

	// Non-interactive mode: a single query was passed on the command line
	if strings.TrimSpace(query) != "" {
		queryAttributes := attributes
		if len(queryAttributes) == 0 {
			queryAttributes = []string{"*"}
		}

		results, err := searcher.Query(searchBase, query, queryAttributes, ldap.ScopeWholeSubtree)
		if err != nil {
			return fmt.Errorf("Error performing LDAP query: %w", err)
		}

		if xlsxOutput != "" {
			fmt.Printf("[>] Exporting %d results to %s ... ", len(results), xlsxOutput)
			if err := core.ExportToXLSX(xlsxOutput, results, queryAttributes); err != nil {
				fmt.Println()
				return fmt.Errorf("Error exporting to XLSX: %w", err)
			}
			fmt.Println("done.")
		} else {
			core.PrintResults(results)
			fmt.Printf("└──> LDAP query returned %d results.\n", len(results))
		}
		return nil
	}

	// Interactive mode
	console := core.NewConsole(searcher, searchBase)
	console.Run()

	return nil
}
