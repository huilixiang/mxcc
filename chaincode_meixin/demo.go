package main

import (
	"bytes"
	"time"
	//"errors"
	"encoding/json"
	"fmt"
	//"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	//"strings"
)

var logger = logging.MustGetLogger("demo")
var chainID = "mxcc4"
var idArr = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

type ctorMsg struct {
	Function string   `json:"function"`
	Args     []string `json:"args"`
}
type params struct {
	Type        int                `json:"type"`
	ChaincodeID protos.ChaincodeID `json:"chaincodeID"`
	CtorMsg     ctorMsg            `json:"ctorMsg"`
}

type requestJson struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  params `json:"params"`
	Id      int    `json:"id,omitempty"`
}

type resultJson struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type responseJson struct {
	Jsonrpc string     `json:"jsonrpc"`
	Result  resultJson `json:"result"`
	Id      int        `json:"id"`
}

func buildRequestBody(method string, chain string, function string, args []string) ([]byte, error) {
	ccid := protos.ChaincodeID{}
	ccid.Name = chain
	cm := ctorMsg{}
	cm.Function = function
	cm.Args = args
	pm := params{}
	pm.Type = 1
	pm.ChaincodeID = ccid
	pm.CtorMsg = cm

	rj := requestJson{}
	rj.Jsonrpc = "2.0"
	rj.Method = method
	rj.Params = pm
	rj.Id = 1
	b, err := json.Marshal(rj)
	if err != nil {
		fmt.Printf("json.Marshal(requestJson) err:[%s]", err)
		return nil, err
	}
	//logger.Debugf("request json: %s", string(b))
	return b, nil
}

func parseResp(rb []byte) (*responseJson, error) {
	rj := &responseJson{}
	err := json.Unmarshal(rb, rj)
	if err != nil {
		//logger.Warningf("parse response json err:[%s]", err)
		return nil, err
	}
	return rj, err
}

//init chaincode
func deploy() ([]byte, error) {
	args := []string{"a", "b", "c"}
	return ccClient("deploy", chainID, "init", args)
}

func queryBalance(function string, uid string) ([]byte, error) {
	args := []string{uid}
	return ccClient("query", chainID, function, args)
}

func invokePayerDeposit(uid string, amount string) ([]byte, error) {
	args := []string{uid, amount}
	function := "payerDeposit"
	return ccClient("invoke", chainID, function, args)
}

func invokeShareOrDownload(uid string, actType string, app string, bonusType string, prevTX string, function string) ([]byte, error) {
	args := []string{uid, actType, app, bonusType, prevTX}
	return ccClient("invoke", chainID, function, args)
}

func ccClient(method string, chain string, function string, args []string) ([]byte, error) {
	b, err := buildRequestBody(method, chain, function, args)
	if err != nil {
		fmt.Println("error json.marshal request json")
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	resp, err := http.Post("http://0.0.0.0:7050/chaincode", "application/json;charset=utf-8", buf)
	if err != nil {
		logger.Warningf("[%s] http.Post err:[%s]", function, err)
		return nil, err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warningf("read resp err:[%s]", err)
		return nil, err
	}
	logger.Debugf(" response:[%s]", result)
	respJson, err := parseResp([]byte(result))
	if err != nil {
		logger.Warningf("error parsing reponse for [%s], err:[%s]", function, err)
		return nil, err
	}
	if respJson.Result.Status == "OK" {
		logger.Warningf("[%s] success", function)
		return []byte(respJson.Result.Message), nil
	} else {
		logger.Warningf("[%s] failed", function)
		return []byte(respJson.Result.Message), nil
	}

}

func main() {
	var msg []byte
	var err error
	var payerID = "payer_a"
	//chaincode init
	msg, err = deploy()
	time.Sleep(2 * time.Second)
	//query init balance
	msg, err = queryBalance("payerBalance", payerID)
	if err == nil {
		logger.Debugf("payer:[%s]`balance is [%s]", payerID, string(msg))
	}
	/*
		msg, err = queryBalance("payeeBalance", "payee_a")
		if err == nil {
			logger.Debugf("payee:[%s]`balance is [%s]", "payer_a", string(msg))
		}
	*/
	//payer deposit
	msg, err = invokePayerDeposit(payerID, "200")
	if err == nil {
		logger.Debugf("invokePayerDeposit transaction id:[%s]", string(msg))
	}
	msg, err = queryBalance("payerBalance", payerID)
	if err == nil {
		logger.Debugf("payer:[%s]`balance is [%s]", payerID, string(msg))
	}
	prevTX := ""
	//prevTX 为空表明此用户为分享链的第一个用户
	for i := 0; i < 9; i++ {
		curID := idArr[i]
		msg, err = invokeShareOrDownload("payee_"+curID, "0", "gome+", "1", prevTX, "share")
		if err == nil {
			prevTX = string(msg)
			logger.Debugf("share transaction id:[%s]", prevTX)
		}
	}
	time.Sleep(5 * time.Second)
	//只有分享是不会有返利的
	for i := 0; i < 10; i++ {
		curID := idArr[i]
		msg, err = queryBalance("payeeBalance", "payee_"+curID)
		if err == nil {
			logger.Debugf("payee:[%s]`balance is [%s]", "payee_"+curID, string(msg))
		}
	}
	//return
	//download 发生了
	msg, err = invokeShareOrDownload("payee_"+idArr[9], "1", "gome+", "1", prevTX, "download")
	if err == nil {
		prevTX = string(msg)
		logger.Debugf("download transaction id:[%x]", prevTX)
	}
	time.Sleep(20 * time.Second)
	for i := 0; i < 10; i++ {
		curID := idArr[i]
		msg, err = queryBalance("payeeBalance", "payee_"+curID)
		if err == nil {
			logger.Debugf("payee:[%s]`balance is [%s]", "payee_"+curID, string(msg))
		}
	}
	msg, err = queryBalance("payerBalance", payerID)
	if err == nil {
		logger.Debugf("payer:[%s]`balance is [%s]", payerID, string(msg))
	}

}
