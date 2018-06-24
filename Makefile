include .env # Requires SLACK_CLIENT_ID and SLACK_CLIENT_SECRET

CMD=blaster

install: clean
	go build -o ./bin/$(CMD) ./cmd/$(CMD)

clean:
	rm -f ./$(CMD)

run: install
	GIN_MODE=debug PORT=5000 ./bin/$(CMD)

heroku: install
	heroku local
