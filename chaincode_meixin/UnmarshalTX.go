package main

import (
	"encoding/json"
	//"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	//"strconv"
	//"strings"

	//"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/golang/protobuf/proto"
	//"github.com/hyperledger/fabric/core/util"
	"github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
)

type UnmarshalTX struct {
}

var logger = logging.MustGetLogger("UnmarshalTX")

func (t *UnmarshalTX) GetChain() {

}

func main() {
	args := os.Args
	resp, err := http.Get("http://127.0.0.1:7050/chain/blocks/" + args[1])
	if err != nil {
		logger.Warningf("get chain failed, err:[%s]", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warningf("error reading resp.body:[%s]", err)
		return
	}
	fmt.Println(string(body))
	block := &protos.Block{}
	err = json.Unmarshal(body, block)
	//block, err = protos.UnmarshallBlock(body)
	if err != nil {
		logger.Warningf("Error json.Unmarshal:[%s]", err)
		return
	}
	fmt.Printf("stateHash.type:[%s]\n", reflect.TypeOf(block.StateHash))
	fmt.Printf("stateHash:[%s]\n", string(block.StateHash))
	fmt.Printf("block.String:[%s]\n", block.String())
	txes := block.GetTransactions()
	for _, tx := range txes {
		fmt.Printf("transaction string:[%s]\n", tx.String())
		ccis := &protos.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(tx.Payload, ccis)
		if err != nil {
			fmt.Printf("unmarshal chaincodeinvocationspec err:[%s]", err)
		}
		fmt.Println("unmarshal payload succ")
		cm := ccis.ChaincodeSpec.GetCtorMsg()
		fmt.Printf("cm`type:[%s]\n", reflect.TypeOf(cm))
		fmt.Printf("cm`compactstring:[%s]\n", cm.String())
		fmt.Printf("cm`compactstring:[%s]\n", cm)
		cmJson, err := proto.Marshal(cm)
		if err != nil {
			fmt.Printf("json.marshal CtorMsg err: [%s]", err)
		}
		fmt.Println(cmJson)
	}

}
