FROM golang:1.10
LABEL maintainer="Noah Santschi-Cooney (noah@santschi-cooney.ch)"

WORKDIR /go/2Bot
COPY . .

EXPOSE 80

RUN mkdir -p /go/2Bot/images
RUN mkdir -p /go/2Bot/json

ENV GOBIN=/go/2Bot
ENV GOPATH=/go
ENV PATH=/go/2Bot:${PATH}

RUN go get -d -v ./...
RUN go install -v ./...

VOLUME ["/go/2Bot/images", "/go/2Bot/json"]

CMD 2Bot