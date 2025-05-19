package main

import (
	"blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()
	defer func() {
		if err := bc.DB().Close(); err != nil {
			panic(err)
		}
	}()

	cli := &CLI{bc}
	cli.Run()
}
