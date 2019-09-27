FROM golang:latest AS builder

WORKDIR /root

RUN git clone https://github.com/coredns/coredns

WORKDIR /root/coredns

RUN go mod download

COPY . /root/tordns

RUN echo "replace github.com/schoentoon/tordns => /root/tordns" >> go.mod

RUN echo 'tordns:github.com/schoentoon/tordns' >> plugin.cfg

RUN make

FROM debian:latest

RUN apt-get update && apt-get install -y dnsutils ca-certificates

COPY --from=builder /root/coredns/coredns /bin/coredns

ENTRYPOINT [ "/bin/coredns" ]