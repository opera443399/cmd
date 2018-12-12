FROM alpine:3.8
RUN apk add --no-cache ca-certificates \
    && rm -rf /var/cache/apk/*

ENV APP_VERSION=0.1.2
COPY httpHead /
ENTRYPOINT ["/httpHead"]
