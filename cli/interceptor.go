package main

import (
	"context"
	"log"
	"os"

	"github.com/lightninglabs/lndclient"
	"github.com/urfave/cli"
)

const (
	defaultRPCPort     = "10009"
	defaultRPCHostPort = "localhost:" + defaultRPCPort
)

func main() {
	app := &cli.App{
		Name:  "intercept",
		Usage: "Intercepts Htlcs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "macaroon",
				Usage: "admin macaroon for this lnd instance",
			},
			&cli.StringFlag{
				Name:  "tlspath",
				Usage: "tls path for this lnd instance",
			},
			&cli.StringFlag{
				Name:  "host",
				Value: defaultRPCHostPort,
				Usage: "host:port of this lnd's rpc " +
					"instance, e.g. localhost:10009",
			},
			&cli.StringFlag{
				Name:  "network",
				Usage: "mainnet or signet",
			},
		},
		Action: interceptHtlc,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func interceptHtlc(cliCtx *cli.Context) error {
	network := lndclient.NetworkSignet
	switch cliCtx.String("network") {
	case "mainnet":
		network = lndclient.NetworkMainnet

	case "regtest":
		network = lndclient.NetworkRegtest
	}

	lndServices, err := lndclient.NewLndServices(
		&lndclient.LndServicesConfig{
			LndAddress:  cliCtx.String("host"),
			Network:     network,
			MacaroonDir: cliCtx.String("macaroon"),
			TLSPath:     cliCtx.String("tlspath"),
		},
	)
	if err != nil {
		return err
	}
	defer lndServices.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lndServices.Router.InterceptHtlcs(ctx, func(_ context.Context,
		htlc lndclient.InterceptedHtlc) (
		*lndclient.InterceptedHtlcResponse, error) {

		log.Printf("Intercepted htlc")
		log.Printf("OutgoingChanId: %v, AmountIn: %v, AmountOut: %v",
			htlc.OutgoingChannelID, htlc.AmountInMsat,
			htlc.AmountOutMsat)

		<-ctx.Done()

		return &lndclient.InterceptedHtlcResponse{
			Action: lndclient.InterceptorActionResume,
		}, nil
	})

	return nil
}
