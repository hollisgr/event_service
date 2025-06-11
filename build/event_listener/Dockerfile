FROM golang:1.24-alpine 

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY ["../go.mod", "../go.sum", "./"]

RUN go mod download

# build
COPY ./ ./

RUN go build -o event_listener ./cmd/event_listener/event_listener.go
CMD ["./event_listener"]