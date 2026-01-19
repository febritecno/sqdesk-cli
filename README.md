# SQDesk

A lightweight yet intelligent terminal-based database client focused on fast query execution, AI-assisted typing, and easy data navigation.

## Features

- 🎨 **Beautiful TUI**: Modern terminal interface with Dracula, Nord, and Default themes
- 🗄️ **Multi-Database Support**: PostgreSQL, MySQL, and SQLite
- 🤖 **AI Integration**: Generate SQL from natural language using Gemini, Claude, or OpenAI
- ✨ **Smart Auto-completion**: Context-aware suggestions based on your schema
- 📊 **Data Browser**: Preview tables with a single keystroke
- ⚡ **Fast Query Execution**: Run queries with Ctrl+Enter

## Installation

```bash
# Clone the repository
git clone https://github.com/febritecno/sqdesk.git
cd sqdesk

# Install dependencies
go mod tidy

# Build
go build -o sqdesk ./cmd/sqdesk

# Run
./sqdesk
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Enter` | Execute query |
| `Ctrl+G` | AI: Generate SQL from natural language |
| `Ctrl+K` | AI: Refactor selected SQL |
| `Tab` | Accept suggestion / Switch pane |
| `F2` | Open settings |
| `Ctrl+Q` | Quit |
| `↑↓` | Navigate table list |
| `Enter` | Preview selected table |
| `PgUp/PgDown` | Navigate results |

## Configuration

Configuration is stored in `~/.config/sqdesk/config.yaml`:

```yaml
theme: dracula
ai:
  provider: gemini
  api_key: your-api-key
  model: gemini-1.5-flash
connections:
  - name: Production
    driver: postgres
    host: localhost
    port: 5432
    user: postgres
    password: secret
    database: mydb
active_connection: 0
```

## First Run

On first run, SQDesk will guide you through:

1. **Theme Selection**: Choose your preferred color scheme
2. **AI Configuration**: Set up AI provider for intelligent SQL generation
3. **Database Connection**: Configure your first database connection

## Layout

```
+-------------------------------------------------------------+
| SQDesk CLI | [DB: Production_v1] | Model: Gemini 1.5 Pro    |
+----------+--------------------------------------------------+
| TABLES   | 1  SELECT *                                      |
| > users  | 2  FROM transactions                             |
|   orders | 3  WHERE status = 'pending'                      |
|   items  | 4  _                                             |
|          +--------------------------------------------------+
|          | AI Prompt: [ Tambahkan filter tanggal hari ini ] |
+----------+--------------------------------------------------+
| RESULTS (3 rows found)                                      |
+----+------------+----------+-----------+--------------------+
| ID | USER_ID    | AMOUNT   | STATUS    | DATE               |
+----+------------+----------+-----------+--------------------+
| 1  | 99         | 50000    | pending   | 2024-01-18         |
+----+------------+----------+-----------+--------------------+
| [Ctrl+Enter] Run [Ctrl+G] AI [F2] Settings [Ctrl+Q] Quit    |
+-------------------------------------------------------------+
```

## Development

```bash
# Run in development
go run ./cmd/sqdesk

# Run tests
go test ./...

# Build for all platforms
make build-all
```

## License

MIT
