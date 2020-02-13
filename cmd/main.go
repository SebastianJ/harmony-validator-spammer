package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/SebastianJ/harmony-validator-spammer/config"
	"github.com/SebastianJ/harmony-validator-spammer/staking"
	"github.com/urfave/cli"
)

func main() {
	// Force usage of Go's own DNS implementation
	os.Setenv("GODEBUG", "netdns=go")

	app := cli.NewApp()
	app.Name = "Harmony Validator Spammer - stress tests a staking enabled Harmony network/blockchain"
	app.Version = fmt.Sprintf("%s/%s-%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	app.Usage = "Use --help to see all available arguments"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "network",
			Usage: "Which network to use (valid options: localnet, devnet, testnet, mainnet)",
			Value: "",
		},

		cli.StringFlag{
			Name:  "from",
			Usage: "Which address to send tokens from (must exist in the keystore)",
			Value: "",
		},

		cli.StringFlag{
			Name:  "passphrase",
			Usage: "Passphrase to use for unlocking the keystore",
			Value: "",
		},

		cli.BoolFlag{
			Name:  "infinite",
			Usage: "If the program should run in an infinite loop",
		},

		cli.IntFlag{
			Name:  "count",
			Usage: "How many transactions to send in total",
			Value: 1000,
		},

		cli.IntFlag{
			Name:  "pool-size",
			Usage: "How many validators to create simultaneously",
			Value: 100,
		},

		cli.IntFlag{
			Name:  "confirmation-wait-time",
			Usage: "How long to wait for transactions to get confirmed",
			Value: 0,
		},

		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable more verbose output",
		},
	}

	app.Authors = []cli.Author{
		{
			Name:  "Sebastian Johnsson",
			Email: "",
		},
	}

	app.Action = func(context *cli.Context) error {
		return run(context)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run(context *cli.Context) error {
	basePath, err := filepath.Abs(context.GlobalString("path"))
	if err != nil {
		return err
	}

	if err := config.Configure(basePath, context); err != nil {
		return err
	}

	staking.CreateValidators()

	return nil
}
