package config

import (
	"errors"
	"path/filepath"
	"strings"

	sdkAccounts "github.com/SebastianJ/harmony-sdk/accounts"
	sdkNetwork "github.com/SebastianJ/harmony-sdk/network"
	tfConfig "github.com/SebastianJ/harmony-tf/config"
	tfTesting "github.com/SebastianJ/harmony-tf/testing"
	tfUtils "github.com/SebastianJ/harmony-tf/utils"
	"github.com/SebastianJ/harmony-validator-spammer/utils"
	"github.com/harmony-one/go-sdk/pkg/sharding"
	"github.com/urfave/cli"
)

// Configuration - the central configuration for the test suite tool
var Configuration Config

// Staking - staking and validator settings used for creating validators
var Staking tfTesting.StakingParameters

// Configure - configures the test suite tool using a combination of the YAML config file as well as command arguments
func Configure(basePath string, context *cli.Context) (err error) {
	configPath := filepath.Join(basePath, "config.yml")
	if err = loadYamlConfig(configPath); err != nil {
		return err
	}

	stakingPath := filepath.Join(basePath, "staking.yml")
	if err = loadStakingConfig(stakingPath); err != nil {
		return err
	}

	if Configuration.BasePath == "" {
		Configuration.BasePath = basePath
	}

	Configuration.Application.Verbose = true           // this configures the output using Harmony TF's logger
	sdkNetwork.Verbose = context.GlobalBool("verbose") // this configures the raw tx dump logs from Harmony-SDK

	if err = configureNetworkConfig(context); err != nil {
		return err
	}

	if err = configureBaseConfig(context); err != nil {
		return err
	}

	tfConfig.ConfigureStylingConfig()

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

	Configuration.Network.Gas.Initialize()
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

	tfConfig.Configuration.Network = tfConfig.Network{
		Name:   Configuration.Network.Name,
		Mode:   Configuration.Network.Mode,
		Node:   Configuration.Network.Node,
		Shards: Configuration.Network.Shards,
		API:    Configuration.Network.API,
		Gas:    Configuration.Network.Gas,
	}

	return nil
}

func configureBaseConfig(context *cli.Context) error {
	fromAddress := context.GlobalString("from")
	if fromAddress == "" {
		return errors.New("you need to specify the sender address")
	}

	Configuration.Funding.Account.Address = fromAddress

	if Configuration.Funding.Account.Name == "" {
		Configuration.Funding.Account.Name = sdkAccounts.FindAccountNameByAddress(Configuration.Funding.Account.Address)
	}

	if passphrase := context.GlobalString("passphrase"); passphrase != "" && passphrase != Configuration.Application.Passphrase {
		Configuration.Application.Passphrase = passphrase
	}

	Configuration.Funding.Account.Passphrase = Configuration.Application.Passphrase
	tfConfig.Configuration.Account.Passphrase = Configuration.Application.Passphrase

	Configuration.Funding.Gas.Initialize()

	tfConfig.Configuration.Funding = tfConfig.Funding{
		Account:              Configuration.Funding.Account,
		ConfirmationWaitTime: Configuration.Funding.ConfirmationWaitTime,
		Attempts:             Configuration.Funding.Attempts,
		Gas:                  Configuration.Funding.Gas,
	}

	tfConfig.Configuration.Funding.ConfirmationWaitTime = tfUtils.ConfirmationWaitTimeNetworkAdjustment(tfConfig.Configuration.Network.Name, tfConfig.Configuration.Funding.ConfirmationWaitTime)

	Configuration.Application.Infinite = context.GlobalBool("infinite")

	if count := context.GlobalInt("count"); count > 0 && count != Configuration.Application.Count {
		Configuration.Application.Count = count
	}

	if poolSize := context.GlobalInt("pool-size"); poolSize > 0 && poolSize != Configuration.Application.PoolSize {
		Configuration.Application.PoolSize = poolSize
	}

	// Initialize the staking params - this converts values to numeric.Dec etc.
	Staking.Initialize()
	Staking.ConfirmationWaitTime = tfUtils.ConfirmationWaitTimeNetworkAdjustment(tfConfig.Configuration.Network.Name, Staking.ConfirmationWaitTime)

	return nil
}

func loadYamlConfig(path string) error {
	Configuration = Config{}
	return utils.ParseYaml(path, &Configuration)
}

func loadStakingConfig(path string) error {
	Staking = tfTesting.StakingParameters{}
	return utils.ParseYaml(path, &Staking)
}
