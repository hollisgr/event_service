FROM golang:1.24-alpine 

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY ["../go.mod", "../go.sum", "./"]

RUN go mod download

# build
COPY ./ ./

RUN go build -o scheduler ./cmd/scheduler/scheduler.go
CMD ["./scheduler"]