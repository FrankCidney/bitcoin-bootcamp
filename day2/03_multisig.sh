#!/bin/bash

# Try to create 2 wallets. If any of the walletes already exist, just load them instead
bitcoin-cli -regtest createwallet "alice" 2>/dev/null || \
bitcoin-cli -regtest loadwallet "alice" 2>/dev/null

bitcoin-cli -regtest createwallet "bob" 2>/dev/null || \
bitcoin-cli -regtest loadwallet "bob" 2>/dev/null

# Get a public key from each wallet
ALICE_PUBKEY=$(bitcoin-cli -regtest -rpcwallet=alice getaddressinfo \
  $(bitcoin-cli -regtest -rpcwallet=alice getnewaddress) | jq -r '.pubkey')

BOB_PUBKEY=$(bitcoin-cli -regtest -rpcwallet=bob getaddressinfo \
  $(bitcoin-cli -regtest -rpcwallet=bob getnewaddress) | jq -r '.pubkey')

echo "Alice pubkey: $ALICE_PUBKEY"
echo "Bob pubkey:   $BOB_PUBKEY"

# Create the 2-of-2 multisig address
MULTISIG=$(bitcoin-cli -regtest -rpcwallet=alice createmultisig 2 \
  "[\"$ALICE_PUBKEY\",\"$BOB_PUBKEY\"]")

MULTISIG_ADDRESS=$(echo $MULTISIG | jq -r '.address')
REDEEM_SCRIPT=$(echo $MULTISIG | jq -r '.redeemScript')

echo "Multisig address: $MULTISIG_ADDRESS"
echo "Redeem script:    $REDEEM_SCRIPT"

# Func the multisig address
MINER_WALLET=$(bitcoin-cli -regtest createwallet "miner" 2>/dev/null || \
bitcoin-cli -regtest loadwallet "miner" 2>/dev/null)
MINER_ADDRESS=$(bitcoin-cli -regtest -rpcwallet=miner getnewaddress)
bitcoin-cli -regtest generatetoaddress 101 $MINER_ADDRESS

# Send 1 BTC to the multisig address
TXID=$(bitcoin-cli -regtest -rpcwallet=miner sendtoaddress $MULTISIG_ADDRESS 1.0)
echo "Funded multisig, txid: $TXID"

# Mine a block to confirm it
bitcoin-cli -regtest generatetoaddress 1 $MINER_ADDRESS

# Spend from the multisig
# Get the output index of our funding transaction
VOUT=$(bitcoin-cli -regtest getrawtransaction $TXID true | jq -r \
  ".vout | to_entries | .[] | select(.value.scriptPubKey.address==\"$MULTISIG_ADDRESS\") | .key")

# Get an address to send the funds to (back to miner in this case)
DEST_ADDRESS=$(bitcoin-cli -regtest -rpcwallet=miner getnewaddress)

# Create raw transaction (unsigned)
RAW_TX=$(bitcoin-cli -regtest createrawtransaction \
  "[{\"txid\":\"$TXID\",\"vout\":$VOUT}]" \
  "[{\"$DEST_ADDRESS\":0.9999}]")

echo "Raw (unsigned) tx: $RAW_TX"

# Alice signs first
ALICE_SIGNED=$(bitcoin-cli -regtest -rpcwallet=alice signrawtransactionwithwallet \
  $RAW_TX "[{\"txid\":\"$TXID\",\"vout\":$VOUT,\"scriptPubKey\":\"$(bitcoin-cli -regtest getrawtransaction $TXID true | jq -r '.vout['$VOUT'].scriptPubKey.hex')\",\"redeemScript\":\"$REDEEM_SCRIPT\",\"amount\":1.0}]")

ALICE_SIGNED_HEX=$(echo $ALICE_SIGNED | jq -r '.hex')
echo "Alice signed: $ALICE_SIGNED_HEX"

# Bob signs the already-alice-signed transaction
BOB_SIGNED=$(bitcoin-cli -regtest -rpcwallet=bob signrawtransactionwithwallet \
  $ALICE_SIGNED_HEX "[{\"txid\":\"$TXID\",\"vout\":$VOUT,\"scriptPubKey\":\"$(bitcoin-cli -regtest getrawtransaction $TXID true | jq -r '.vout['$VOUT'].scriptPubKey.hex')\",\"redeemScript\":\"$REDEEM_SCRIPT\",\"amount\":1.0}]")

FINAL_TX=$(echo $BOB_SIGNED | jq -r '.hex')
COMPLETE=$(echo $BOB_SIGNED | jq -r '.complete')
echo "Both signed, complete: $COMPLETE"

# Broadcast the fully signed transaction
if [ "$COMPLETE" = "true" ]; then
  SPEND_TXID=$(bitcoin-cli -regtest sendrawtransaction $FINAL_TX)
  echo "Broadcast txid: $SPEND_TXID"
  bitcoin-cli -regtest generatetoaddress 1 $MINER_ADDRESS
  echo "Done! Funds moved out of multisig."
else
  echo "Transaction not fully signed yet"
fi