# SecretHub Proxy

> [SecretHub](https://secrethub.io) is a developer tool to help you keep database passwords, API tokens, and other secrets out of IT automation scripts.

The SecretHub Proxy provides a RESTful interface to the SecretHub Client

## Installation

### Binary

Download and extract the [latest release](https://github.com/keylockerbv/secrethub-proxy/releases/latest) of the SecretHub Proxy. Start it using:

```
secrethub-proxy -C $(cat ~/.secrethub/credential) -p 8080
```

### Docker

```
docker run -p 8080:8080 --name secrethub -v /$HOME/.secrethub:/secrethub secrethubio/proxy
```

## Development

Get the source code:

```
mkdir -p $GOPATH/src/github.com/keylockerbv
cd $GOPATH/src/github.com/keylockerbv
git clone https://github.com/keylockerbv/secrethub-proxy
```

Build it using:

```
make build
```
