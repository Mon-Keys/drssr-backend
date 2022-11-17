FROM golang:1.18 AS build

ADD . /app
WORKDIR /app
RUN go build ./cmd/main.go

FROM ubuntu:20.04

WORKDIR /usr/src/app

COPY . .
COPY --from=build /app/main/ .

EXPOSE 3001
CMD ./main