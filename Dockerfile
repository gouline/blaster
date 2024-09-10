ARG GO_VERSION=1.23

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /go/src/app

COPY internal/ ./internal/
COPY go.mod go.sum ./
COPY main.go ./

RUN go mod download

# Builds the application as a staticly linked one, to allow it to run on alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app

COPY static/ ./static/
COPY templates/ ./templates/

COPY --from=build /go/src/app .

CMD ["./app"]
