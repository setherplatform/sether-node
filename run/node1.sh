#!/bin/bash

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd -P)

NODE_ID="node1"
NODE_DIR=$SCRIPT_DIR/$NODE_ID

mkdir "$NODE_DIR"
cp -r "$SCRIPT_DIR"/keystore "$NODE_DIR"
cp "$NODE_ID.key" "$NODE_DIR/node.key"

"$SCRIPT_DIR/../build/sether" --testnet --port 22221 --genesis "$SCRIPT_DIR/privnet.genesis" \
  --identity "node1" --nodekey "$NODE_DIR/node.key" \
  --datadir "$NODE_DIR" --verbosity=3 \
  --bootnodes "" \
  --validator.id 1 \
  --validator.pubkey "0xc004135f0d5860bc30341d22cc44f3007d0bf35ee815cc827215c96d7d9aba0fb906c5e8c6eb651cf24625b9151cb2e919259b4f006301292d158e5415d66564b81e" \
  --validator.password "$NODE_DIR/keystore/validator/password"

