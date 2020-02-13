package config

import (
	sdkNetwork "github.com/SebastianJ/harmony-sdk/network"
	tfConfig "github.com/SebastianJ/harmony-tf/config"
	tfTesting "github.com/SebastianJ/harmony-tf/testing"
	goSdkRpc "github.com/harmony-one/go-sdk/pkg/rpc"
)

// Config - represents the config
type Config struct {
	BasePath    string                      `yaml:"-"`
	Verbose     bool                        `yaml:"-"`
	Network     Network                     `yaml:"network"`
	Application Application                 `yaml:"network"`
	Funding     tfConfig.Funding            `yaml:"funding"`
	Staking     tfTesting.StakingParameters `yaml:"staking"`
}

// Application - represents the transactions settings group
type Application struct {
	Infinite bool `yaml:"infinite"`
	Count    int  `yaml:"count"`
	PoolSize int  `yaml:"pool_size"`
}

// Network - represents the network settings group
type Network struct {
	Name   string                  `yaml:"name"`
	Mode   string                  `yaml:"mode"`
	Node   string                  `yaml:"-"`
	Shards int                     `yaml:"-"`
	RPC    *goSdkRpc.HTTPMessenger `yaml:"-"`
	API    sdkNetwork.Network      `yaml:"-"`
}
