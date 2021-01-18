FROM alpine:latest

COPY build/DroneExternalConfig /usr/bin/

VOLUME /config
WORKDIR /config

ENTRYPOINT ["/usr/bin/DroneExternalConfig"]
