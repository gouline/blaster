include .env # Requires SLACK_CLIENT_ID and SLACK_CLIENT_SECRET

IMAGE := blaster:latest

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

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) .

.PHONY: docker-run
docker-run: docker-build
	docker run --rm \
		-e DEBUG=1 \
		-e PORT=$(PORT) \
		-e CERT_FILE="$(CERT_FILE)" \
		-e KEY_FILE="$(KEY_FILE)" \
		-e SLACK_CLIENT_ID="$(SLACK_CLIENT_ID)" \
		-e SLACK_CLIENT_SECRET="$(SLACK_CLIENT_SECRET)" \
		-p $(PORT):$(PORT) \
		-v "$(PWD)/certs:/app/certs" \
		$(IMAGE)
