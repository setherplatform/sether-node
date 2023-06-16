#!/bin/bash

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd -P)
rm -rf $SCRIPT_DIR/node1/sether-node
rm -rf $SCRIPT_DIR/node1/chaindata
rm -rf $SCRIPT_DIR/node2/sether-node
rm -rf $SCRIPT_DIR/node2/chaindata
rm -rf $SCRIPT_DIR/node3/sether-node
rm -rf $SCRIPT_DIR/node3/chaindata
