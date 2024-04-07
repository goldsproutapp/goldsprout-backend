FROM golang:1.22 as build

WORKDIR /app

ENV CGO_ENABLED 0
ENV GOOS linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /investment-tracker

# TODO: perhaps alpine would be better?
FROM ubuntu:20.04

COPY --from=build /investment-tracker /investment-tracker
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh

RUN apt update
# entrypoint script uses nc to wait for db server
RUN apt install -y netcat-openbsd

EXPOSE 3000

ENTRYPOINT [ "/docker-entrypoint.sh" ]
