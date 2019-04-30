<p align="center">
  <img src="https://secrethub.io/img/secrethub-logo.svg" alt="SecretHub" width="380px"/>
</p>
<h1 align="center">
  <i>HTTP Proxy<sup><a href="#beta">BETA</a></sup></i>
</h1>

The SecretHub HTTP Proxy adds a RESTful interface to the [SecretHub Client](https://github.com/secrethub/secrethub-go). 
Apps can this way still use SecretHub, without having to directly include the client as a binary dependency.

You can configure it with a SecretHub credential at start, thereby removing the need of passing it in on every request. 

> [SecretHub](https://secrethub.io) is a developer tool to help you keep database passwords, API tokens, and other secrets out of IT automation scripts.

### A note on security

The SecretHub HTTP Proxy opens up the configured SecretHub account over HTTP. 
This moves the responsibility of securing your secrets to the domain of network security, which comes with its own risks. 
So use this with caution and make sure the credential you pass in only has access to only those secrets it needs. 

It is recommended to [create a service account](https://secrethub.io/docs/reference/service-command/), tightly control it with [access rules](https://secrethub.io/docs/reference/acl-command/), and use the service credential instead of your own SecretHub account.

```
secrethub service init my-org/my-repo --permission read --desc my-app
```

## Installation

### Binary

Download and extract the [latest release](https://github.com/secrethub/secrethub-http-proxy/releases/latest) of the SecretHub HTTP Proxy. Start it with your SecretHub credential:

```
./secrethub-http-proxy -C $(cat ~/.secrethub/credential) -p 8080
```

If upon signup you've chosen to lock your credential with a passphrase, you will get prompted for your passphrase.

### Docker

You can also run the proxy as a [Docker container](https://hub.docker.com/r/secrethub/http-proxy). 
Assuming you have a SecretHub credential stored in the default `$HOME/.secrethub` location, you can run it with the credential mounted as a volume:

```
docker run -p 127.0.0.1:8080:8080 --name secrethub -v $HOME/.secrethub:/secrethub secrethub/http-proxy
```

You can also pass in the credential as an environment variable:

```
docker run -p 127.0.0.1:8080:8080 --name secrethub -e SECRETHUB_CREDENTIAL=$(cat $HOME/.secrethub/credential) secrethub/http-proxy
```

If upon signup you've chosen to lock your credential with a passphrase, run the container with `-it` to get prompted for your passphrase.

```
docker run -it -p 127.0.0.1:8080:8080 --name secrethub -e SECRETHUB_CREDENTIAL=$(cat $HOME/.secrethub/credential) secrethub/http-proxy
```

Alternatively, the passphrase can be sourced from the `SECRETHUB_CREDENTIAL_PASSPHRASE` environment variable.

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

## BETA

This project is currently in beta and we'd love your feedback! Check out the [issues](https://github.com/secrethub/secrethub-http-proxy/issues) and feel free to suggest cool ideas, use cases, or improvements.

Because it's still in beta, you can expect to see some changes introduced. Pull requests are very welcome.

## Terraform State Backend

For those of you using [Terraform](https://www.terraform.io), the SecretHub HTTP Proxy can function as a [Terraform Backend](https://www.terraform.io/docs/backends/index.html) for your `.tfstate`. 
Read more about this on our [blog post]().

## Development

Get the source code:

```
git clone https://github.com/secrethub/secrethub-http-proxy
```

To build the binary from source, use:

```
make build
```

To build the Docker image from scratch, you can use:

```
docker build -t secrethub-http-proxy .
```