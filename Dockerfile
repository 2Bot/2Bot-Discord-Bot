FROM golang:1.10
LABEL maintainer="Noah Santschi-Cooney (noah@santschi-cooney.ch)"

WORKDIR /go/src/github.com/Strum355/2Bot-Discord-Bot
COPY . .

EXPOSE 80

RUN go get -d -v ./...
RUN go install -v ./...

CMD 2Bot-Discord-Bot