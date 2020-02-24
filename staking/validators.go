package staking

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"

	sdkAccounts "github.com/SebastianJ/harmony-sdk/accounts"
	sdkCrypto "github.com/SebastianJ/harmony-sdk/crypto"
	sdkNetwork "github.com/SebastianJ/harmony-sdk/network"
	sdkTxs "github.com/SebastianJ/harmony-sdk/transactions"
	"github.com/SebastianJ/harmony-tf/accounts"
	"github.com/SebastianJ/harmony-tf/balances"
	"github.com/SebastianJ/harmony-tf/crypto"
	"github.com/SebastianJ/harmony-tf/funding"
	"github.com/SebastianJ/harmony-tf/logger"
	tfStaking "github.com/SebastianJ/harmony-tf/staking"
	"github.com/SebastianJ/harmony-tf/testing"
	"github.com/SebastianJ/harmony-validator-spammer/config"
	"github.com/SebastianJ/harmony-validator-spammer/utils"
	goSdkAccount "github.com/harmony-one/go-sdk/pkg/account"
)

// CreateValidators - mass creates validators
func CreateValidators() {
	rpcClient, _ := config.Configuration.Network.API.RPCClient(0)

	index := 0
	nonce := -1
	receivedNonce := sdkNetwork.CurrentNonce(rpcClient, config.Configuration.Funding.Account.Address)
	nonce = int(receivedNonce)

	if config.Configuration.Application.Infinite {
		for {
			index, nonce = PerformCreateValidators(index, nonce)
		}
	} else {
		pools := 1
		if config.Configuration.Application.Count > config.Configuration.Application.PoolSize {
			pools = int(math.RoundToEven(float64(config.Configuration.Application.Count) / float64(config.Configuration.Application.PoolSize)))
		}

		for poolIndex := 0; poolIndex < pools; poolIndex++ {
			index, nonce = PerformCreateValidators(index, nonce)
		}
	}
}

// PerformCreateValidators - performs the actual creation via goroutines
func PerformCreateValidators(index int, nonce int) (int, int) {
	var waitGroup sync.WaitGroup

	for i := 0; i < config.Configuration.Application.PoolSize; i++ {
		waitGroup.Add(1)

		go CreateValidator(index, nonce, &waitGroup)

		index++
		nonce++
	}

	waitGroup.Wait()

	return index, nonce
}

// CreateValidator - creates a given validator
func CreateValidator(index int, nonce int, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()
	accountName := fmt.Sprintf("ValidatorSpammer_Account_%d", index)
	logger.AccountLog(fmt.Sprintf("Generating a new account: %s", accountName), config.Configuration.Application.Verbose)

	account, err := accounts.GenerateAccount(accountName)
	if err != nil {
		logger.ErrorLog(err.Error(), config.Configuration.Application.Verbose)
		return err
	}

	logger.AccountLog(fmt.Sprintf("Generated account: %s, address: %s", account.Name, account.Address), config.Configuration.Application.Verbose)

	fundingAccountBalance, err := balances.GetShardBalance(config.Configuration.Funding.Account.Address, 0)
	if err != nil {
		logger.ErrorLog(fmt.Sprintf("Failed to retrieve shard balance - error: %s", err.Error()), config.Configuration.Application.Verbose)
		return err
	}

	//testing.FundingLog(fmt.Sprintf("Available funding amount in the funding account %s, address: %s is %f", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, fundingAccountBalance), config.Configuration.Application.Verbose)
	fundingAmount := funding.CalculateFundingAmount(config.Staking.Amount, fundingAccountBalance, 1)
	logger.FundingLog(fmt.Sprintf("Funding account %s, address: %s with %f", account.Name, account.Address, fundingAmount), config.Configuration.Application.Verbose)
	funding.PerformFundingTransaction(&config.Configuration.Funding.Account, 0, account.Address, 0, fundingAmount, nonce, config.Configuration.Funding.Gas.Limit, config.Configuration.Funding.Gas.Price, config.Configuration.Funding.ConfirmationWaitTime, config.Configuration.Funding.Attempts)

	accountStartingBalance, _ := balances.GetShardBalance(account.Address, 0)
	logger.AccountLog(fmt.Sprintf("Using account %s, address: %s to create a new validator", account.Name, account.Address), config.Configuration.Application.Verbose)
	logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has a starting balance of %f in shard %d before the test", account.Name, account.Address, accountStartingBalance, 0), config.Configuration.Application.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending create validator transaction - will wait %d seconds for it to finalize", config.Staking.ConfirmationWaitTime), config.Configuration.Application.Verbose)
	fmt.Println("")

	config.Staking.Validator.Address = account.Address
	blsKeys := crypto.GenerateBlsKeys(config.Staking.BLSKeyCount, "")
	rawTx, err := tfStaking.CreateValidator(account, config.Staking, blsKeys)
	if err != nil {
		logger.ErrorLog(fmt.Sprintf("Failed to create validator - error: %s", err.Error()), config.Configuration.Application.Verbose)
		return err
	}

	tx := sdkTxs.ToTransaction(account.Address, 0, account.Address, 0, rawTx, err)

	if config.Staking.ConfirmationWaitTime > 0 {
		txResultColoring := logger.ResultColoring(tx.Success, true).Render(fmt.Sprintf("tx successful: %t", tx.Success))
		logger.TransactionLog(fmt.Sprintf("Performed create validator - transaction hash: %s, %s", tx.TransactionHash, txResultColoring), config.Configuration.Application.Verbose)
	} else {
		logger.TransactionLog(fmt.Sprintf("Performed create validator - transaction hash: %s", tx.TransactionHash), config.Configuration.Application.Verbose)
	}

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)\n", config.Configuration.Application.Verbose)

	if config.Staking.ConfirmationWaitTime > 0 {
		if tx.Success {
			exportKeys(account, blsKeys)
		}

		testing.Teardown(&account, 0, config.Configuration.Funding.Account.Address, 0)
	} else {
		goSdkAccount.RemoveAccount(account.Name)
	}

	return nil
}

func exportKeys(account sdkAccounts.Account, blsKeys []sdkCrypto.BLSKey) error {
	dirPath := filepath.Join(config.Configuration.BasePath, "generated", account.Address)
	keystorePath := filepath.Join(dirPath, fmt.Sprintf("%s.key", account.Address))

	if err := utils.CreateDirectory(dirPath); err != nil {
		return err
	}

	keystoreJSON, err := account.ExportKeystore(config.Configuration.Application.Passphrase)
	if err != nil {
		return err
	}

	if len(keystoreJSON) > 0 {
		os.Remove(keystorePath)
		ioutil.WriteFile(keystorePath, keystoreJSON, 0755)
	}

	for _, blsKey := range blsKeys {
		blsKeyPath := filepath.Join(dirPath, fmt.Sprintf("%s.key", blsKey.PublicKeyHex))
		encrypted, err := blsKey.Encrypt(config.Configuration.Application.Passphrase)

		if err == nil && len(encrypted) > 0 {
			os.Remove(blsKeyPath)
			ioutil.WriteFile(blsKeyPath, []byte(encrypted), 0755)
		}
	}

	return nil
}
