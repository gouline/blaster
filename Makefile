include .env # Requires SLACK_CLIENT_ID and SLACK_CLIENT_SECRET

export HOST ?= localhost
export PORT ?= 4000
export CERT_FILE ?= certs/localhost.crt
export KEY_FILE ?= certs/localhost.key

.PHONY: run
run:
	DEBUG=1 air

.PHONY: test
test:
	go test -v ./...
