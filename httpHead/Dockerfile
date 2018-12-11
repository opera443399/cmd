FROM alpine:3.8
RUN apk add --no-cache ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/* /tmp/* /var/tmp/*
COPY httpHead /
ENTRYPOINT ["/httpHead"]
