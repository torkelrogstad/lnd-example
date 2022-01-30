package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/macaroon.v2"
)

// Type-safe declaration of CLI params. Note how you can specify both
// default values, and mark params as required.
// Docs on all the options: https://pkg.go.dev/github.com/jessevdk/go-flags
type config struct {
	TLSPath       string `long:"tls-cert" default:"./tls.cert" required:"true" description:"path to TLS certificate"`
	MacaroonPath  string `long:"macaroon" default:"./admin.macaroon" required:"true" description:"path to macaroon"`
	RpcServer     string `long:"server" required:"true" description:"remote server location"`
	EnableGrpcLog bool   `long:"grpclog" description:"enable gRPC logging"`
}

func main() {
	log.Println("lnd-example: starting")

	// Apply a timeout to our context, to ensure we don't hang forever.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var conf config
	if _, err := flags.Parse(&conf); err != nil {
		log.Fatal(err)
	}

	if conf.EnableGrpcLog {
		enableGrpcLogger()
	}

	cert, err := credentials.NewClientTLSFromFile(conf.TLSPath, "")
	if err != nil {
		log.Fatalf("could not read TLS cert: %s", err)
	}

	macBytes, err := os.ReadFile(conf.MacaroonPath)
	if err != nil {
		log.Fatalf("could not read macaroon: %s", err)
	}

	var mac macaroon.Macaroon
	if err := mac.UnmarshalBinary(macBytes); err != nil {
		log.Fatalf("could not read macaroon bytes into struct: %s", err)
	}

	start := time.Now()

	// Dial with the context, so we automatically error out after it has timed
	// out. With the default grpc.Dial method, you can potentially hang forever.
	conn, err := grpc.DialContext(ctx, conf.RpcServer,
		// Gives a nice error message, otherwise we get "context canceled". Should
		// be on by default, IMO.
		grpc.WithReturnConnectionError(),

		// Don't retry non-temporary connection issues. Should also be on by
		// default, IMO.
		grpc.FailOnNonTempDialError(true),

		// Block until we have a valid connection. Nice for returning errors
		// early, instead of waiting until our first RPC call.
		grpc.WithBlock(),

		// If your node has a valid TLS certificate (i.e. not a self-signed cert),
		// you can omit this.
		grpc.WithTransportCredentials(cert),

		// Authenticate with our macaroon for every RPC
		grpc.WithPerRPCCredentials(macaroons.MacaroonCredential{Macaroon: &mac}),
	)
	if err != nil {
		log.Fatalf("could not dial to LND: %s", err)
	}

	log.Printf("dialed to LND\tduration=%s", time.Since(start))

	lnd := lnrpc.NewLightningClient(conn)
	info, err := lnd.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		log.Fatalf("could not get info: %s", err)
	}

	log.Printf("LND info: %s", info.String())
}

// Enable gRPC logging. The underlying gRPC library handles this in a stupid
// way: there's no way to do logging for a specific gRPC connection. You
// have to enable it globally. This method needs to be called before any
// gRPC methods.
func enableGrpcLogger() {
	var (
		info = io.Discard // very noisy, we don't care about this
		warn = os.Stderr
		err  = os.Stderr
	)
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(info, warn, err))
}
