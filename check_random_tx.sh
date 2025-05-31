#!/bin/bash

# 配置
RPC_URL="http://127.0.0.1:8545"

echo "🎲 开始随机检查交易状态..."
echo "RPC节点: $RPC_URL"
echo "================================"

# 1. 获取最新区块高度
echo "📊 获取最新区块高度..."
latest_block=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  $RPC_URL | jq -r '.result')

if [ "$latest_block" == "null" ] || [ -z "$latest_block" ]; then
    echo "❌ 错误: 无法获取最新区块高度"
    exit 1
fi

# 转换为十进制显示
latest_block_decimal=$(printf "%d" $latest_block)
echo "✅ 最新区块高度: $latest_block_decimal ($latest_block)"

# 2. 获取最新区块信息
echo ""
echo "📦 获取最新区块信息..."
block_info=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$latest_block\",false],\"id\":1}" \
  $RPC_URL)

# 检查区块是否存在
if [ "$(echo $block_info | jq -r '.result')" == "null" ]; then
    echo "❌ 错误: 无法获取区块信息"
    exit 1
fi

# 3. 检查交易数量
tx_count=$(echo $block_info | jq -r '.result.transactions | length')
echo "📝 区块中交易数量: $tx_count"

if [ "$tx_count" -eq 0 ]; then
    echo "❌ 错误: 区块中没有交易"
    exit 1
fi

# 4. 随机选择一笔交易
random_index=$((RANDOM % tx_count))
random_tx_hash=$(echo $block_info | jq -r ".result.transactions[$random_index]")
echo "🎲 随机选择第 $((random_index + 1)) 笔交易 (索引: $random_index)"
echo "🔗 交易hash: $random_tx_hash"

# 5. 获取交易收据
echo ""
echo "📋 获取交易收据..."
tx_receipt=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"$random_tx_hash\"],\"id\":1}" \
  $RPC_URL)

# 检查收据是否存在
if [ "$(echo $tx_receipt | jq -r '.result')" == "null" ]; then
    echo "❌ 错误: 无法获取交易收据"
    exit 1
fi

# 6. 提取并显示status字段
tx_status=$(echo $tx_receipt | jq -r '.result.status')
gas_used=$(echo $tx_receipt | jq -r '.result.gasUsed')
gas_used_decimal=$(printf "%d" $gas_used)

echo ""
echo "================================"
echo "📊 交易状态结果:"
echo "区块高度: $latest_block_decimal"
echo "交易索引: $random_index"
echo "交易位置: $((random_index + 1))/$tx_count"
echo "交易hash: $random_tx_hash"
echo "Gas使用量: $gas_used_decimal ($gas_used)"
echo "交易状态: $tx_status"

# 解释状态
if [ "$tx_status" == "0x1" ]; then
    echo "✅ 状态说明: 交易成功"
elif [ "$tx_status" == "0x0" ]; then
    echo "❌ 状态说明: 交易失败"
else
    echo "⚠️  状态说明: 未知状态 ($tx_status)"
fi

echo "================================"
echo "🎉 检查完成!"
