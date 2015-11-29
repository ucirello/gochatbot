vendor:
	go get github.com/Masterminds/glide
	GO15VENDOREXPERIMENT=1 $(GOPATH)/bin/glide install

all: vendor
	GO15VENDOREXPERIMENT=1 go build -tags "$(GCBFLAGS)"

docker: vendor
	docker build -t ccirello/gochatbot .
