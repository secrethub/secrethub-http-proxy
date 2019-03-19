# SecretHub Proxy

> [SecretHub](https://secrethub.io) is a developer tool to help you keep database passwords, API tokens, and other secrets out of IT automation scripts.

The SecretHub Proxy adds a RESTful HTTP interface to the [SecretHub Client](https://). 
Apps can this way still use SecretHub, without having to directly include the client as a binary dependency.

You can be configure it with a SecretHub credential at start, thereby removing the need of passing it in on every request. 

### A note on security

The SecretHub Proxy opens up the configured SecretHub account over HTTP. 
This moves the responsibility of securing your secrets to the domain of network security, which comes with its own risks. 
So use this with caution and make sure the account has access to only those secrets it absolutely needs. 
 
It is recommended to [create a service account](https://secrethub.io/docs/reference/service-command/), tightly control it with [access rules](https://secrethub.io/docs/reference/acl-command/), and use the service credential instead of your own SecretHub account.

```
secrethub service init --permission read --desc my-app
```

## Installation

### Binary

Download and extract the [latest release](https://github.com/keylockerbv/secrethub-proxy/releases/latest) of the SecretHub Proxy. Start it with your SecretHub credential:

```
./secrethub-proxy -C $(cat ~/.secrethub/credential) -p 8080
```

If upon signup you've chosen to lock your credential with a passphrase:

```
./secrethub-proxy -C $(cat ~/.secrethub/credential) -P mycredentialpassphrase -p 8080
```

### Docker

You can also run the proxy as a [Docker container](https://hub.docker.com/r/secrethubio/proxy). 
Assuming you have a SecretHub credential stored in the default `$HOME/.secrethub` location, you can run it with the credential mounted as volume:

```
docker run -p 8080:8080 --name secrethub -v $HOME/.secrethub:/secrethub secrethubio/proxy
```

You can also pass in the credential as an environment variable:

```
docker run -p 8080:8080 --name secrethub -e SECRETHUB_CREDENTIAL=$(cat $HOME/.secrethub/credential) secrethubio/proxy
```

If upon signup you've chosen to lock your credential with a passphrase:

```
docker run -p 8080:8080 --name secrethub -e SECRETHUB_CREDENTIAL=$(cat $HOME/.secrethub/credential) -e SECRETHUB_CREDENTIAL_PASSPHRASE=mycredentialpassphrase secrethubio/proxy
```

## Usage

With the proxy up and running, you can perform the following HTTP requests:

### `/v1beta/secrets/raw/:path`

Example:

```
/v1beta/secrets/raw/my-org/my-repo/my-secret
```

#### `GET`

Returns the secret contents as bytes.

#### `POST`

Creates or updates a secret. Expects the secret contents as bytes.

#### `DELETE`

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
