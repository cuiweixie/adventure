# Adventure

## 1. 编译
```shell
make
```

## 2. 操作
### 2.1 初始化账户
```shell
adventure evm batch-transfer 10 -i ${ip} -s ${private_key} -a ${address_file}
```
* -i: ip地址
  * 必填
  * 支持cosmos端口、eth端口
* -s: 私钥 (不是助记词)
  * 必填
  * 对应地址，拥有足够的okt
* -a: 账户地址文件路径
  * 选填，如果为空，代码默认内置2000个固定账户
  * 0x地址格式

### 2.2 压力测试
公共参数
* --ips, -i: ip地址列表
  * 必填
  * 支持cosmos、eth两者的域名或ip地址
* --concurrency, -c: 启动的协程数量
  * 选填, 默认1
* --sleep, -t: 单协程的每轮睡眠时间，毫秒
  * 选填, 默认1000ms
* --private-key-file, -p: 账户私钥文件路径
  * 选填, 如果为空，代码默认内置2000个固定账户
  * 私钥 (不是助记词)

#### 2.2.1 转账
```shell
adventure evm bench transfer -i ${ip1},${ip2},${ip3} -c 100 -p ${private_key_file}
```

* --fixed, -f: 转账to地址是否固定一个
  * 选填
  * false, 默认, 每个账户转到对应的一个固定地址
  * true, 所有交易均转到同一个地址

#### 2.2.2 压力测试合约
```shell
adventure evm bench operate -i ${ip1},${ip2},${ip3} -c 100 --opts 1,1,1,1,1 --times 1 --contract 0x6cc0277c979325800294774d7ae478A96B824271 --id 0 
```

* --contract: router合约地址或测试合约地址
* --direct: 默认false; 设置为true时，工具会直接往测试合约发tx，而不是router合约地址；--id就不需要设置，--contract直接设置为具体的合约地址
* --id: 测试合约id
* --opts: 每个操作码在单次循环的执行次数
* --times: 循环次数

#### 2.2.3 测试网uniswap挖卖提
```shell
adventure evm bench wmt -i ${ip1},${ip2},${ip3}  -c 250     
```

#### 2.2.4 测试网查询
```shell
 adventure evm bench query -i https://exchaintestrpc.okex.org -t 1000 -o 1,1,1,1,1,1,1,1,1
```

* -o: 每个查询接口在每秒创建的协程数量
  * 0: eth_blockNumber
  * 1: eth_getBalance
  * 2: eth_getBlockByNumber
  * 3: eth_gasPrice
  * 4: eth_getCode
  * 5: eth_getTransactionCount
  * 6: eth_getTransactionReceipt
  * 7: net_version
  * 8: eth_call


## wasm 压测
### 压测 cw20 转账
在对应目录准备好配置文件，如下：
```json
{
  "restUrls": [
    "http://127.0.0.1:8545"
  ],
  "tendermintUrls": [
    "http://127.0.0.1:26657"
  ],
  "wasmFilePath": "./config/devnet/wasm_contract/cw20.wasm",
  "contractAddress": "",
  "privateKeysFile": "./config/devnet/acc_pri_10",
  "concurrentNum": 1,
  "threshold": 180000
}
```

生成默认模板配置，将在控制台输入模板配置文件信息
```shell
 adventure wasm bench cw20-config
```

执行命令开启压测，如果合约地址位空，将自动部署合约
```shell
adventure wasm bench cw20 --f config/devnet/cw20-local.json
```

### 压测 OKT 原生代币转账
在对应目录准备好配置文件，如下：
```json
{
  "restUrls": [
    "http://127.0.0.1:8545"
  ],
  "tendermintUrls": [
    "http://127.0.0.1:26657"
  ],
  "wasmFilePath": "./config/devnet/wasm_contract/okt.wasm",
  "contractAddress": "",
  "privateKeysFile": "./config/devnet/acc_pri_10",
  "concurrentNum": 1,
  "threshold": 180000
}
```

生成默认模板配置，将在控制台输入模板配置文件信息
```shell
 adventure wasm bench okt-config
```

执行命令开启压测，如果合约地址位空，将自动部署合约
```shell
adventure wasm bench okt --f config/devnet/cwokt-local.json
```

