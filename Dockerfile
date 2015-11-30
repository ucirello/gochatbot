FROM golang:alpine

MAINTAINER Carlos Cirello <carlos.cirello.nl@gmail.com>

ADD . /go/src/cirello.io/gochatbot/

RUN GO15VENDOREXPERIMENT=1 go install -tags "all" cirello.io/gochatbot

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/gochatbot