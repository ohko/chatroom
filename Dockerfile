FROM scratch
COPY chatroom /
CMD ["/chatroom"]

# GOOS=linux GOARCH=amd64 go build -mod vendor -v -o chatroom -ldflags "-s -w" ./
# docker build -t ohko/chatroom .