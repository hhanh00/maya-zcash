#!/bin/sh
set -x
./src/zcashd --datadir=regtest --daemon
sleep 10
./src/zcash-cli --datadir=regtest generate 150
./src/zcash-cli --datadir=regtest z_getnewaccount
./src/zcash-cli --datadir=regtest z_getaddressforaccount 0
UA=`./src/zcash-cli --datadir=regtest listaddresses | jq -r '.[0].unified[0].addresses[0].address'`
./src/zcash-cli --datadir=regtest z_shieldcoinbase '*' $UA
sleep 5
./src/zcash-cli --datadir=regtest z_getoperationresult
./src/zcash-cli --datadir=regtest generate 10
sleep 1
./src/zcash-cli --datadir=regtest z_sendmany $UA '[{"address": "tmWksakBYGg7Lqtm1EqSqvPkVYJHYxGq6Za", "amount": 5.40}]' 1 null 'AllowRevealedRecipients'
sleep 5
./src/zcash-cli --datadir=regtest z_getoperationresult
./src/zcash-cli --datadir=regtest generate 10
sleep 1
