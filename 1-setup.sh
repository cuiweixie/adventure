#!/bin/bash

# Capture stdout and stderr, and remove ANSI color codes
adventure evm bench erc20-init 100000000000000000000 -i http://127.0.0.1:8545 -a ./config/devnet/addr_20000_wmt -s 815405dddb0e2a99b12af775fd2929e526704e1d1aea6a0b4e74dc33e2f7fcd2 2>&1 | tee temp_output.txt

# Extract contract address from output, remove ANSI color codes
contract_address=$(cat temp_output.txt | sed 's/\x1b\[[0-9;]*m//g' | grep -o '0x[a-fA-F0-9]\{40\}' | tail -n 1)

# Check if contract address was successfully extracted
if [ -n "$contract_address" ]; then
    echo "CONTRACT_ADDRESS=$contract_address" > .env
    echo "Contract address has been written to .env file: $contract_address"
else
    echo "Error: Failed to extract contract address from output"
    echo "Please check temp_output.txt file content:"
    cat temp_output.txt
fi
