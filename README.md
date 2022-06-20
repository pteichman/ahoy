# Ahoy, team!

Ahoy is a command line client for the Spring '83 protocol:
https://github.com/robinsloan/spring-83-spec

This code is currently targeting draft-20220616.md. I can't make any
promises yet about the CLI or library interfaces breaking over time, but I
like backward compatible software as much as you do.

## Installing without this repo (just give me the command!)

Install the Go toolchain if you don't already have it: https://go.dev/dl/

And then:

```sh
$ go install github.com/pteichman/ahoy/cmd/ahoy@latest
$ ~/go/bin/ahoy --help
```

## Building from a local clone

```sh
$ go build ./cmd/ahoy
$ ./ahoy --help
```

## Server selection and authentication

Your Spring '83 identity is a text file containing the hex encoding of a
private and public key, concatenated together. You can generate one of
these files:

```sh
$ ahoy keygen
Checked 9361687 candidates in 22.79s
Pubkey: 47e0f417f42634b42917124c8c9709714ac28c632830c2f96f8e52beb83e0623
Wrote: spring-83-keypair-2022-06-18-47e0f417f426.txt
```

This keypair must remain private and cannot be regenerated if lost.
