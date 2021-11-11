FROM golang:1.16-alpine3.12 as build

RUN apk add --update make git bash file

WORKDIR /go/src/memcached-nats-proxy
COPY . ./

ARG VERSION
ARG GIT_TAG
ARG GIT_BRANCH
ARG GIT_COMMIT

RUN make build

FROM alpine:latest AS memcached-nats-proxy

WORKDIR /usr/bin
COPY --from=build /go/src/memcached-nats-proxy/.build .

CMD ["/usr/bin/memcached-nats-proxy"]
