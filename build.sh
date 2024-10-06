#!/bin/sh

GOOS=linux GOARCH=amd64 go build -mod vendor -v -o chatroom -ldflags "-s -w" ./
docker build -t ohko/chatroom .
docker push ohko/chatroom
rm -rf chatroom