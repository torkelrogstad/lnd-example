# LND gRPC example

Example code for connecting to LND over gRPC. 

Doing gRPC connections can be a very pleasant experience, as long as 
you slightly modify the default options. This repository is intended
to be a small example of what I regard to be best practices when working
with gRPC clients in Go. 

This includes: 

  * a simple Makefile
  * type-safe CLI configuration, with `go-flags` 
  * reasonable dial options for the gRPC connection. If these aren't applied,
  you get nonsensical errors that are hard to debug, and connections that 
  hang forever. 


## Running this

Note that prior to running this, you need to have your TLS certificate
and macaroon locally. By default we look for files named `tls.cert` and
`admin.macaroon` in the current directory. 

```bash
$ make lnd-example && ./lnd-example --server=your.lnd.server:10009
```