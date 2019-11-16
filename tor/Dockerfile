FROM debian:buster

RUN apt-get update && apt-get install -y \
    tor \
    tor-geoipdb \
    sudo \
    --no-install-recommends

EXPOSE 9050

COPY torrc /etc/tor/torrc
RUN chown -R debian-tor /etc/tor

RUN mkdir -p /var/lib/tor/.tor
VOLUME /var/lib/tor/.tor
RUN chown -R debian-tor /var/lib/tor/.tor

COPY entrypoint.sh /entrypoint.sh
RUN chmod ugo+rx /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
