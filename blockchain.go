package blockchain

import (
	"errors"
	"log"

	"github.com/boltdb/bolt"
)

const (
	dbFile           = "blockchain.db"
	blockChainBucket = "blocksBucket"
	lastBucketKey    = "l"
)

// Blockchain 区块链链表
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func (bc *Blockchain) AddBlock(data string) error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockChainBucket))
		if b == nil {
			err := errors.New("block chain not found")
			return err
		}

		newBlock := NewBlock(data, bc.tip)
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
		return err
	}

	return nil
}

func (bc *Blockchain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.tip, bc.db}
}

func (bc *Blockchain) DB() *bolt.DB {
	return bc.db
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Printf("open db error: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockChainBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, _ = tx.CreateBucket([]byte(blockChainBucket))
			_ = b.Put(genesis.Hash, genesis.Serialize())
			_ = b.Put([]byte(lastBucketKey), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte(lastBucketKey))
		}
		return nil
	})
	if err != nil {
		log.Printf("update db error: %v", err)
	}

	bc := &Blockchain{tip, db}

	return bc
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
