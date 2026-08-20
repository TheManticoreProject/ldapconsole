package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/chzyer/readline"
	goldap "github.com/go-ldap/ldap/v3"
)

// command describes an interactive console command for the help listing.
type command struct {
	name        string
	description []string
}

var consoleCommands = []command{
	{"diff", []string{"Show the differences between the last two requests."}},
	{"exit", []string{"Exits the ldapconsole script."}},
	{"help", []string{"Displays this help message."}},
	{"infos", []string{"Get information about the remote ldap server."}},
	{"presetquery", []string{
		"Use a builtin preset query.",
		"Available preset queries are: " + strings.Join(PresetNames(), ", "),
	}},
	{"query", []string{
		"Perform a LDAP query.",
		"You can query specific attributes by adding 'select <attribute 1> ... <attribute n>'.",
		"Syntax: query <filter> [select <attribute 1> ... <attribute n>]",
	}},
	{"rootdse", []string{"Queries the rootDSE."}},
	{"searchbase", []string{
		"Sets the search base for the LDAP queries.",
		"The accepted values are: 'defaultNamingContext', 'configurationNamingContext' or a distinguishedName.",
	}},
	{"searchscope", []string{
		"Sets the search scope for the LDAP queries.",
		"The accepted values are: BASE, LEVEL, SUBTREE.",
	}},
}

// Console implements the interactive REPL of ldapconsole.
type Console struct {
	searcher    *LDAPSearcher
	searchBase  string
	searchScope int

	lastQueryResults []*goldap.Entry
	prevQueryResults []*goldap.Entry
}

// NewConsole creates an interactive console bound to a searcher and a default search base.
func NewConsole(searcher *LDAPSearcher, searchBase string) *Console {
	return &Console{
		searcher:    searcher,
		searchBase:  searchBase,
		searchScope: ldap.ScopeWholeSubtree,
	}
}

// prompt returns the colored prompt showing the current search base.
func (c *Console) prompt() string {
	return fmt.Sprintf("[\x1b[95m%s\x1b[0m]> ", c.searchBase)
}

// newCompleter builds the tab-completion tree for the interactive console.
func newCompleter() *readline.PrefixCompleter {
	presetItems := make([]readline.PrefixCompleterInterface, 0, len(PresetNames()))
	for _, name := range PresetNames() {
		presetItems = append(presetItems, readline.PcItem(name))
	}

	return readline.NewPrefixCompleter(
		readline.PcItem("query"),
		readline.PcItem("presetquery", presetItems...),
		readline.PcItem("rootdse"),
		readline.PcItem("searchbase"),
		readline.PcItem("searchscope",
			readline.PcItem("BASE"),
			readline.PcItem("LEVEL"),
			readline.PcItem("SUBTREE"),
		),
		readline.PcItem("infos"),
		readline.PcItem("diff"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)
}

// Run starts the interactive read-eval-print loop. It returns when the user
// exits or stdin is closed. Command history (with up/down arrow navigation)
// is persisted to a history file across sessions, and TAB triggers completion.
func (c *Console) Run() {
	historyFile := filepath.Join(os.TempDir(), ".ldapconsole_history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            c.prompt(),
		HistoryFile:       historyFile,
		AutoComplete:      newCompleter(),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		fmt.Printf("%s[!] Error initializing interactive console: %s%s\n", colorRed, err, colorReset)
		return
	}
	defer rl.Close()

	for {
		rl.SetPrompt(c.prompt())

		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			// Ctrl-C on a non-empty line clears it; on an empty line it exits.
			if len(line) == 0 {
				return
			}
			continue
		} else if err == io.EOF {
			// Ctrl-D closes the console.
			return
		}

		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}

		command := strings.ToLower(fields[0])
		arguments := fields[1:]

		if command == "exit" {
			return
		}
		c.dispatch(command, arguments)
	}
}

