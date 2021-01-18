FROM alpine:latest

COPY build/* /usr/bin/

VOLUME /config
WORKDIR /config

ENTRYPOINT ["/usr/bin/DroneExternalConfig"]
