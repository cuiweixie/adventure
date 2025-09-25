# Adventure

## 1. Compilation
```shell
make
```

## 2. Operations
### 2.1 Account Initialization
```shell
adventure evm batch-transfer 10 -i ${ip} -s ${private_key} -a ${address_file}
```
* -i: IP address
  * Required
  * Supports cosmos port and eth port
* -s: Private key (not mnemonic)
  * Required
  * Corresponding address should have sufficient OKT
* -a: Account address file path
  * Optional, if empty, code defaults to 2000 built-in fixed accounts
  * Currently supports direct use of private key files and 0x address format

### 2.2 Load Testing
Common Parameters
* --ips, -i: IP address list
  * Required
  * Supports cosmos and eth domain names or IP addresses
* --concurrency, -c: Number of goroutines to start
  * Optional, default is 1
* --sleep, -t: Sleep time per round for single goroutine, in milliseconds
  * Optional, default is 1000ms
* --private-key-file, -p: Account private key file path
  * Optional, if empty, code defaults to 2000 built-in fixed accounts
  * Private key (not mnemonic)

#### 2.2.1 Transfer
```shell
adventure evm bench transfer -i ${ip1},${ip2},${ip3} -c 100 -p ${private_key_file}
```

* --fixed, -f: Whether to fix the transfer destination address
  * Optional
  * false, default, each account transfers to a corresponding fixed address
  * true, all transactions transfer to the same address

#### 2.2.2 Contract Load Testing
```shell
adventure evm bench operate -i ${ip1},${ip2},${ip3} -c 100 --opts 1,1,1,1,1 --times 1 --contract 0x6cc0277c979325800294774d7ae478A96B824271 --id 0
```

* --contract: Router contract address or test contract address
* --direct: Default false; when set to true, the tool will send transactions directly to the test contract instead of the router contract address; --id doesn't need to be set, --contract is set directly to the specific contract address
* --id: Test contract ID
* --opts: Number of executions for each opcode in a single loop
* --times: Number of loops

#### 2.2.3 Testnet Uniswap Mining/Selling/Withdrawal
```shell
adventure evm bench wmt -i ${ip1},${ip2},${ip3}  -c 250
```

#### 2.2.4 Testnet Query
```shell
 adventure evm bench query -i https://exchaintestrpc.okex.org -t 1000 -o 1,1,1,1,1,1,1,1,1
```

* -o: Number of goroutines created per second for each query interface
  * 0: eth_blockNumber
  * 1: eth_getBalance
  * 2: eth_getBlockByNumber
  * 3: eth_gasPrice
  * 4: eth_getCode
  * 5: eth_getTransactionCount
  * 6: eth_getTransactionReceipt
  * 7: net_version
  * 8: eth_call


## WASM Load Testing
### CW20 Transfer Load Testing
#### Transfer Gas Fees
10 OKT is sufficient
```shell
adventure evm batch-transfer 10 -i http://localhost:8545 -a config/devnet/address_10 -s 8ff3ca2d9985c3a52b459e2f6e7822b23e1af845961e22128d5f372fb9aa5f17
```
Prepare the configuration file in the corresponding directory as follows:
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

Generate default template configuration, template configuration file information will be input in the console
```shell
 adventure wasm bench cw20-config
```

Execute command to start load testing, if contract address is empty, contract will be automatically deployed
```shell
adventure wasm bench cw20 --f config/devnet/cw20-local.json
```

### OKT Native Token Transfer Load Testing
#### Transfer Gas Fees
Transfer as many OKT to each account as there are accounts participating in load testing (account files used in configuration file)
```shell
adventure evm batch-transfer 100000 -i http://localhost:8545 -a config/devnet/address_10 -s 8ff3ca2d9985c3a52b459e2f6e7822b23e1af845961e22128d5f372fb9aa5f17
```

Prepare the configuration file in the corresponding directory as follows:
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

Generate default template configuration, template configuration file information will be input in the console
```shell
 adventure wasm bench okt-config
```

Execute command to start load testing, if contract address is empty, contract will be automatically deployed
```shell
adventure wasm bench okt --f config/devnet/cwokt-local.json
```

### Load Testing EO RO WO
#### Preparation
Load testing is divided into directly executing compute/read/writeTest and executing compute/read/writeTest through router contracts. Load testing contracts are in the config/devnet/wasm_contract directory, and the code can automatically deploy contracts.

Prepare the configuration file as follows:
```json
{
  "PrivateKeysFile": "./config/devnet/acc_pri_10",
  "restUrls": [
    "http://127.0.0.1:8545"
  ],
  "tendermintUrls": [
    "http://127.0.0.1:26657"
  ],
  "contractAddress": "",
  "contractPath": "./config/devnet/wasm_contract/readTest.wasm",
  "routerContractPath": "./config/devnet/wasm_contract/router.wasm",
  "WasmOperRouter": true,
  "concurrentNum": 1,
  "threshold": 1800,
  "chainId": "exchain-67"
}
```
The input items that need explanation are as follows:
* contractAddress: If the contract is already deployed, directly fill in the contract 0x address, otherwise set to empty, and it will be automatically deployed
* contractPath: Path to the contract file
* routerContractPath: Path to the router contract

The cwoperate-config command displays the default template configuration, template configuration file information will be input in the console
```shell
 adventure wasm bench cwoperate-config
```

Example test file path: config/testnet/operate_test.json

#### Account Initialization
Load testing requires private keys to be passed in. These private keys need to have corresponding accounts on the chain and OKT for paying gas, so initial transfers are required.

```shell
adventure evm batch-transfer 100000 -i http://localhost:8545 -a ./config/devnet/acc_pri_10 -s 8ff3ca2d9985c3a52b459e2f6e7822b23e1af845961e22128d5f372fb9aa5f17
```


Execute command to start load testing, if contract address is empty, contract will be automatically deployed
```shell
adventure wasm bench cwoperate --type read --opt 1,1,1,1 --times 1 --f ./config/testnet/operate_test.json
```
Where:
* type field: input the specific operation to execute
* opt field: input the opt required to call the specific operation contract
* times field: input the number of executions