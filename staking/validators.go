package staking

import (
	"fmt"
	"sync"

	sdkNetwork "github.com/SebastianJ/harmony-sdk/network"
	sdkTxs "github.com/SebastianJ/harmony-sdk/transactions"
	"github.com/SebastianJ/harmony-tf/accounts"
	"github.com/SebastianJ/harmony-tf/balances"
	"github.com/SebastianJ/harmony-tf/crypto"
	"github.com/SebastianJ/harmony-tf/funding"
	tfStaking "github.com/SebastianJ/harmony-tf/staking"
	"github.com/SebastianJ/harmony-tf/testing"
	"github.com/SebastianJ/harmony-validator-spammer/config"
)

// CreateValidators - mass creates validators
func CreateValidators() {
	rpcClient, _ := config.Configuration.Network.API.RPCClient(0)

	nonce := -1
	receivedNonce := sdkNetwork.CurrentNonce(rpcClient, config.Configuration.Funding.Account.Address)
	nonce = int(receivedNonce)

	index := 0

	for {
		var waitGroup sync.WaitGroup

		for i := 0; i < config.Configuration.Application.PoolSize; i++ {
			go CreateValidator(index, nonce, &waitGroup)

			index++
			nonce++
		}
	}
}

// CreateValidator - creates a given validator
func CreateValidator(index int, nonce int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	accountName := fmt.Sprintf("[ValidatorSpammer]_Account_%d", index)
	testing.AccountLog("", fmt.Sprintf("Generating a new account: %s", accountName), config.Configuration.Verbose)
	account := accounts.GenerateTypedAccount(accountName)

	fundingAccountBalance, err := balances.GetShardBalance(config.Configuration.Funding.Account.Address, 0)
	if err != nil {
		testing.ErrorLog("", fmt.Sprintf("Failed to create validator - error: %s", err.Error()), config.Configuration.Verbose)
	}

	testing.AccountLog("", fmt.Sprintf("Generated account: %s, address: %s", account.Name, account.Address), config.Configuration.Verbose)

	fundingAmount := funding.CalculateFundingAmount(config.Configuration.Staking.Amount, fundingAccountBalance, 1)
	testing.FundingLog("", fmt.Sprintf("Available funding amount in the funding account %s, address: %s is %f", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, fundingAccountBalance), config.Configuration.Verbose)
	funding.PerformFundingTransaction(config.Configuration.Funding.Account.Address, 0, account.Address, 0, fundingAmount, nonce, config.Configuration.Funding.Gas.Limit, config.Configuration.Funding.Gas.Price, config.Configuration.Funding.ConfirmationWaitTime, config.Configuration.Funding.Attempts)

	accountStartingBalance, _ := balances.GetShardBalance(account.Address, 0)

	testing.AccountLog("", fmt.Sprintf("Using account %s, address: %s to create a new validator", account.Name, account.Address), config.Configuration.Verbose)
	testing.BalanceLog("", fmt.Sprintf("Account %s, address: %s has a starting balance of %f in shard %d before the test", account.Name, account.Address, accountStartingBalance, 0), config.Configuration.Verbose)

	config.Configuration.Staking.Validator.Address = account.Address
	blsKeys := crypto.GenerateBlsKeys(config.Configuration.Staking.BLSKeyCount)
	fmt.Println("")

	rawTx, err := tfStaking.CreateValidator(config.Configuration.Staking, blsKeys)
	if err != nil {
		testing.ErrorLog("", fmt.Sprintf("Failed to create validator - error: %s", err.Error()), config.Configuration.Verbose)
	}

	tx := sdkTxs.ToTransaction(account.Address, 0, account.Address, 0, rawTx, err)

	txResultColoring := testing.ResultColoring(tx.Success, true).Render(fmt.Sprintf("tx successful: %t", tx.Success))
	testing.TransactionLog("", fmt.Sprintf("Performed create validator - transaction hash: %s, %s", tx.TransactionHash, txResultColoring), config.Configuration.Verbose)
	testing.TeardownLog("", "Performing test teardown (returning funds and removing accounts)\n", config.Configuration.Verbose)

	testing.Teardown(account.Name, account.Address, 0, config.Configuration.Funding.Account.Address, 0)
}
