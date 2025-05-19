package main

import (
	"blockchain"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
	bc *blockchain.Blockchain
}

func (cli *CLI) Run() {
	cli.validateArgs()

	switch os.Args[1] {
	case "addblock":
		addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
		addBlockData := addBlockCmd.String("data", "", "Block data")
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			println("Error parsing addblock command")
			os.Exit(1)
		}
		if addBlockCmd.Parsed() {
			if *addBlockData == "" {
				println("Data is required")
				os.Exit(1)
			}
			err = cli.bc.AddBlock(*addBlockData)
			if err != nil {
				println("Error adding block:", err)
				os.Exit(1)
			}
		}
	case "printchain":
		printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			println("Error parsing printchain command")
			os.Exit(1)
		}
		if printChainCmd.Parsed() {
			cli.printChain()
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printChain() {
	it := cli.bc.Iterator()
	for block := it.Next(); block != nil; block = it.Next() {
		pow := blockchain.NewProofOfWork(block)

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	println("Usage: cli <command>")
	println("Commands:")
	println("  addblock -data DATA Add a block about DATA to the blockchain")
	println("  printchain - Print the blockchain")
}
