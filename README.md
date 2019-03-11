# SecretHub Proxy

> [SecretHub](https://secrethub.io) is a developer tool to help you keep database passwords, API tokens, and other secrets out of IT automation scripts.

The SecretHub Proxy adds a RESTful interface to the SecretHub Client and can be configured with a SecretHub credential at start, thereby removing the need of passing it in on every request. 
This moves the responsibility of securing your secrets to the domain of network security.

## Installation

### Binary

Download and extract the [latest release](https://github.com/keylockerbv/secrethub-proxy/releases/latest) of the SecretHub Proxy. Start it with your SecretHub credential:

```
./secrethub-proxy -C $(cat ~/.secrethub/credential) -p 8080
```

### Docker

You can also run the proxy as a [Docker container](). Assuming you have a SecretHub credential stored in the default `$HOME/.secrethub` location, run:

```
docker run -p 8080:8080 --name secrethub -v /$HOME/.secrethub:/secrethub secrethubio/proxy
```

## Usage

With the proxy up and running, you can perform the following HTTP requests:

### `/v1/secrets/:path`

Example:

```
/v1/secrets/my-org/my-repo/my-secret
```

#### `GET`

Returns the secret contents as bytes.

### `POST`

Creates or updates a secret. Expects the secret contents as bytes.

### `DELETE`

Deletes the entire secret and its history.

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
