package blockchain

import "fmt"

func IntToHex(n int64) []byte {
	return []byte(fmt.Sprintf("%x", n))
}
