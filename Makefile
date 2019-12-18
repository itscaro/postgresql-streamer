# TODO add .PHONY
PROJECT=postgresql-streamer-go
LOG_LEVEL?=2
EXEC_OPTS=
ifeq (,$(shell which docker))
EXEC=sh -c
else
cmd-%: override EXEC_OPTS+=-e PGS_PROJECT_NAME
cmd-%: override EXEC_OPTS+=-e PGS_PG_DSN
cmd-%: override EXEC_OPTS+=-e PGS_RABBITMQ_URI
cmd-%: override EXEC_OPTS+=-e PGS_WAL2JSON_ADD_TABLES
cmd-%: override EXEC_OPTS+=-e PGS_WAL2JSON_INCLUDE_TYPES
override EXEC_OPTS+=$(shell [ -t 0 ] || echo ' -T ')
EXEC=docker-compose exec $(EXEC_OPTS) app sh -c
endif
cmd-%: export PGS_PROJECT_NAME=test
cmd-%: export PGS_PG_DSN=postgres://postgres@postgres/postgres?sslmode=disable
cmd-%: export PGS_RABBITMQ_URI=amqp://user:password@rabbitmq:5672/
cmd-%: export PGS_WAL2JSON_ADD_TABLES=public.*
cmd-%: export PGS_WAL2JSON_INCLUDE_TYPES=0

docker-compose.override.yml: docker-compose.dev.yml
	cp $< $@

up up-app: docker-compose.override.yml

up:
	docker-compose up --build -d postgres rabbitmq

up-app:
	docker-compose up --build -d app

down:
	docker-compose down

sh:
	docker-compose exec app sh

logs:
	docker-compose logs -f postgres rabbitmq

restart:
	docker-compose restart postgres rabbitmq

.PHONY: build
build:
	docker build -t $(PROJECT)-local -f Dockerfile .

.PHONY: extract-from-image
extract-from-image:
	mkdir bin
	docker container create --name $(PROJECT)-prebuilt $(PROJECT)-local
	docker container cp $(PROJECT)-prebuilt:/app/. bin/
	-docker container rm -f $(PROJECT)-prebuilt

cmd-start:
	$(EXEC) './postgresql-streamer-amd64 -l $(LOG_LEVEL) wal start --file test'

cmd-stop:
	$(EXEC) './postgresql-streamer-amd64 -l $(LOG_LEVEL) wal stop'

cmd-init-db:
	$(EXEC) './postgresql-streamer-amd64 -l $(LOG_LEVEL) dev init-db'

cmd-reinit-db:
	$(EXEC) './postgresql-streamer-amd64 -l $(LOG_LEVEL) dev reinit-db'

cmd-gendata: GENDATA_NB?=3
cmd-gendata:
	$(EXEC) './postgresql-streamer-amd64 -l $(LOG_LEVEL) dev gendata $(GENDATA_NB)'

cmd-gendata-trx: GENDATA_NB?=3
cmd-gendata-trx:
	$(EXEC) './postgresql-streamer-amd64 -l $(LOG_LEVEL) dev gendata-trx $(GENDATA_NB)'
