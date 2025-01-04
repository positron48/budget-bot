# Budget Bot

A Telegram bot built with Symfony for tracking expenses and income using Google Spreadsheets.

## Features

- Track expenses and income through Telegram messages
- Integration with Google Spreadsheets for data storage
- Automatic category detection based on description
- Multi-user support
- Flexible date input formats
- Command to list all available spreadsheets

## Requirements

- Docker
- Docker Compose
- Google Sheets API credentials
- Telegram Bot Token

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/budget-bot.git
cd budget-bot
```

2. Copy the environment file:
```bash
cp .env.example .env
```

3. Configure your environment variables in `.env`:
- Set your Telegram Bot Token
- Configure Google Sheets API credentials
- Adjust other settings as needed

4. Build and start the Docker containers:
```bash
docker-compose up -d --build
```

5. Install dependencies:
```bash
docker-compose run composer install
```

6. Create the database:
```bash
docker-compose exec php bin/console doctrine:database:create
docker-compose exec php bin/console doctrine:migrations:migrate
```

## Usage

### Basic Commands

- `/start` - Initialize the bot
- `/list` - Get a list of all available spreadsheets

### Adding Transactions

Format: `[date] [+]amount description`

Examples:
- `1000 groceries` - Adds expense of 1000 for groceries today
- `yesterday 500 restaurant` - Adds expense from yesterday
- `12.12 +5000 salary` - Adds income for December 12
- `1500.50 taxi` - Adds expense with cents

## Development

### Project Structure

- `src/Command/` - Telegram bot commands
- `src/Service/` - Core services (Google Sheets, Category detection, etc.)
- `src/Entity/` - Database entities
- `src/Repository/` - Database repositories

### Adding New Categories

Edit the category mappings in `config/categories.yaml` to add new category aliases.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is open-source and available under the MIT License. 