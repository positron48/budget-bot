.PHONY: up down build ci cs-check cs-fix phpstan test restart permissions tunnel

include .env.local
export

# Docker commands
up:
	docker-compose up -d

down:
	docker-compose down

build:
	docker-compose build
	docker-compose up -d

restart:
	docker-compose restart

permissions:
	mkdir -p var/cache var/log
	chmod -R 777 var

tunnel:
	ssh -R $(SSH_TUNNEL_REMOTE_PORT):localhost:$(SSH_TUNNEL_LOCAL_PORT) $(SSH_TUNNEL_USER)@$(SSH_TUNNEL_HOST) -N

# CI commands
ci: cs-check phpstan test

cs-check:
	docker-compose exec php composer cs-check -- --allow-risky=yes

cs-fix:
	docker-compose exec php composer cs-fix -- --allow-risky=yes

phpstan:
	docker-compose exec php composer phpstan

test:
	docker-compose exec php composer test