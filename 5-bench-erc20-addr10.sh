#!/bin/bash

# 检查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "错误: .env 文件不存在，请先运行 1-setup.sh"
    exit 1
fi

# 从 .env 文件中读取合约地址
source .env

# 检查是否成功读取到合约地址
if [ -z "$CONTRACT_ADDRESS" ]; then
    echo "错误: 未能从 .env 文件中读取到 CONTRACT_ADDRESS"
    echo ".env 文件内容:"
    cat .env
    exit 1
fi

echo "使用合约地址: $CONTRACT_ADDRESS"

# 执行 ERC20 基准测试
adventure evm bench erc20 --f ./config/poly_test/fork6_leo_erc20.json --contract $CONTRACT_ADDRESS
