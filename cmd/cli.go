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
	case "send":
		sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
		from := sendCmd.String("from", "", "Source address")
		to := sendCmd.String("to", "", "Destination address")
		amount := sendCmd.Int("amount", 0, "Amount to send")
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			println("Error parsing send command")
			os.Exit(1)
		}
		if sendCmd.Parsed() {
			if *from == "" || *to == "" || *amount <= 0 {
				println("From, to and amount are required")
				os.Exit(1)
			}
			cli.send(*from, *to, *amount)
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
	case "createblockchain":
		createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
		address := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			println("Error parsing createblockchain command")
			os.Exit(1)
		}
		if createBlockchainCmd.Parsed() {
			if *address == "" {
				println("Address is required")
				os.Exit(1)
			}
			cli.createBlockchain(*address)
		}
	case "getbalance":
		getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
		address := getBalanceCmd.String("address", "", "The address to get balance for")
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			println("Error parsing getbalance command")
			os.Exit(1)
		}
		if getBalanceCmd.Parsed() {
			if *address == "" {
				println("Address is required")
				os.Exit(1)
			}
			cli.getBalance(*address)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := blockchain.CreateBlockchain(address)
	_ = bc.DB().Close()
	fmt.Println("create blockchain done!")
}

func (cli *CLI) getBalance(address string) {
	bc := blockchain.NewBlockchain(address)
	defer func() {
		_ = bc.DB().Close()
	}()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// send 发送交易
func (cli *CLI) send(from, to string, amount int) {
	bc := blockchain.NewBlockchain(from)
	defer func() {
		_ = bc.DB().Close()
	}()

	transaction := blockchain.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*blockchain.Transaction{transaction})
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	bc := blockchain.NewBlockchain("")
	it := bc.Iterator()
	for block := it.Next(); block != nil; block = it.Next() {
		pow := blockchain.NewProofOfWork(block)

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Trsansations: %s\n", block.HashTransactions())
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
	fmt.Println("Usage: cli <command>")
	fmt.Println("Commands:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}
