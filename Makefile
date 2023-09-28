.PHONY: *
up:
	docker-compose up -d --force-recreate --remove-orphans
down:
	docker-compose down -v
stop:
	docker-compose stop
