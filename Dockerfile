FROM alpine

COPY secrethub-http-proxy /usr/bin/secrethub-http-proxy
RUN apk add --no-cache ca-certificates && update-ca-certificates

EXPOSE 8080

CMD secrethub-http-proxy -C ${SECRETHUB_CREDENTIAL:-$(cat /secrethub/credential)} -P ${SECRETHUB_CREDENTIAL_PASSPHRASE} -h 0.0.0.0 -p 8080
