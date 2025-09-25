package query

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	success = "sucess"
	fail    = "failed"
)

type Request struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Response struct {
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
}

func createRequest(method string, params interface{}) Request {
	return Request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}

func call(postBody []byte, reqType int, ip string) (rpcRes *Response, err error) {
	client := &http.Client{}

	//post request
	req, err := http.NewRequest("POST", ip, bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	resp, reqErr := client.Do(req)
	elapsed := strconv.FormatInt(time.Since(startTime).Milliseconds(), 10) + "ms"
	if reqErr != nil {
		log.Println(reqType, elapsed, fail, reqErr)
		return nil, reqErr
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&rpcRes)
	if err != nil {
		log.Println(reqType, elapsed, fail, err)
		return nil, err
	}
	if rpcRes.Error != nil {
		log.Println(reqType, elapsed, fail, rpcRes.Error)
		return nil, err
	}
	log.Println(reqType, elapsed, success, string(rpcRes.Result))
	return rpcRes, nil
}
