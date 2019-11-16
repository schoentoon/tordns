#!/bin/sh

chown -R debian-tor /var/lib/tor/
chmod -R 700 /var/lib/tor/

sudo -u debian-tor tor -f /etc/tor/torrc