FROM golang:1.13-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

ARG PACKAGE

WORKDIR "/go/src/${PACKAGE}"

COPY ./ .

RUN go get ./...