// dispatch routes a parsed command to its handler.
func (c *Console) dispatch(command string, arguments []string) {
	switch command {
	case "help":
		printConsoleHelp()
	case "infos":
		c.handleInfos()
	case "presetquery":
		c.handlePresetQuery(arguments)
	case "query":
		c.handleQuery(arguments)
	case "rootdse":
		c.handleRootDSE(arguments)
	case "searchbase":
		c.handleSearchBase(arguments)
	case "searchscope":
		c.handleSearchScope(arguments)
	case "diff":
		c.handleDiff()
	default:
		fmt.Println("Unknown command. Type \"help\" for help.")
	}
}

func (c *Console) handleInfos() {
	rootDSE, err := c.searcher.Session.GetRootDSE()
	if err != nil {
		fmt.Printf("%s[!] Error fetching RootDSE: %s%s\n", colorRed, err, colorReset)
		return
	}
	PrintColoredResult(rootDSE)
}

func (c *Console) handlePresetQuery(arguments []string) {
	if len(arguments) == 0 {
		fmt.Println("[!] Usage: presetquery <preset_name>")
		PrintPresetHelp()
		return
	}
	c.searcher.PerformPreset(arguments[0])
}

func (c *Console) handleQuery(arguments []string) {
	filter, attributes := parseSelect(arguments)
	if strings.TrimSpace(filter) == "" {
		fmt.Printf("%s[!] Empty query.%s\n", colorRed, colorReset)
		return
	}

	results, err := c.searcher.Query(c.searchBase, filter, attributes, c.searchScope)
	if err != nil {
		fmt.Printf("%s[!] %s%s\n", colorRed, err, colorReset)
		return
	}

	c.storeResults(results)
	PrintResults(results)
	fmt.Printf("└──> LDAP query returned %d results.\n", len(results))
}

func (c *Console) handleRootDSE(arguments []string) {
	_, attributes := parseSelect(append([]string{"(objectClass=*)"}, arguments...))

	results, err := c.searcher.Query("", "(objectClass=*)", attributes, ldap.ScopeBaseObject)
	if err != nil {
		fmt.Printf("%s[!] %s%s\n", colorRed, err, colorReset)
		return
	}

	c.storeResults(results)
	for _, entry := range results {
		entry.DN = "RootDSE"
		PrintColoredResult(entry)
	}
	fmt.Printf("└──> LDAP query returned %d results.\n", len(results))
}

func (c *Console) handleSearchBase(arguments []string) {
	if len(arguments) == 0 {
		fmt.Printf("[i] Current search base is: %s\n", c.searchBase)
		return
	}
	searchBase := strings.Join(arguments, " ")
	// Convert a dotted FQDN (e.g. "lab.local") to a distinguishedName.
	if strings.Contains(searchBase, ".") && !strings.Contains(strings.ToLower(searchBase), "dc=") {
		parts := strings.Split(searchBase, ".")
		for i, part := range parts {
			parts[i] = "DC=" + part
		}
		searchBase = strings.Join(parts, ",")
	}
	c.searchBase = searchBase
}

func (c *Console) handleSearchScope(arguments []string) {
	if len(arguments) == 0 {
		fmt.Println("[!] Usage: searchscope <BASE|LEVEL|SUBTREE>")
		return
	}
	switch strings.ToLower(arguments[0]) {
	case "base":
		c.searchScope = ldap.ScopeBaseObject
		fmt.Println("[i] Search scope set to BASE.")
	case "level":
		c.searchScope = ldap.ScopeSingleLevel
		fmt.Println("[i] Search scope set to LEVEL.")
	case "subtree":
		c.searchScope = ldap.ScopeWholeSubtree
		fmt.Println("[i] Search scope set to SUBTREE.")
	default:
		fmt.Printf("%s[!] Unknown search scope \"%s\". Accepted values: BASE, LEVEL, SUBTREE.%s\n", colorRed, arguments[0], colorReset)
	}
}

