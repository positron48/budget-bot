.PHONY: up down build ci cs-check cs-fix phpstan test restart permissions

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

# CI commands
ci: cs-check phpstan test

cs-check:
	docker-compose exec php composer cs-check

cs-fix:
	docker-compose exec php composer cs-fix

phpstan:
	docker-compose exec php composer phpstan

test:
	docker-compose exec php composer test