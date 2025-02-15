.PHONY= build up down re pmm-status mongo-reg mongo-insert export-all import-all clean

PMMT_BIN_NAME?=pmm-transferer
PMM_DUMP_PATTERN?=pmm-dump-*.tar.gz

PMM_URL?="http://admin:admin@localhost:8282"
PMM_VM_URL?="http://admin:admin@localhost:8282/prometheus"
PMM_CH_URL?="http://localhost:9000?database=pmm"

PMM_MONGO_USERNAME?=pmm_mongodb
PMM_MONGO_PASSWORD?=password
PMM_MONGO_URL?=mongodb:27017

ADMIN_MONGO_USERNAME?=admin
ADMIN_MONGO_PASSWORD?=admin

DUMP_FILENAME=dump.tar.gz

BRANCH:=$(shell git branch --show-current)
COMMIT:=$(shell git rev-parse --short HEAD)

all: build re mongo-reg mongo-insert export-all re import-all

build:
	go build -ldflags "-X 'main.GitBranch=$(BRANCH)' -X 'main.GitCommit=$(COMMIT)'" -o $(PMMT_BIN_NAME) pmm-transferer/cmd/transferer

up:
	mkdir -p setup/pmm && touch setup/pmm/agent.yaml && chmod 0666 setup/pmm/agent.yaml
	docker-compose up -d
	sleep 5 # waiting for pmm server to be ready :(
	docker exec pmm-client pmm-agent setup

down:
	docker-compose down --volumes
	rm -rf setup/pmm

re: down up

pmm-status:
	docker exec pmm-client pmm-admin status

mongo-reg:
	docker exec pmm-client pmm-admin add mongodb \
		--username=$(PMM_MONGO_USERNAME) --password=$(PMM_MONGO_PASSWORD) mongo $(PMM_MONGO_URL)

mongo-insert:
	docker exec mongodb mongo -u $(ADMIN_MONGO_USERNAME) -p $(ADMIN_MONGO_PASSWORD) \
		--eval 'db.getSiblingDB("mydb").mycollection.insert( [{ "a": 1 }, { "b": 2 }] )' admin

export-all:
	./$(PMMT_BIN_NAME) export -v --dump-path $(DUMP_FILENAME) \
		--pmm-url=$(PMM_URL) --dump-core --dump-qan

export-vm:
	./$(PMMT_BIN_NAME) export -v --dump-path $(DUMP_FILENAME) \
		--pmm-url=$(PMM_URL) --dump-core

export-ch:
	./$(PMMT_BIN_NAME) export -v --dump-path $(DUMP_FILENAME) \
		--pmm-url=$(PMM_URL) --dump-qan

import-all:
	./$(PMMT_BIN_NAME) import -v --dump-path $(DUMP_FILENAME) \
		--pmm-url=$(PMM_URL) --dump-core --dump-qan

clean:
	rm -f $(PMMT_BIN_NAME) $(PMM_DUMP_PATTERN) $(DUMP_FILENAME)