// handleDiff prints the differences between the last two query results.
func (c *Console) handleDiff() {
	prev := entriesByDN(c.prevQueryResults)
	last := entriesByDN(c.lastQueryResults)

	for dn := range prev {
		if _, ok := last[dn]; !ok {
			fmt.Printf("[!] key \"%s\" was deleted in last results.\n", dn)
		}
	}
	for dn := range last {
		if _, ok := prev[dn]; !ok {
			fmt.Printf("[!] key \"%s\" was added in last results.\n", dn)
		}
	}

	commonDNs := make([]string, 0)
	for dn := range prev {
		if _, ok := last[dn]; ok {
			commonDNs = append(commonDNs, dn)
		}
	}
	sort.Strings(commonDNs)

	for _, dn := range commonDNs {
		oldAttrs := attributeMap(prev[dn])
		newAttrs := attributeMap(last[dn])

		names := unionKeys(oldAttrs, newAttrs)
		printedDN := false
		for _, name := range names {
			oldValue, oldOk := oldAttrs[name]
			newValue, newOk := newAttrs[name]
			if oldOk && newOk && oldValue == newValue {
				continue
			}
			if !printedDN {
				fmt.Println(dn)
				printedDN = true
			}
			fmt.Printf("    \"%s%s%s\":\n", colorGreen, name, colorReset)
			if oldOk {
				fmt.Printf("      > Old value: %s\n", oldValue)
			} else {
				fmt.Println("      > Old value: None (attribute was not present in the last response)")
			}
			if newOk {
				fmt.Printf("      > New value: %s\n", newValue)
			} else {
				fmt.Println("      > New value: None (attribute is not present in the response)")
			}
		}
	}
}

// storeResults rotates the result history, keeping the last two query results.
func (c *Console) storeResults(results []*goldap.Entry) {
	c.prevQueryResults = c.lastQueryResults
	c.lastQueryResults = results
}

// parseSelect splits console arguments into an LDAP filter and a list of
// attributes following an optional "select" keyword.
func parseSelect(arguments []string) (string, []string) {
	selectIndex := -1
	for i, arg := range arguments {
		if strings.ToLower(arg) == "select" {
			selectIndex = i
			break
		}
	}

	if selectIndex == -1 {
		return strings.TrimSpace(strings.Join(arguments, " ")), []string{"*"}
	}

	filter := strings.TrimSpace(strings.Join(arguments[:selectIndex], " "))
	rawAttrs := arguments[selectIndex+1:]
	if len(rawAttrs) == 0 {
		return filter, []string{}
	}
	// Allow comma- or space-separated attribute lists.
	attributes := strings.FieldsFunc(strings.Join(rawAttrs, " "), func(r rune) bool {
		return r == ' ' || r == ','
	})
	return filter, attributes
}

// entriesByDN indexes entries by their distinguished name.
func entriesByDN(entries []*goldap.Entry) map[string]*goldap.Entry {
	index := make(map[string]*goldap.Entry, len(entries))
	for _, entry := range entries {
		index[entry.DN] = entry
	}
	return index
}

// attributeMap flattens an entry's attributes into a name -> joined-values map.
func attributeMap(entry *goldap.Entry) map[string]string {
	attrs := make(map[string]string, len(entry.Attributes))
	for _, attr := range entry.Attributes {
		values := append([]string(nil), formatAttributeValues(attr)...)
		sort.Strings(values)
		attrs[attr.Name] = strings.Join(values, ", ")
	}
	return attrs
}

// unionKeys returns the sorted union of keys of two maps.
func unionKeys(a, b map[string]string) []string {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// printConsoleHelp prints the list of interactive console commands.
func printConsoleHelp() {
	fmt.Println("│")
	for _, cmd := range consoleCommands {
		padded := cmd.name + " \x1b[90m" + strings.Repeat("─", max(1, 15-len(cmd.name))) + "\x1b[0m"
		if len(cmd.description) == 0 {
			fmt.Printf("│ ■ %s\x1b[90m┤\x1b[0m\n", padded)
		} else {
			fmt.Printf("│ ■ %s\x1b[90m┤\x1b[0m %s\n", padded, cmd.description[0])
			for _, extra := range cmd.description[1:] {
				fmt.Printf("│ %s\x1b[90m│\x1b[0m %s\n", strings.Repeat(" ", 18), extra)
			}
		}
		fmt.Println("│")
	}
}
