FROM alpine:3.8
RUN apk add --no-cache ca-certificates \
    && rm -rf /var/cache/apk/*
COPY httpHead /
ENTRYPOINT ["/httpHead"]
