# Harmony Tx Sender
Harmony tx sender is a tool to bulk send transactions on Harmony's blockchain.

## Prerequisites
You need to import an existing key with funds to the keystore.

hmy is automatically downloaded as a part of the installation script.

Import a key using the following command:
```
./hmy keys import-ks ABSOLUTE_PATH_TO_YOUR_KEY NAME_OF_YOUR_KEY --passphrase ""
```

Find the address of your newly imported key:
```
./hmy keys list
```

## Installation

```
bash <(curl -s -S -L https://raw.githubusercontent.com/SebastianJ/harmony-validator-spammer/master/scripts/install.sh)
```

The installer script will also create the data/ folder where you'll find the files receivers.txt and data.txt

`data/receivers.txt` is the file where you enter the receiver accounts you want the tx sender to send tokens to
`data/data.txt` is the tx data that you want the tx sender to use for every transaction it sends.

## Usage
```
./harmony-validator-spammer --from YOUR_SENDER_ACCOUNT_ADDRESS --from-shard 0 --to-shard 0 --count 1000 --pool-size 100
```

### All options:

```
NAME:
   Harmony Tx Sender - stress test and bulk transaction sending tool - Use --help to see all available arguments

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   go1.13.7/darwin-amd64

AUTHOR:
   Sebastian Johnsson

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --mode value                    How to execute transactions - synchronously or asynchronously (possible values: sync, async)
   --network value                 Which network to use (valid options: localnet, devnet, testnet, mainnet)
   --path value                    The path relative to the binary where config.yml and data files can be found (default: "./")
   --from value                    Which address to send tokens from (must exist in the keystore)
   --from-shard value              What shard to send tokens from (default: 0)
   --passphrase value              Passphrase to use for unlocking the keystore
   --to-shard value                What shard to send tokens to (default: 0)
   --amount value                  How many tokens to send per tx
   --count value                   How many transactions to send in total (default: 1000)
   --pool-size value               How many transactions to send simultaneously (default: 100)
   --confirmation-wait-time value  How long to wait for transactions to get confirmed (default: 0)
   --help, -h                      show help
   --version, -v                   print the version
```
