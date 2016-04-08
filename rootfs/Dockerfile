FROM alpine:3.3

RUN apk add -U ca-certificates && rm -rf /var/cache/apk/*

COPY bin/boot /bin/boot

CMD ["boot"]
