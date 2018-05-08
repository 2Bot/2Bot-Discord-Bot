FROM golang:alpine AS builder
LABEL maintainer="Noah Santschi-Cooney (noah@santschi-cooney.ch)"

WORKDIR /go/src/github.com/Strum355/2Bot-Discord-Bot

ENV GOBIN=/go/2Bot
ENV GOPATH=/go

COPY . .

RUN apk update && \
    apk add --no-cache opus git && \
    git remote set-url origin https://github.com/Strum355/2Bot-Discord-Bot

RUN go get -d -v ./... && \ 
    go install -v ./...

FROM alpine

COPY --from=builder /go/2Bot /go/2Bot

ENV PATH=/go/2Bot:/go/2Bot/ffmpeg:${PATH}

RUN mkdir -p /go/2Bot/images/ && \
    mkdir -p /go/2Bot/json/ && \
    mkdir -p /go/2Bot/emoji/ && \
    mkdir -p /go/2Bot/ffmpeg

RUN apk --no-cache add ca-certificates && \
    update-ca-certificates

WORKDIR /go/2Bot

ADD https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-64bit-static.tar.xz .
RUN tar -xf ffmpeg-release-64bit-static.tar.xz --strip 1 -C ./ffmpeg

VOLUME ["/go/2Bot/images", "/go/2Bot/json"]

EXPOSE 8080

CMD ["2Bot-Discord-Bot"]