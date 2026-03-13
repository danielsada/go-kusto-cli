# go-kusto-cli

[![CI](https://github.com/danielsada/go-kusto-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/danielsada/go-kusto-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielsada/go-kusto-cli)](https://goreportcard.com/report/github.com/danielsada/go-kusto-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight, zero-dependency Go CLI for querying [Azure Data Explorer (Kusto)](https://learn.microsoft.com/en-us/azure/data-explorer/) via the REST API v2.

## Features

- **No SDK dependencies** — uses raw HTTP calls to the [Kusto v2 REST API](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/api/rest/), only stdlib
- **Azure CLI authentication** — obtains a bearer token via `az account get-access-token`
- **Multiple input modes** — inline query (`-e`), script file (`-s`), or piped stdin
- **Output formats** — ASCII table, CSV, or JSON
- **Script mode** — execute multiple semicolon-separated queries from a `.kql` file
- **Auto-inference** — detects `cluster('...')` and `database('...')` in your KQL so `-c` and `-d` flags become optional

## Installation

**Download a binary** from the [latest release](https://github.com/danielsada/go-kusto-cli/releases/latest) — pre-built for Linux, macOS, and Windows (amd64 & arm64).

Or install with Go:

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

- **Go 1.23+** (build only)
- [Azure CLI](https://aka.ms/installazurecli) installed and logged in (`az login`)
- Permissions to query the target Kusto cluster

## Quick Start

```bash
# Simple inline query — results printed as an ASCII table
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB \
  -e "StormEvents | take 10"
```

## Usage

```
go-kusto-cli -c <cluster> -d <database> [-e <query> | -s <script>] [-f table|csv|json] [-o <file>]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-c` | Kusto cluster URL (optional if `cluster('...')` is in the query) | — |
| `-d` | Database name (optional if `database('...')` is in the query) | — |
| `-e` | Inline KQL query | — |
| `-s` | Path to a `.kql` script file | — |
| `-f` | Output format: `table`, `csv`, `json` | `table` |
| `-o` | Write output to a file instead of stdout | stdout |

### Query Input Priority

When multiple inputs are available, the CLI resolves them in this order:

1. `-e` flag (inline query)
2. `-s` flag (script file)
3. stdin (piped input)

### Examples

**Export to CSV:**

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB \
  -e "StormEvents | count" -f csv -o results.csv
```

**JSON output:**

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB \
  -e "StormEvents | take 5" -f json
```

**Pipe a query via stdin:**

```bash
echo "StormEvents | count" | go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB
```

**Run a script file with multiple queries** (separated by `;` or blank lines):

```kql
// queries.kql
StormEvents | count
;
StormEvents | take 10
```

```bash
go-kusto-cli -c https://mycluster.westus.kusto.windows.net -d MyDB -s queries.kql
```

**Auto-infer cluster and database from the query — no flags needed:**

```bash
go-kusto-cli -e "cluster('mycluster').database('MyDB').StormEvents | take 10"
```

```bash
# Works with full URLs too
echo "cluster('https://mycluster.westus.kusto.windows.net').database('Logs').Events | count" \
  | go-kusto-cli
```

## How It Works

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  Input   │──▶│  Auth    │──▶│  Client  │──▶│  Parser  │──▶│ Formatter│
│ Resolver │   │ (az CLI) │   │ (REST v2)│   │ (frames) │   │(tbl/csv/ │
│          │   │          │   │          │   │          │   │   json)  │
└──────────┘   └──────────┘   └──────────┘   └──────────┘   └──────────┘
```

1. **Input** — resolves the query from `-e`, `-s`, or stdin; infers cluster/database from `cluster()`/`database()` calls
2. **Auth** — obtains a bearer token via `az account get-access-token --resource https://api.kusto.windows.net`
3. **Client** — sends `POST <cluster>/v2/rest/query` with `{db, csl}` payload
4. **Parser** — decodes the Kusto frame-based JSON response and extracts the `PrimaryResult` table
5. **Formatter** — renders the result as an ASCII table, CSV, or JSON to stdout or a file

## Development

```bash
make lint    # golangci-lint (errcheck, staticcheck, revive, …)
make vet     # go vet
make test    # go test -race
make build   # build binary
make all     # lint + vet + test + build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

## License

[MIT](LICENSE) © Daniel Sada
