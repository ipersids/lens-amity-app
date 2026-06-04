set dotenv-load := false

alias docker := start

[private]
default:
    just --list

# build docker images
[group('docker')]
build:
    set -a; source ./server/.env; set +a; \
    docker compose -f docker-compose.dev.yml build

# starts containers; [optional] use flags to specify behavior, e.g.: -d, to run it in background
[group('docker')]
up +flags='':
    docker compose -f docker-compose.dev.yml up {{ flags }}

# runs database migration
[group('docker')]
migrate:
    docker compose -f docker-compose.dev.yml exec api go run ./cmd/migrator/main.go

# stops containers
[group('docker')]
down:
    docker compose -f docker-compose.dev.yml down

# stops containers and removes named volumes and orphan containers
[group('docker')]
remove:
    docker compose -f docker-compose.dev.yml down -v --remove-orphans

# displays log output from services; [optional] use flags to specify behavior, e.g. db --tail=10
[group('docker')]
logs +flags='':
    docker compose -f docker-compose.dev.yml logs {{ flags }}

# runs app in development mode
[group('docker')]
start: build (up '-d') migrate

[group('docker')]
[private]
exec-db user='postgres' db='test':
    docker exec -it db psql -U {{ user }} -d {{ db }}
