ARG GO_VERSION=1.12.5

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --update --no-cache ca-certificates make git curl mercurial build-base 

ARG PACKAGE=github.com/darxkies/k8s-tew
ARG WORKING_DIRECTORY=/go/src/${PACKAGE}/

ENV GO111MODULE=on

RUN echo ${WORKING_DIRECOTORY}
RUN mkdir -p ${WORKING_DIRECTORY}

WORKDIR ${WORKING_DIRECTORY}

COPY go.mod go.sum ${WORKING_DIRECTORY}

RUN go mod download

RUN go get github.com/gobuffalo/packr/...

CMD ["make", "build-binaries"]
