FROM golang:alpine

MAINTAINER Carlos Cirello <carlos.cirello.nl@gmail.com>

# Uncomment the following line to import plugins into the image
# COPY gochatbot-plugin-* /
COPY gochatbot-container /

CMD ["/gochatbot-container"]