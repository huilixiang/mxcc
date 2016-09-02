package main

import (
	"encoding/json"
	"errors"
	"fmt"
	//"reflect"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/golang/protobuf/proto"
	//"github.com/hyperledger/fabric/core/util"
	//"github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
)

var shareTxPrefix = "USTX_"
var downloadTxPrefix = "UDTX_"
var payerPrefix = "PAYER_"
var payeePrefix = "PAYEE_"

var rebateLogger = logging.MustGetLogger("rebate")

//type BonusType int

const (
	FirstFour = iota
	LastFour
)

const (
	ShareAct = iota
	DownloadAct
)

type RebateChaincode struct {
}

type ActionInfo struct {
	PrevTx    string
	Uid       string
	ActType   int
	BonusType int
	App       string
}

//
func (r *RebateChaincode) Init(stub shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {
	return nil, nil
}

func (r *RebateChaincode) Invoke(stub shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {
	ccStub := stub.(*shim.ChaincodeStub)
	if "payerDeposit" == function {
		_, err := r.payerDeposit(ccStub, args)
		if err != nil {
			return nil, err
		}
	} else if "share" == function {
		ret, err := r.share(ccStub, args)
		rebateLogger.Debugf("share ret:[%s], err:[%s]", ret, err)
	} else if "download" == function {
		ret, err := r.download(ccStub, args)
		rebateLogger.Debugf("download ret:[%s], err:[%s]", ret, err)
	}
	return nil, nil
}

func (r *RebateChaincode) Delete(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

func (r *RebateChaincode) Query(stub shim.ChaincodeStubInterface, function string,
	args []string) ([]byte, error) {
	ccStub := (stub).(*shim.ChaincodeStub)
	if "payerBalance" == function {
		payerID := args[0]
		balance, err := r.getPayerBalance(ccStub, payerID)
		if err != nil {
			return nil, err
		}
		rebateLogger.Warningf("payer:[%s]`s balance:[%d]", payerID, balance)
		return []byte(strconv.Itoa(balance)), nil
	} else if "payeeBalance" == function {

	}

	return nil, nil
}

func (r *RebateChaincode) download(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	return nil, nil
}

func (r *RebateChaincode) share(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Exception 5")
	}
	uid := args[0]
	actType, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("Error parsing actType:[%s]", args[1])
	}
	app := args[2]
	bonusType, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, fmt.Errorf("Error parsing bonusType:[%s]", args[3])
	}
	prevTXID := args[4]
	actionInfo := ActionInfo{Uid: uid, PrevTx: prevTXID, BonusType: bonusType, App: app, ActType: actType}
	curTXID := stub.UUID
	stateKey := shareTxPrefix + curTXID
	jsonStr, err := json.Marshal(actionInfo)
	if err != nil {
		return nil, fmt.Errorf("Json.Marshal actionInfo error")
	}
	err = stub.PutState(stateKey, []byte(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("stub.PutState error: [%s]", err)
	}
	rebateLogger.Debugf("share success, json:[%s]", jsonStr)
	return nil, nil
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
	balance, err := r.getPayerBalance(stub, payerID)
	if err != nil {
		return nil, err
	}
	balance += amount
	err = stub.PutState(payerPrefix+payerID, []byte(strconv.Itoa(balance)))
	if err != nil {
		warning := fmt.Sprintf("payer:[%s] deposit failed, err:[%s]", payerID, err)
		rebateLogger.Warningf(warning)
		return nil, errors.New(warning)
	}
	return []byte(strconv.Itoa(balance)), nil
}

func (r *RebateChaincode) rebate(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	//下载动作的transaction id. 根据相应的返利策略查找分享链的上各种人物，然后计算分红
	var err error
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Excepting 3")
	}
	downloadTxID := args[0]
	bonusType, err := strconv.Atoi(args[1])
	payerID := args[2]
	var payees []string
	switch bonusType {
	case FirstFour:
		//分享链上的前四个人
		break
	case LastFour:
		//分享链上的最后四个人
		payees, err = r.findLastN(stub, downloadTxID, 4)
		break
	default:
		return nil, errors.New(fmt.Sprintf("there is no matched BonusType:[%d]", bonusType))
	}
	balance, err := r.getPayerBalance(stub, payerID)
	if err != nil {
		return nil, err
	}
	rebateMoney := 10
	for _, payee := range payees {
		rebateLogger.Warningf("rebate user:[%s], money:[%d]", payee, 10)
		//给用户转帐
		payeeKey := payeePrefix + payee
		payeeBalanceBytes, err := stub.GetState(payeeKey)
		if err != nil {
			rebateLogger.Warningf("getting payee:[%s]`s balance err:[%s]", payee, err)
			continue
		}
		payeeBalance := 0
		if payeeBalanceBytes != nil {
			payeeBalance, err = strconv.Atoi(string(payeeBalanceBytes))
		}
		balance = balance - rebateMoney
		payeeBalance = payeeBalance + rebateMoney
		err = stub.PutState(payeeKey, []byte(strconv.Itoa(payeeBalance)))
		if err != nil {
			rebateLogger.Warningf("payer:[%s] rebate:[%d] to payee:[%s] err:[%s]", payerID, rebateMoney, payee, err)
			return nil, err
		}
	}
	err = stub.PutState(payerPrefix+payerID, []byte(strconv.Itoa(balance)))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *RebateChaincode) getPayerBalance(stub *shim.ChaincodeStub, payerID string) (int, error) {

	var payerKey = payerPrefix + payerID
	balanceBytes, err := stub.GetState(payerKey)
	if err != nil {
		warning := fmt.Sprintf("GetState failed for payer:[%s] deposit, err:[%s]", payerID, err)
		rebateLogger.Warningf(warning)
		return -1, errors.New(warning)
	}
	balance := 0
	if balanceBytes == nil {
		rebateLogger.Warningf("payer [%s] not exists. it`s the first time to deposit for him/her", payerID)
	} else {
		balance, err = strconv.Atoi(string(balanceBytes))
		if err != nil {
			warning := fmt.Sprintf("payer:[%s]`s balance[%s] is abnormal", payerID, string(balanceBytes))
			rebateLogger.Warningf(warning)
			return -1, errors.New(warning)
		}
	}
	return balance, nil
}

func (r *RebateChaincode) findTopN(stub *shim.ChaincodeStub, rightTxID string, n int) (payees []string, err error) {
	return nil, nil
}

func (r *RebateChaincode) findLastN(stub *shim.ChaincodeStub, rightTxID string, n int) ([]string, error) {
	var payees = make([]string, n)
	key := downloadTxPrefix + rightTxID
	for i := 0; i < n; {
		stateBytes, err := stub.GetState(key)
		if err != nil {
			rebateLogger.Warningf("GetState by key:[%s] err:[%s]", key, err)
			continue
		}
		//stateStr := string(stateBytes)
		var ai ActionInfo
		err = json.Unmarshal(stateBytes, &ai)
		if err != nil {
			rebateLogger.Warningf("json unmarshal state error")
			continue
		}
		prevTx := ai.PrevTx
		uid := ai.Uid
		payees[i] = uid
		key = shareTxPrefix + prevTx
		i++
	}
	return payees, nil
}

func main() {
	err := shim.Start(new(RebateChaincode))
	if err != nil {
		rebateLogger.Warningf("Error starting RebateChaincode: [%s]", err)
	}
}
