package config

import (
	"errors"
	"path/filepath"
	"strings"

	sdkNetwork "github.com/SebastianJ/harmony-sdk/network"
	"github.com/SebastianJ/harmony-validator-spammer/utils"
	"github.com/harmony-one/go-sdk/pkg/sharding"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Configuration - the central configuration for the test suite tool
var Configuration Config

// Configure - configures the test suite tool using a combination of the YAML config file as well as command arguments
func Configure(basePath string, context *cli.Context) (err error) {
	configPath := filepath.Join(basePath, "config.yml")
	if err = loadYamlConfig(configPath); err != nil {
		return err
	}

	if Configuration.BasePath == "" {
		Configuration.BasePath = basePath
	}

	Configuration.Verbose = context.GlobalBool("verbose")
	// Set the verbosity level of harmony-sdk
	sdkNetwork.Verbose = Configuration.Verbose

	// It's very important that configureTransactionsConfig gets executed first since it sets config fields that are later used by other configuration steps
	if err = configureBaseConfig(context); err != nil {
		return err
	}

	if err = configureNetworkConfig(context); err != nil {
		return err
	}

	return nil
}

func configureBaseConfig(context *cli.Context) error {
	fromAddress := context.GlobalString("from")
	if fromAddress == "" {
		return errors.New("you need to specify the sender address")
	}

	Configuration.Funding.Account.Address = fromAddress

	if passphrase := context.GlobalString("passphrase"); passphrase != "" && passphrase != Configuration.Funding.Account.Passphrase {
		Configuration.Funding.Account.Passphrase = passphrase
	}

	Configuration.Funding.Account.Unlock()

	Configuration.Funding.Account.Nonce = sdkNetwork.CurrentNonce(Configuration.Network.RPC, Configuration.Funding.Account.Address)

	Configuration.Funding.Gas.Initialize()

	Configuration.Application.Infinite = context.GlobalBool("infinite")

	if count := context.GlobalInt("count"); count >= 0 && count != Configuration.Application.Count {
		Configuration.Application.Count = count
	}

	if poolSize := context.GlobalInt("pool-size"); poolSize >= 0 && poolSize != Configuration.Application.PoolSize {
		Configuration.Application.PoolSize = poolSize
	}

	return nil
}

func configureNetworkConfig(context *cli.Context) (err error) {
	if network := context.GlobalString("network"); network != "" && network != Configuration.Network.Name {
		Configuration.Network.Name = network
	}

	Configuration.Network.Name = sdkNetwork.NormalizedNetworkName(Configuration.Network.Name)
	if Configuration.Network.Name == "" {
		return errors.New("you need to specify a valid network name to use! Valid options: localnet, devnet, testnet, staking or mainnet")
	}

	Configuration.Network.Mode = strings.ToLower(Configuration.Network.Mode)
	mode := strings.ToLower(context.GlobalString("mode"))
	if mode != "" && mode != Configuration.Network.Mode {
		Configuration.Network.Mode = mode
	}

	Configuration.Network.API = sdkNetwork.Network{
		Name: Configuration.Network.Name,
		Mode: Configuration.Network.Mode,
	}

	Configuration.Network.API.Initialize()

	Configuration.Network.Node = Configuration.Network.API.NodeAddress(0)

	shardingStructure, err := sharding.Structure(Configuration.Network.Node)
	if err != nil {
		return err
	}

	Configuration.Network.Shards = len(shardingStructure)

	Configuration.Network.RPC, err = Configuration.Network.API.RPCClient(0)
	if err != nil {
		return err
	}

	return nil
}

func loadYamlConfig(path string) error {
	Configuration = Config{}

	yamlData, err := utils.ReadFileToString(path)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(yamlData), &Configuration)

	if err != nil {
		return err
	}

	return nil
}
