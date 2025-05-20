package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

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
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(t)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	t.ID = hash[:]
}

func (t *Transaction) IsCoinbase() bool {
	// 判断是否是创世交易
	return len(t.Vin) == 1 && len(t.VOut) == 1 && t.Vin[0].TxId == nil && t.Vin[0].VOut == -1
}

type TXInput struct {
	// 关联的输出的交易的id
	TxId []byte
	// 关联的输出交易id的下标
	VOut int
	// 解锁输出交易的脚本
	ScriptSig string
}

// CanUnlockOutputWith 判断这笔输入交易的钱的钱是否来自于我
func (txIn *TXInput) CanUnlockOutputWith(address string) bool {
	return txIn.ScriptSig == address
}

type TXOutput struct {
	// 输出交易的值
	Value int
	// 锁定输出交易的脚本
	ScriptPubKey string
}

// CanBeUnlockedWith 判断这笔输出交易的钱是否能被我解锁,是否是给我的钱。
func (txOut *TXOutput) CanBeUnlockedWith(address string) bool {
	return txOut.ScriptPubKey == address
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

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var (
		inputs  []TXInput
		outputs []TXOutput
	)

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("not enough funds")
	}

	for transactionId, outs := range validOutputs {
		txID, err := hex.DecodeString(transactionId)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	transaction := &Transaction{nil, inputs, outputs}
	transaction.SetID()

	return transaction
}
