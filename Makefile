include .env # Requires SLACK_CLIENT_ID and SLACK_CLIENT_SECRET

CMD := blaster

HOST ?= 127.0.0.1
PORT ?= 5001

$(CMD): clean
	go build -o ./bin/$(CMD) ./cmd/$(CMD)

clean:
	rm -f ./bin/$(CMD)

test: $(CMD)
	go test -v ./...

run: $(CMD)
	GIN_MODE=debug \
	HOST=$(HOST) \
	PORT=$(PORT) \
	./bin/$(CMD)
