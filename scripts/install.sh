#!/usr/bin/env bash

echo "Installing Harmony Validator Spammer + hmy"
curl -LO https://harmony.one/hmycli && mv hmycli hmy && chmod u+x hmy
curl -LOs http://tools.harmony.one.s3.amazonaws.com/release/linux-x86_64/harmony-validator-spammer && chmod u+x harmony-validator-spammer
curl -LOs https://raw.githubusercontent.com/SebastianJ/harmony-validator-spammer/master/config.yml
curl -LOs https://raw.githubusercontent.com/SebastianJ/harmony-validator-spammer/master/staking.yml
echo "Harmony Validator Spammer is now ready to use!"
echo "Invoke it using ./harmony-validator-spammer - see ./harmony-validator-spammer --help for all available options"
