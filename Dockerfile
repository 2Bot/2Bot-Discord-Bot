FROM golang:1.10-alpine
LABEL maintainer="Noah Santschi-Cooney (noah@santschi-cooney.ch)"

RUN mkdir -p /go/2Bot/images/ && mkdir -p /go/2Bot/json/ && mkdir -p /go/2Bot/emoji/ && mkdir -p /go/2Bot/ffmpeg

WORKDIR /go/2Bot

ADD https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-64bit-static.tar.xz .
RUN tar -xf ffmpeg-release-64bit-static.tar.xz --strip 1 -C ./ffmpeg

WORKDIR /go/src/github.com/Strum355/2Bot-Discord-Bot
COPY . .

RUN apk update && apk add --no-cache opus git && git remote set-url origin https://github.com/Strum355/2Bot-Discord-Bot

ENV GOBIN=/go/2Bot
ENV GOPATH=/go
ENV PATH=/go/2Bot:/go/2Bot/ffmpeg:${PATH}

VOLUME ["/go/2Bot/images", "/go/2Bot/json"]

RUN go get -d -v ./... && go install -v ./...

EXPOSE 8080

WORKDIR /go/2Bot

CMD ["2Bot-Discord-Bot"]