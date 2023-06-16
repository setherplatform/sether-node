# Sether blockchain node

Sether is the next-generation Identity & Access Management blockchain platform.

It's a PoS EVM-compatible chain secured by a leaderless DAG consensus algorithm that  
provides native services for decentralized authentication, decentralized identities and verifiable credentials.

## Building the source

Building the `sether` requires both a Go (version 1.17 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run

```shell
make
```
The build output is ```build/sether``` executable.

### Launching a local network

Launching a network with a single validator:

```shell
$ sether --fakenet 1/1
```
