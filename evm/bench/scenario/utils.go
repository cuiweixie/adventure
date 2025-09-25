package scenario

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	URL         string
	AccountList string
	Account     string
}

/*
*
This file provides some basic util functions
*/
func PanicErr(err error) {
	PanicErr(err)
}

/*
*
Read each line in the file and put it into []string
*/
func ReadDataFromFile(path string) (lines []string) {
	f, err := os.Open(path)
	PanicErr(err)
	rd := bufio.NewReader(f)
	lines = make([]string, 0)
	num := 0
	for true {
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		lines = append(lines, strings.TrimSpace(line))
		num++
	}
	log.Printf("Read lines %d from file", num)
	return
}

/*
*
Read json configuration and map to struct
*/
func ReadJson2Struct(filepath string, obj interface{}) {
	file, ferr := os.Stat(filepath)
	if ferr != nil {
		log.Printf("file path does not exist, %s", filepath)
	}
	log.Printf("%v", file)
	b, err := ioutil.ReadFile(filepath)
	PanicErr(err)
	err1 := json.Unmarshal(b, &obj)
	if err != nil {
		log.Printf("error is: %s", err1)
	}
}

const (
	success = "success"
	fail    = "failed"
)

func DoPost(url string, postBody []byte) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	startTime := time.Now()
	resp, respErr := client.Do(req)
	elapsed := time.Since(startTime)

	if respErr != nil {
		log.Println(postBody, strconv.FormatInt(elapsed.Milliseconds(), 10)+"ms", fail, respErr)
		return nil, respErr
	}

	log.Println(postBody, strconv.FormatInt(elapsed.Milliseconds(), 10)+"ms", success)
	defer resp.Body.Close()
	return resp, nil
}

func DoGet(url string, body []byte) {

}
