package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/okex/adventure/common"
	"github.com/okex/adventure/common/client"
)

/*
*
Iterate through accounts, get account balance, transfer, then get user balance again, verify balance change, balance decrease
*/
func GetBalTxBal(p BasepParam, e func(ethcmm.Address) []TxParam) {
	clients := client.GenerateClients(p.ips)    // generate CosmosClient or EthClient
	accounts := generateAccounts(p.privateKeys) // generate accounts

	for i := 0; i < p.concurrency; i++ {
		go func(gIndex int) {
			for j := 0; ; j++ {
				aIndex := (gIndex + j*p.concurrency) % len(accounts) // make sure accounts will be picked in order by round-robin
				acc := accounts[aIndex]
				cli := clients[aIndex%len(clients)]
				execBalTxBal(gIndex, cli, acc, e)
				time.Sleep(time.Millisecond * time.Duration(p.sleep))
			}
		}(i)
	}

	select {}
}

func execBalTxBal(gIndex int, cli client.Client, acc *EthAccount, e func(ethcmm.Address) []TxParam) {
	//Get balance
	resp, _ := GetAccBalance(gIndex, acc, TestNetUrl)
	//Slice to remove ""
	bal1 := string(resp.Result)[1 : len(string(resp.Result))-1]
	//Execute tx
	execute(gIndex, cli, acc, e)
	//Get balance again
	resp2, _ := GetAccBalance(gIndex, acc, TestNetUrl)
	//Slice to remove ""
	bal2 := string(resp2.Result)[1 : len(string(resp2.Result))-1]

	//Verify bal2 is less than bal1
	bRet := AssertCompare(bal1, bal2, "bal1 should be greater than or equal to bal2")
	sRet := "FAIL"
	if bRet == true {
		sRet = "SUCCESS"
	}
	log.Println(fmt.Errorf("[g%d] finish to do execBalTxBal assertion, and %s", gIndex, sRet))
}

/**
Hexadecimal data comparison using Big.Int
*/

func AssertCompare(val1 string, val2 string, errInfo string) (bRet bool) {
	if strings.Contains(val1, "0x") {
		val1 = val1[2:len(val1)]
	}
	if strings.Contains(val2, "0x") {
		val2 = val2[2:len(val2)]
	}
	bRet = false
	a, err1 := hex.DecodeString(val1)
	b, err2 := hex.DecodeString(val2)
	if err1 != nil || err2 != nil {
		log.Println(fmt.Errorf("err1 is : %s; err2 is : %s", err1, err2))
	}
	intA := new(big.Int).SetBytes(a)
	intB := new(big.Int).SetBytes(b)
	n := intA.Cmp(intB)
	if n > 0 || n == 0 {
		bRet = true
		return bRet
	}
	log.Println(fmt.Errorf("fail to assert, error happen: %s, intA is: %s; intB is : %s", errInfo, intA, intB))
	return bRet
}

func GetAccBalance(gIndex int, acc *EthAccount, url string) (rpcResp *RPCResp, err error) {

	params := make([]string, 0, 5)
	//Construct request
	address := common.GetEthAddressFromPK(acc.GetPrivateKey())
	res, _ := GetBlockNumber(gIndex, acc, url)
	params = append(params, address.String())

	if res.Result == nil {
		log.Println(fmt.Errorf("Error: block num is nil "))
	}
	//Note here, the result obtained contains two quotes, remove them through slicing
	params = append(params, string(res.Result)[1:len(string(res.Result))-1])
	//log.Println(fmt.Errorf("params[0] is : %s;  params[1] is : %s; ", params[0], params[1]))
	//Call function, get return
	rpcResp, err = EthGetBalanceApi(url, params)

	if err != nil {
		log.Println(fmt.Errorf("[g%d] failed to get account balance, error: %s", gIndex, err))
		return nil, err
	}
	return rpcResp, nil
}

func GetBlockNumber(gIndex int, acc *EthAccount, url string) (rpcResp *RPCResp, err error) {

	rpcResp, err = EthBlockNumberApi(url)
	if err != nil {
		log.Println(fmt.Errorf("[g%d] failed to get block number, error: %s", gIndex, err))
		return nil, err
	}
	return rpcResp, nil
}

func RespToMap(resp *http.Response) map[string]interface{} {
	tmp := make(map[string]interface{})
	r, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		log.Println(fmt.Errorf("Response to map is err: %s", e))
	}
	json.Unmarshal([]byte(string(r)), &tmp)
	return tmp
}
