include .env # Requires SLACK_CLIENT_ID and SLACK_CLIENT_SECRET

CMD=blaster

$(CMD): clean
	go build -o ./bin/$(CMD) ./cmd/$(CMD)

clean:
	rm -f ./$(CMD)

test: $(CMD)
	go test -v ./...

run: $(CMD)
	GIN_MODE=debug PORT=5000 ./bin/$(CMD)

heroku: $(CMD)
	heroku local
