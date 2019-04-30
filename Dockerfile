FROM golang:1.12-alpine as build_base
WORKDIR /build
ENV GO111MODULE=on
RUN apk add --update git
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM build_base as build
RUN apk add --update make
COPY . .
RUN make build

FROM alpine
COPY --from=build /build/secrethub-http-proxy /usr/bin/secrethub-http-proxy
RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

EXPOSE 8080

CMD secrethub-http-proxy -C ${SECRETHUB_CREDENTIAL:-$(cat /secrethub/credential)} -P ${SECRETHUB_CREDENTIAL_PASSPHRASE:-""} -h 0.0.0.0 -p 8080
