FROM golang

ARG CONFIG_FILE="config.json"
ENV CONFIG_FILE=$CONFIG_FILE

WORKDIR /go/src/discord-ncov
ADD . /go/src/discord-ncov

COPY $CONFIG_FILE /go/src/discord-ncov/config.json

# Build it:
RUN go get
CMD go run main.go