package blockchain

import (
	"encoding/hex"
	"errors"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	dbFile              = "blockchain.db"
	blockChainBucket    = "blocksBucket"
	lastBucketKey       = "l"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

// Blockchain 区块链链表
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain 创建区块链和存储的数据库
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		log.Println("Blockchain already exists")
		os.Exit(1)
	}

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var genesis *Block
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockChainBucket))
		if b != nil {
			return err
		}

		coinbaseTxn := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis = NewGenesisBlock(coinbaseTxn)
		b, _ = tx.CreateBucket([]byte(blockChainBucket))
		_ = b.Put(genesis.Hash, genesis.Serialize())
		_ = b.Put([]byte(lastBucketKey), genesis.Hash)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := &Blockchain{genesis.Hash, db}
	return bc
}

func NewBlockchain(address string) *Blockchain {
	if !dbExists() {
		log.Println("No existing blockchain found")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Printf("open db error: %v", err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockChainBucket))
		if b == nil {
			log.Println("block chain not found")
			os.Exit(1)
		}
		tip = b.Get([]byte(lastBucketKey))
		if tip == nil {
			log.Println("last bucket not found")
			os.Exit(1)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := &Blockchain{tip, db}
	return bc
}

// MineBlock 挖一个块，并把交易存储进去
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockChainBucket))
		if b == nil {
			log.Panic("block chain not found")
		}

		newBlock := NewBlock(transactions, bc.tip)
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte(lastBucketKey), newBlock.Hash)
		if err != nil {
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindUnspentTransactions 查找我未花费的交易，所有我的收到的钱且未支出的钱。(钱包余额)
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.VOut {
				// Was the output spent? 这笔交易输出是否已经关联到其他交易的输入中，如果已经关联了，表示这笔交易收到的钱已经花出去了。
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// 如果这笔输出能被我解锁，就是这笔交易中我收到的钱（且未支出）
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					// 这笔交易的输入是否能被我解锁，如果能被我解锁，就是我花出去的钱
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxId)
						// 记录我已经花费的交易 这笔交易中的index
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.VOut)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO 找到所有未花费的交易输出，即我所有的余额。
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.VOut {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxns := bc.FindUnspentTransactions(address)
	accumulated := 0

	for _, tx := range unspentTxns {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.VOut {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

func (bc *Blockchain) DB() *bolt.DB {
	return bc.db
}

func (bc *Blockchain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.tip, bc.db}
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

type BlockChainIterator struct {
	current []byte
	db      *bolt.DB
}

func (bc *BlockChainIterator) Next() *Block {
	var block *Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockChainBucket))
		if b == nil {
			return errors.New("block chain not found")
		}

		data := b.Get(bc.current)
		if len(data) == 0 {
			return nil
		}
		block = DeserializeBlock(data)
		return nil
	})
	if err != nil {
		log.Printf("get block error: %v", err)
	}
	if block == nil {
		return nil
	}

	bc.current = block.PrevBlockHash
	return block
}
