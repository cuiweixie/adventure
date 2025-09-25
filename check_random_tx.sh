#!/bin/bash

# Configuration
RPC_URL="http://127.0.0.1:8545"

echo "🎲 Starting random transaction status check..."
echo "RPC Node: $RPC_URL"
echo "================================"

# 1. Get latest block height
echo "📊 Getting latest block height..."
latest_block=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  $RPC_URL | jq -r '.result')

if [ "$latest_block" == "null" ] || [ -z "$latest_block" ]; then
    echo "❌ Error: Unable to get latest block height"
    exit 1
fi

# Convert to decimal for display
latest_block_decimal=$(printf "%d" $latest_block)
echo "✅ Latest block height: $latest_block_decimal ($latest_block)"

# 2. Get latest block information
echo ""
echo "📦 Getting latest block information..."
block_info=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$latest_block\",false],\"id\":1}" \
  $RPC_URL)

# Check if block exists
if [ "$(echo $block_info | jq -r '.result')" == "null" ]; then
    echo "❌ Error: Unable to get block information"
    exit 1
fi

# 3. Check transaction count
tx_count=$(echo $block_info | jq -r '.result.transactions | length')
echo "📝 Transaction count in block: $tx_count"

if [ "$tx_count" -eq 0 ]; then
    echo "❌ Error: No transactions in block"
    exit 1
fi

# 4. Randomly select a transaction
random_index=$((RANDOM % tx_count))
random_tx_hash=$(echo $block_info | jq -r ".result.transactions[$random_index]")
echo "🎲 Randomly selected transaction #$((random_index + 1)) (index: $random_index)"
echo "🔗 Transaction hash: $random_tx_hash"

# 5. Get transaction receipt
echo ""
echo "📋 Getting transaction receipt..."
tx_receipt=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"$random_tx_hash\"],\"id\":1}" \
  $RPC_URL)

# Check if receipt exists
if [ "$(echo $tx_receipt | jq -r '.result')" == "null" ]; then
    echo "❌ Error: Unable to get transaction receipt"
    exit 1
fi

# 6. Extract and display status field
tx_status=$(echo $tx_receipt | jq -r '.result.status')
gas_used=$(echo $tx_receipt | jq -r '.result.gasUsed')
gas_used_decimal=$(printf "%d" $gas_used)

echo ""
echo "================================"
echo "📊 Transaction Status Result:"
echo "Block Height: $latest_block_decimal"
echo "Transaction Index: $random_index"
echo "Transaction Position: $((random_index + 1))/$tx_count"
echo "Transaction Hash: $random_tx_hash"
echo "Gas Used: $gas_used_decimal ($gas_used)"
echo "Transaction Status: $tx_status"

# Explain status
if [ "$tx_status" == "0x1" ]; then
    echo "✅ Status Description: Transaction Successful"
elif [ "$tx_status" == "0x0" ]; then
    echo "❌ Status Description: Transaction Failed"
else
    echo "⚠️  Status Description: Unknown Status ($tx_status)"
fi

echo "================================"
echo "🎉 Check Complete!"
