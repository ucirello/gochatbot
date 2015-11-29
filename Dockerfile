FROM golang

ADD . /go/src/cirello.io/gochatbot/

RUN GO15VENDOREXPERIMENT=1 go install cirello.io/gochatbot

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/gochatbot