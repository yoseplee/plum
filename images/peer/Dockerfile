ARG GO_VER=1.14.4
ARG ALPINE_VER=3.12

FROM alpine:${ALPINE_VER} as peer-base
RUN apk add --no-cache tzdata
# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	bash \
	gcc \
	git \
	make \
	musl-dev
ADD . $GOPATH/src/github.com/yoseplee/plum
WORKDIR $GOPATH/src/github.com/yoseplee/plum/core

FROM golang as peer
RUN go build .
RUN mv core ../peer

FROM peer-base
ENV PLUM_ROOT=/usr/local/plum
COPY --from=peer /go/src/github.com/yoseplee/plum /usr/local/plum
COPY --from=peer /go/src/github.com/yoseplee/plum/peer /usr/local/bin
EXPOSE 50051
ENV PEER_ID 0
ENV PEER_AMOUNT -1
ENV CONSENSUS PBFT
ENTRYPOINT peer -id=${PEER_ID} -lport=:50051 -amount=${PEER_AMOUNT} -local=true -consensus=${CONSENSUS}