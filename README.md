# go-kusto-cli

A lightweight Go CLI for querying Azure Data Explorer (Kusto) via the REST API v2.

## Features

- **No SDK dependencies** — uses raw HTTP calls to the Kusto v2 REST API
- **Authentication** via Azure CLI (`az account get-access-token`)
- **Multiple input modes** — inline query, script file, or stdin
- **Output formats** — ASCII table, CSV, or JSON
- **Script mode** — execute multiple queries from a `.kql` file
- **Auto-inference** — detects `cluster('...')` and `database('...')` in your KQL, so `-c` and `-d` flags become optional

## Installation

```bash
go install github.com/danielsada/go-kusto-cli@latest
```

Or build from source:

```bash
git clone https://github.com/danielsada/go-kusto-cli.git
cd go-kusto-cli
make build
```

## Prerequisites

- [Azure CLI](https://aka.ms/installazurecli) installed and logged in (`az login`)
- Permissions to query the target Kusto cluster

## Usage

```
go-kusto-cli -c <cluster> -d <database> [-e <query> | -s <script>] [-f table|csv|json] [-o <file>]
```

### Flags

| Flag | Description |
|------|-------------|
| `-c` | Kusto cluster URL — optional if `cluster('...')` is in the query |
| `-d` | Database name — optional if `database('...')` is in the query |
| `-e` | Inline KQL query |
| `-s` | Path to a KQL script file |
| `-f` | Output format: `table` (default), `csv`, `json` |
| `-o` | Output file path (default: stdout) |

### Query Input Priority

1. `-e` flag (inline query)
2. `-s` flag (script file)
3. stdin (pipe a query)

### Examples

**Inline query with table output:**

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB -e "StormEvents | take 10"
```

**CSV output to a file:**

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB -e "StormEvents | count" -f csv -o results.csv
```

**JSON output:**

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB -e "StormEvents | take 5" -f json
```

**Pipe query via stdin:**

```bash
echo "StormEvents | count" | go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB
```

**Script file with multiple queries:**

```bash
# queries.kql
StormEvents | count
;
StormEvents | take 10
```

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB -s queries.kql
```

**Infer cluster and database from KQL (no flags needed):**

```bash
go-kusto-cli -e "cluster('mycluster').database('MyDB').StormEvents | take 10"
```

```bash
# Also works with full URLs and script files
echo "cluster('https://mycluster.westus.kusto.windows.net').database('Logs').Events | count" | go-kusto-cli
```

## Development

```bash
make lint    # Run golangci-lint
make vet     # Run go vet
make test    # Run tests
make build   # Build binary
make all     # All of the above
```

## License

MIT
