# Budget Bot

Telegram bot for tracking expenses and income using Google Spreadsheets.

## Requirements

- PHP 8.2+
- Composer
- Docker & Docker Compose
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone git@github.com:positron48/budget-bot.git
cd budget-bot
```

2. Copy environment file and configure it:
```bash
cp .env.example .env.local
```

Edit `.env.local` and set the following variables:
- `APP_SECRET` - random 32 character string
- `TELEGRAM_BOT_TOKEN` - get from [@BotFather](https://t.me/BotFather)
- `TELEGRAM_BOT_USERNAME` - your bot's username (without @)
- `GOOGLE_SHEETS_CREDENTIALS_PATH` - path to Google Sheets API credentials file

3. Install dependencies:
```bash
make build
```

4. Set up the database:
```bash
docker-compose exec php bin/console doctrine:migrations:migrate
```

## Development

Start the development server:
```bash
make up
```

For local development with webhook, you can use SSH tunnel:
1. Configure tunnel settings in `.env.local`:
```env
SSH_TUNNEL_HOST=your.domain.com
SSH_TUNNEL_USER=root
SSH_TUNNEL_LOCAL_PORT=80
SSH_TUNNEL_REMOTE_PORT=8080
```

2. Start the tunnel:
```bash
make tunnel
```

3. Set up webhook URL:
```bash
docker-compose exec php bin/console app:set-webhook https://bot.your.domain.com/webhook
```

## Available Commands

- `make up` - Start the services
- `make down` - Stop the services
- `make build` - Build and start the services
- `make restart` - Restart the services
- `make tunnel` - Create SSH tunnel for webhook development
- `make permissions` - Fix var directory permissions
- `make cs-check` - Check code style
- `make cs-fix` - Fix code style
- `make phpstan` - Run static analysis
- `make test` - Run tests
- `make ci` - Run all checks (cs, phpstan, tests)

## Production Deployment

1. Configure your web server (Nginx example):
```nginx
server {
    listen 443 ssl;
    server_name bot.your.domain.com;

    ssl_certificate /etc/letsencrypt/live/bot.your.domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/bot.your.domain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name bot.your.domain.com;
    return 301 https://$server_name$request_uri;
}
```

2. Set up SSL certificate:
```bash
certbot --nginx -d bot.your.domain.com
```

3. Set webhook URL:
```bash
bin/console app:set-webhook https://bot.your.domain.com/webhook
```