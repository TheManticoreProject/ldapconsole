![](./.github/banner.png)

<p align="center">
      A cross-platform tool to perform custom LDAP queries against a Windows domain, interactively or as one-off commands, with colored output and XLSX export.
      <br>
      <a href="https://github.com/TheManticoreProject/ldapconsole/actions/workflows/release.yaml" title="Build"><img alt="Build and Release" src="https://github.com/TheManticoreProject/ldapconsole/actions/workflows/release.yaml/badge.svg"></a>
      <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/TheManticoreProject/ldapconsole">
      <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/TheManticoreProject/ldapconsole">
      <br>
</p>

## Features

- [x] Authentications:
  - [x] Authenticate with password
  - [x] Authenticate with LM:NT hashes (Pass the Hash)
  - [x] Authenticate with Kerberos (Pass the Ticket)
  - [x] LDAP and LDAPS
- [x] Interactive mode
  - [x] Colored results
  - [x] Preset queries
  - [x] Configurable search base and search scope
  - [x] `diff` between the last two queries
- [x] Non-interactive mode
  - [x] Colored results
  - [x] Exportable to XLSX format with option `--xlsx`

## Usage

```
$ ./ldapconsole -h
ldapconsole - by Remi GASCOU (Podalirius) @ TheManticoreProject - v2.1.0

Usage: ldapconsole --domain <string> --username <string> [--password <string>] [--hashes <string>] [--debug] [--quiet] --dc-ip <string> [--ldap-port <tcp port>] [--use-ldaps] [--use-kerberos] [--query <string>] [--attribute <string>] [--xlsx <string>]

  Authentication:
    -d, --domain <string>   Active Directory domain to authenticate to.
    -u, --username <string> User to authenticate as.
    -p, --password <string> Password to authenticate with. (default: "")
    -H, --hashes <string>   NT/LM hashes, format is LMhash:NThash. (default: "")

  Configuration:
    --debug         Debug mode. (default: false)
    --quiet         Quiet mode, do not print the banner. (default: false)

  LDAP Connection Settings:
    -dc, --dc-ip <string>       IP Address of the domain controller or KDC (Key Distribution Center) for Kerberos.
    -lp, --ldap-port <tcp port> Port number to connect to LDAP server. (default: 389)
    -L, --use-ldaps             Use LDAPS instead of LDAP. (default: false)
    -k, --use-kerberos          Use Kerberos instead of NTLM. (default: false)

  Non-interactive query:
    -q, --query <string>     LDAP query to perform. If set, ldapconsole runs in non-interactive mode. (default: "")
    -a, --attribute <string> Attributes to extract. Can be specified multiple times.
    -x, --xlsx <string>      Output results of the query to an XLSX file. (default: "")
```

### Interactive mode

```bash
./ldapconsole -u 'user1' -p 'Admin123!' -d 'LAB.local' --dc-ip 192.168.2.1
```

Available console commands:

| Command       | Description                                                              |
|---------------|--------------------------------------------------------------------------|
| `query`       | Perform an LDAP query. Use `query <filter> [select <attr1> <attr2> ...]`. |
| `presetquery` | Run a builtin preset query (e.g. `all_users`, `all_kerberoastables`).    |
| `rootdse`     | Query the RootDSE of the server.                                         |
| `searchbase`  | Set the search base (a distinguishedName or a dotted FQDN).              |
| `searchscope` | Set the search scope (`BASE`, `LEVEL` or `SUBTREE`).                     |
| `infos`       | Print information about the remote LDAP server.                          |
| `diff`        | Show the differences between the last two queries.                       |
| `help`        | Display the help message.                                                |
| `exit`        | Exit ldapconsole.                                                        |

### Extract the list of the computers with an obsolete OS to an Excel file

```bash
./ldapconsole -d LAB.local -u Administrator -p 'Admin123!' --dc-ip 10.0.0.101 \
  -q '(&(objectCategory=Computer)(|(operatingSystem=Windows 2000*)(operatingSystem=Windows Vista*)(operatingSystem=Windows XP*)(operatingSystem=Windows 7*)(operatingSystem=Windows 8*)(operatingSystem=Windows Server 200*)(operatingSystem=Windows Server 2012*)))' \
  -a 'operatingSystem' -a 'operatingSystemVersion' -x ComputersWithObsoleteOSes.xlsx
```

## Build

```bash
go build -ldflags="-s -w" -o ldapconsole
```

## Contributing

Pull requests are welcome. Feel free to open an issue if you want to add other features.

## Credits

  - [Podalirius](https://github.com/p0dalirius) for the creation of the original [ldapconsole](https://github.com/p0dalirius/ldapconsole) project in Python.
