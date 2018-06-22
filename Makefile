include .env # Requires SLACK_CLIENT_ID and SLACK_CLIENT_SECRET

CMD=blaster

install: clean
	go build

run: install
	GIN_MODE=release PORT=5000 ./$(CMD)

clean:
	rm -f ./$(CMD)
