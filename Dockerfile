FROM golang:latest AS builder

WORKDIR /root

RUN git clone https://github.com/coredns/coredns

WORKDIR /root/coredns

RUN go mod download

COPY . /root/tordns

RUN echo "replace github.com/schoentoon/tordns => /root/tordns" >> go.mod

RUN echo 'tordns:github.com/schoentoon/tordns' >> plugin.cfg

RUN make

FROM debian:buster

RUN apt-get update && apt-get install -y \
    dnsutils \
    ca-certificates \
    tor \
    --no-install-recommends

COPY --from=builder /root/coredns/coredns /bin/coredns

USER debian-tor

ENTRYPOINT [ "/bin/coredns" ]