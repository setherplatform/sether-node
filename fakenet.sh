#!/bin/bash

./build/sether --fakenet 1/1 --tracenode --http.addr 0.0.0.0 --http --http.corsdomain "*" \
    --http.addr 0.0.0.0 --http.vhosts="*" --ws --ws.origins "*" \
    --http.api=eth,web3,net,txpool,sethn,abft,debug \
    --ws.api=eth,web3,net,txpool,sethn,abft
