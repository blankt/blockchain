package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

const subsidy = 1

// Transaction 一次交易包含的值
type Transaction struct {
	// 交易id
	ID []byte
	// 这笔交易关联的多个上一笔交易的输入
	Vin []TXInput
	//  这笔交易关联的输出
	VOut []TXOutput
}

func (t *Transaction) SetID() {
	var hash [32]byte
	txCopy := t
	txCopy.ID = nil

	// 计算交易的hash值
	data, _ := json.Marshal(t)
	hash = sha256.Sum256(data)
	t.ID = hash[:]
}

type TXInput struct {
	// 关联的输出的交易的id
	TxId []byte
	// 关联的输出交易id的下标
	VOut int
	// 解锁输出交易的脚本
	ScriptSig string
}

type TXOutput struct {
	// 输出交易的值
	Value int
	// 锁定输出交易的脚本
	ScriptPubKey string
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}
