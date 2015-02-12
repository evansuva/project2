package main

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"

	"github.com/PointCoin/btcjson"
	"github.com/PointCoin/btcutil"
	"github.com/PointCoin/pointcoind/blockchain"
)

const (
	rpcuser = "dave"        // This match your rpcuser and rpcpass in pointcoind.conf
	rpcpass = "crypto$%bux" // and this too.
	cert    = "/home/ubuntu/.pointcoind/rpc.cert"
)

func main() {
	// Setup the client using application constants, die horribly if there's a problem
	client := setupRpcClient(cert, rpcuser, rpcpass)

	// Declare variables to use in our main loop
	var template *btcjson.GetBlockTemplateResult
	var difficulty big.Int

	var hashCounter int
	var err error

	var prevHash string
	var height int64

	for { // Loop forever
		// Get a new block template from pointcoind.
		log.Printf("Requesting a block template\n")
		template, err = client.GetBlockTemplate(&btcjson.TemplateRequest{})
		if err != nil {
			log.Fatal(err)
		}

		// The template returned by GetBlockTemplate provides these fields
		// that you will need to use to create a new block:

		// hash of the previous block
		prevHash = template.PreviousHash

		// difficulty target
		difficulty = formatDiff(template.Bits)

		// height of the next block (number of blocks between genesis block and next block)
		height = template.Height

		txs := formatTransactions(template.Transactions) // returns the transactions from the network
		msg := "Your computing ID"
		a := "PsVSrUSQf72X6GWFQXJPxR7WSAPVRb1gWx"
		coinbaseTx := CreateCoinbaseTx(height, a, msg) // address conversion moved into CreateCoinbaseTx

		txs = prepend(coinbaseTx.MsgTx(), txs)

		merkleRoot := createMerkleRoot(txs)

		nonce := rand.Uint32()
		block := CreateBlock(prevHash, merkleRoot, difficulty, nonce, txs)

		for attempts := 0; attempts < 10000; attempts++ {
			// Hash the header (BlockSha defined in btcwire/blockheader.go)
			hash, _ := block.Header.BlockSha()
			hashCounter += 1
			if lessThanDiff(hash, difficulty) {
				// Success! Send the whole block
				log.Printf("Found good nonce [%d], attempt: [%d]\n", block.Header.Nonce, attempts)
				// We use a btcutil block b/c SubmitBlock demands it.
				err := client.SubmitBlock(btcutil.NewBlock(block), nil)
				if err != nil {
					errStr := fmt.Sprintf("Block Submission to node failed with: %s\n", err)
					log.Println(errStr)
					break
				}

				log.Printf("Block Submitted! Hash: [%s] as big [%s]\n",
					hash, blockchain.ShaHashToBig(&hash).String())
				break
			}

			// Increment the nonce in the block's header. It might overflow, but that's
			// no big deal.
			block.Header.Nonce += 1
			// log.Printf("Trying nonce: %d\n", block.Header.Nonce)
		}
	}
}
