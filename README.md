# Textile HTTP API + Go Example

> Example of how to use the Textile HTTP API in Go, including signing files before writing them to a vault

## Usage

First, copy the `.env.example` file to `.env`, fill in your private key string as `PRIVATE_KEY`, and define a vault identifier as `VAULT_ID`. Then, install the dependencies with `go mod download` and run the example:

```sh
go run main.go
```

This should log something like the following:

```
Signature: 6290011c02ae1349d5ded0bc5e1217da9d1efc9b8295965751e073bed2674b4a430140741fb9a0fa222c81ade54a22946fd8204ab5ddd60d0efa805528aff3b800
Creating vault 'test_signer_impl.data' for account: 0x78C61e68f9f985C43e36dD5ced3f5a24aD0c503e
Create response: {"created":true}
Writing to vault 'test_signer_impl.data'
Write response: []
Getting vault 'test_signer_impl.data' events
Events:
  CID: bafkreifsdtxsbxdws22plke4origeoervejrke2ocalhbcec2omodwj2ju
  Timestamp: 1709021225
  IsArchived: false
  CacheExpiry: 2024-03-05T20:07:07.932928
Downloading event 'bafkreifsdtxsbxdws22plke4origeoervejrke2ocalhbcec2omodwj2ju'
Event downloaded successfully
```

### Signing pkg

This example uses the `signing` package to sign the data before writing it to the vault. It's currently on the [`dtb/signer-lib` branch](https://github.com/tablelandnetwork/basin-cli/blob/dtb/signer-lib/pkg/signing/signing.go) of the `basin-cli` repository. To use it, you can add the following to your `go.mod` file:

```sh
github.com/tablelandnetwork/basin-cli v0.0.11-0.20240227064434-041f68f8efa8
```

It offers a few methods:

- `LoadPrivateKey`: Loads a private key from the given string and creates an ECDSA private key.
- `NewSigner`: Creates a new signer with the given private key, provided by `LoadPrivateKey`.
- `SignFile`: Signs the given file with the signer and returns the signature as a string, which can be used in the URL POST request to write data to a vault.
