package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"
	"github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
)

var shareTxPrefix = "USTX_"
var downloadTxPrefix = "UDTX_"
var payerPrefix = "PAYER_"
var payeePrefix = "PAYEE_"

var rebateLogger = logging.MustGetLogger("rebate")

type BonusType int

const (
	FirstFour BonusType = iota
	LastFour
)

type RebzoateChaincode struct {
}

type ActionInfo struct {
	prevTx    string
	uid       string
	actType   uint
	bonusType uint
	app       string
}

//
func (r *RebateChaincode) Init(stub *shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {

}

func (r *RebateChaincode) Invoke(stub *shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {

}

func (r *RebateChaincode) Delete(stub *shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {

}

func (r *RebateChaincode) Query(stub *shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {

}

func (r *RebateChaincode) payerDeposit(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	payerID := strings.TrimSpace(args[0])
	if len(payerID) < 1 {
		return nil, errors.New("Invalid arguments. payerID cannot be empty")
	}
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Invalid arguments. deposit amount must be integer")
	}
	if amount < 1 {
		return nil, errors.New("Invalid arguments. deposit amount must be positive	integer")
	}
	rebateLogger.Debugf("payer:[%s] deposit:[%d] invoked", payerID, amount)
	var payerKey = payerPrefix + payerID
	balanceBytes, err := stub.GetState(payerKey)
	if err != nil {
		warning := fmt.Sprintf("GetState failed for payer:[%s] deposit, err:[%s]", payerID, err)
		rebateLogger.Warningf(warning)
		return nil, errors.New(warning)
	}
	balance := 0
	if balanceBytes == nil {
		rebateLogger.Warningf("payer [%s] not exists. it`s the first time to deposit for him/her")
	} else {
		balance, err := strconv.Atoi(string(balanceBytes))
		if err != nil {
			warning := fmt.Sprintf("payer:[%s]`s balance[%s] is abnormal", payerID, string(balanceBytes))
			rebateLogger.Warningf(warning)
			return nil, errors.New(warning)
		}
	}
	balance += amount
	err = stub.PutState(payerKey, []byte(strconv.Itoa(balance)))
	if err != nil {
		warning := fmt.Sprintf("payer:[%s] deposit failed, err:[%s]", payerID, err)
		rebateLogger.Warningf(warning)
		return nil, errors.New(warning)
	}
	return nil, nil
}

func (r *RebateChaincode) rebate(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	//下载动作的transaction id. 根据相应的返利策略查找分享链的上各种人物，然后计算分红
	var err error
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Excepting 2")
	}
	downloadTxId := args[0]
	bonusType := strconv.Atoi(args[1])
	switch bonusType {
	case FirstFour:
		//分享链上的前四个人
		break
	case LastFour:
		//分享链上的最后四个人
		break
	default:
		return nil, errors.New("there is no matched BonusType:" + bonusType)
	}
}

func (r *RebateChaincode) findFirstN(stub *shim.ChaincodeStub, rightTxId string, n uint) (payees []string, err error) {
	return nil, nil
}

func (r *RebateChaincode) findLastN(stub *shim.ChaincodeStub, rightTxId string, n uint) ([]string, error) {
	var payees [n]string
	key = downloadTxPrefix + rightTxId
	for i := 0; i < n; {
		stateBytes, err = stub.GetState(key)
		if err != nil {
			rebateLogger.Warningf("GetState by key:[%s] err:[%s]", key, err)
			continue
		}

	}

	return nil, nil
}
