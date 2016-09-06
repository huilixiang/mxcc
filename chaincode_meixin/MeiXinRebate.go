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

var shareTXPrefix = "USTX_"
var downloadTXPrefix = "UDTX_"
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
		payeeID := args[0]
		balance, err := r.getPayeeBalance(ccStub, payeeID)
		if err != nil {
			return nil, err
		}
		rebateLogger.Debugf("payee:[%s]`s balance:[%d]", payeeID, balance)
		return []byte(strconv.Itoa(balance)), nil
	} else if "shareTX" == function {
		txID := args[0]
		txBytes, err := stub.GetState(shareTXPrefix + txID)
		if err != nil {
			return nil, err
		}
		rebateLogger.Debugf("share tx:[%s]`s detail:[%s].", txID, string(txBytes))
		return txBytes, nil
	} else if "downloadTX" == function {
		txID := args[0]
		txBytes, err := stub.GetState(downloadTXPrefix + txID)
		if err != nil {
			return nil, err
		}
		rebateLogger.Debugf("download tx:[%s]`s detail:[%s].", txID, string(txBytes))
		return txBytes, nil

	}
	return nil, nil
}

func (r *RebateChaincode) download(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

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
	stateKey := downloadTXPrefix + curTXID
	jsonStr, err := json.Marshal(actionInfo)
	if err != nil {
		return nil, fmt.Errorf("Json.Marshal actionInfo error")
	}
	err = stub.PutState(stateKey, []byte(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("stub.PutState error: [%s]", err)
	}
	rebateLogger.Debugf("share success, json:[%s]", jsonStr)
	//调用返利处理函数。此处只是demo阶段，应该有一条单独的链来做这项工作
	callArgs := make([]string, 3)
	//应该提供app与软件供应商的对应关系，找到支付人，此处简单的模拟
	callArgs[0] = curTXID
	callArgs[1] = fmt.Sprintf("%d", LastFour)
	callArgs[2] = "payer_a"
	_, err = r.rebate(stub, callArgs)
	if err != nil {
		return nil, err
	}
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
	stateKey := shareTXPrefix + curTXID
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
		return nil, fmt.Errorf("payer:[%s] deposit failed, err:[%s]", payerID, err)
	}
	return []byte(strconv.Itoa(balance)), nil
}

func (r *RebateChaincode) rebate(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	//下载动作的transaction id. 根据相应的返利策略查找分享链的上各种人物，然后计算分红
	var err error
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Excepting 3")
	}
	downloadTXID := args[0]
	bonusType, err := strconv.Atoi(args[1])
	payerID := args[2]
	var payees []string
	switch bonusType {
	case FirstFour:
		//分享链上的前四个人
		break
	case LastFour:
		//分享链上的最后四个人
		payees, err = r.findLastN(stub, downloadTXID, 4)
		break
	default:
		return nil, fmt.Errorf("there is no matched BonusType:[%d]", bonusType)
	}
	balance, err := r.getPayerBalance(stub, payerID)
	if err != nil {
		return nil, err
	}
	rebateMoney := 10
	for _, payee := range payees {
		if len(payee) < 1 {
			break
		}
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
			if err != nil {
				return nil, err
			}
		}
		rebateLogger.Warningf("before rebating payee:[%s]`s balance:[%d], money:[%d]", payee, payeeBalance, rebateMoney)
		balance = balance - rebateMoney
		payeeBalance = payeeBalance + rebateMoney
		err = stub.PutState(payeeKey, []byte(strconv.Itoa(payeeBalance)))
		if err != nil {
			rebateLogger.Warningf("payer:[%s] rebate:[%d] to payee:[%s] err:[%s]", payerID, rebateMoney, payee, err)
			return nil, err
		}
		rebateLogger.Warningf("payer:[%s] rebate:[%d] to payee:[%s], balance:[%d]", payerID, rebateMoney, payee, balance)
	}
	err = stub.PutState(payerPrefix+payerID, []byte(strconv.Itoa(balance)))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *RebateChaincode) getPayeeBalance(stub *shim.ChaincodeStub, payeeID string) (int, error) {
	var key = payeePrefix + payeeID
	balanceBytes, err := stub.GetState(key)
	if err != nil {
		return -1, fmt.Errorf("getting payee:[%s]`s balance err:[%s]", payeeID, err)
	}
	balance := 0
	if balanceBytes != nil {
		balance, err = strconv.Atoi(string(balanceBytes))
		if err != nil {
			return -1, fmt.Errorf("parsing payee:[%s]`s balance:[%s] err:[%s]", payeeID, string(balanceBytes), err)
		}
	}
	return balance, nil
}

func (r *RebateChaincode) getPayerBalance(stub *shim.ChaincodeStub, payerID string) (int, error) {

	var payerKey = payerPrefix + payerID
	balanceBytes, err := stub.GetState(payerKey)
	if err != nil {
		return -1, fmt.Errorf("GetState failed for payer:[%s] deposit, err:[%s]", payerID, err)
	}
	balance := 0
	if balanceBytes == nil {
		rebateLogger.Warningf("payer [%s] not exists. it`s the first time to deposit for him/her", payerID)
	} else {
		balance, err = strconv.Atoi(string(balanceBytes))
		if err != nil {
			return -1, fmt.Errorf("payer:[%s]`s balance[%s] is abnormal", payerID, string(balanceBytes))
		}
	}
	return balance, nil
}

func (r *RebateChaincode) findTopN(stub *shim.ChaincodeStub, rightTxID string, n int) (payees []string, err error) {
	return nil, nil
}

func (r *RebateChaincode) findLastN(stub *shim.ChaincodeStub, rightTxID string, n int) ([]string, error) {
	var payees = make([]string, n)
	key := downloadTXPrefix + rightTxID
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
		if len(prevTx) < 1 {
			break
		}
		key = shareTXPrefix + prevTx
		i++
	}
	rebateLogger.Warningf("baneift users:[%v]", payees)
	return payees, nil
}

func main() {
	err := shim.Start(new(RebateChaincode))
	if err != nil {
		rebateLogger.Warningf("Error starting RebateChaincode: [%s]", err)
	}
}
