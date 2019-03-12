FROM alpine

COPY secrethub-proxy /usr/bin/secrethub-proxy
RUN apk add --no-cache ca-certificates && update-ca-certificates

EXPOSE 8080

CMD secrethub-proxy -C ${SECRETHUB_CREDENTIAL:-$(cat /secrethub/credential)} -h 0.0.0.0
