FROM golang:1.14-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

ARG PACKAGE

WORKDIR "/go/src/${PACKAGE}"

COPY ./ .

RUN export GO111MODULE=on &&\
    go mod download && \
    go mod vendor
