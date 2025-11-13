#!/bin/bash

# 捕获 stdout 和 stderr，并去除 ANSI 颜色代码
adventure evm bench erc20-init 100000000000000000000000 -i http://127.0.0.1:8123 -a ./config/devnet/acc_pri_10_combined_30 -s 5e3b90aaeec4e3d183044af82167ee799cd3ff5abdd912b9dd89b76e0fd006c4 2>&1 | tee temp_output.txt
# 从输出中提取合约地址，去除 ANSI 颜色代码
contract_address=$(cat temp_output.txt | sed 's/\x1b\[[0-9;]*m//g' | grep -o '0x[a-fA-F0-9]\{40\}' | tail -n 1)

# 检查是否成功提取到合约地址
if [ -n "$contract_address" ]; then
    echo "CONTRACT_ADDRESS=$contract_address" > .env
    echo "合约地址已写入 .env 文件: $contract_address"
else
    echo "错误: 未能从输出中提取到合约地址"
    echo "请检查 temp_output.txt 文件内容:"
    cat temp_output.txt
fi
