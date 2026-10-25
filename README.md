# AfterParty Bot

A Telegram bot designed to manage event ticketing, entry control, and sales for events and afterparties.

## 🌟 Features

- **Ticket Validation**: Scan and validate tickets at entry points
- **Entry Tracking**: Mark attendees as entered to prevent ticket reuse
- **Ticket Sales**: Sell tickets directly through the bot with digital ticket generation
- **User Management**: Role-based access control for checkers and sellers
- **Search Capability**: Find tickets by surname or ticket ID
- **VIP Ticket Support**: Handle different ticket tiers with specific permissions

## 🏗️ Architecture

Built using clean architecture principles with Go:

- **Handlers Layer**: Processes Telegram bot interactions
- **Service Layer**: Implements business logic
- **Repository Layer**: Manages database operations
- **Models**: Defines core domain entities

## 🔧 Tech Stack

- **Go** (1.23.0): Core programming language
- **Telegram Bot API**: Bot interface
- **PostgreSQL**: Database for ticket and user information
- **Zap**: Structured logging
- **Goose**: Database migrations
- **GG**: Graphics generation for tickets

## 🚀 Getting Started

### Prerequisites

- Go 1.23.0+
- PostgreSQL database
- Telegram Bot Token

### Environment Setup

Create a `.env` file with the following variables:

```
TELEGRAM_TOKEN=
USERS_COUNT=

SECRET_KEY=
DEPLOYMENT_URL=
TABLE_ID=

DB_HOST=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_PORT=

BASE_LACE=
VIP_LACE=
ORG_LACE=

VIP_TABLES_COUNT=
PRICES=x,x,x,...
DATES=YYYY-MM-DD,YYYY-MM-DD,YYYY-MM-DD,...

ALLOWED_SELLERS=...
ALLOWED_CHECKERS=...
VIP_SELLER=...
SS_SELLER=...

APP_ENV=prod/dev

TICKET_NUMBER_POS_X=...
TICKET_NUMBER_POS_Y=...
TICKET_NUMBER_FONT_SIZE=...
TICKET_NUMBER_COLOR=#ffffff
```

`TICKET_NUMBER_POS_X`, `TICKET_NUMBER_POS_Y`, and `TICKET_NUMBER_FONT_SIZE`
are required. `TICKET_NUMBER_COLOR` is optional and defaults to `#ffffff`.
X/Y are the center of the number on the ticket image.

### Installation

1. Clone the repository
```bash
git clone https://github.com/qRe0/afterparty-bot.git
cd afterparty-bot
```

2. Install dependencies
```bash
go mod download
```

3. Build the application
```bash
go build -o afterparty-bot ./cmd
```

4. Run the application
```bash
./afterparty-bot
```

## 📋 Usage

### Bot Commands

- `/start` - Initialize the bot and display available options
- `Отметить вход` - Mark an attendee as entered (Checkers only)
- `Продать билет` - Sell a ticket to a new attendee (Sellers only)

### User Roles

- **Checkers**: Can validate tickets and mark attendees as entered
- **Sellers**: Can sell tickets to new attendees
- **VIP Sellers**: Can sell both regular and VIP tickets

## 🛠️ Development

### Project Structure

```
afterparty-bot/
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── app/              # Application bootstrapping
│   ├── configs/          # Configuration loading
│   ├── errors/           # Error definitions
│   ├── handlers/         # Telegram message handlers
│   ├── migrations/       # Database migrations
│   ├── models/           # Domain models
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic
│   ├── shared/           # Shared utilities
│   └── tools/            # Developer-only internal tools
├── assets/               # Static assets
└── go.mod                # Go module definition
```

### Ticket Layout Tool

Use the local tool to position the ticket number on `assets/ticket.png` before
a selling phase.

```bash
cd internal/tools/ticket-layout
npm ci
npm run dev
```

Open the printed localhost URL. The tool loads `assets/ticket.png` and
`assets/font.ttf` automatically. Drag the number, resize it with the corner
handle or the numeric input, choose a color, and open the results dialog. Copy
the four lines into `.env`, then restart the bot.

The tool shows temporary center guides after each adjustment and a persistent
visual-center offset readout. The guides are visual only; they do not snap or
change the exported values.

The tool is developer-only and never writes `.env`. Chrome is the primary
browser; current Safari is supported. The committed font asset is normalized
for browser loading; if you replace `assets/font.ttf`, verify it loads in Chrome
before using the tool.

### Adding New Features

1. Define domain models in `internal/models/`
2. Implement repository methods in `internal/repository/`
3. Add business logic in `internal/service/`
4. Create handlers for user interaction in `internal/handlers/`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- **qRe0** - [GitHub](https://github.com/qRe0)
