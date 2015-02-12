package main

import (
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/PointCoin/btcjson"
	"github.com/PointCoin/btcutil"
	"github.com/PointCoin/btcwire"
	"github.com/PointCoin/pointcoind/blockchain"
)

const (
	rpcuser = "dave" // make this match your rpcuser and rpcpass in pointcoind.conf
	rpcpass = "crypto$%bux"
	cert    = "/home/ubuntu/.pointcoind/rpc.cert"
)

func main() {
	// Setup the client using application constants, die horribly if there's a problem
	client := setupRpcClient(cert, rpcuser, rpcpass)

	// Declare variables to use in our main loop
	var template *btcjson.GetBlockTemplateResult
	var block *btcwire.MsgBlock
	var difficulty uint64

	var hashCounter int
	var err error

	var prevHash string
	var height int64

	for { // Loop forever
		// Get a new block template from pointcoind.
		log.Printf("Updating block template\n")
		template, err = client.GetBlockTemplate(&btcjson.TemplateRequest{})
		if err != nil {
			log.Fatal(err)
		}

		// The template returned by GetBlockTemplate provides these fields
		// that you will need to use to create a new block:

		// hash of the previous block
		prevHash = template.PreviousHash

		// difficulty target
		difficulty, _ = strconv.ParseUint(template.Bits, 16, 32)
		
		// height of the next block (number of blocks between genesis block and next block)
		height = template.Height

		block, err = createBlock(prevHash, difficulty, height)
		if err != nil {
			log.Fatal(err)
		}
		
 		//! difficulty = formatDiff(template.Bits)
		log.Printf("Difficulty: %d\n", difficulty)

		for attempts := 0; attempts < 10000; attempts++ {
			// Increment the nonce in the block's header. It might overflow, but that's
			// no big deal.
			block.Header.Nonce += 1
			// log.Printf("Trying nonce: %d\n", block.Header.Nonce)
			
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
				
				log.Printf("Block Submitted! Hash: [%s] asbig [%s]\n", 
					hash, blockchain.ShaHashToBig(&hash).String())
				break
			}
		}
	}
}

// lessThanDiff returns true if the hash satisifies the target difficulty. That
// is to say if the hash interpreted as a big integer is less than the required
// difficulty then return true otherwise return false.
func lessThanDiff(hash btcwire.ShaHash, difficulty big.Int) bool {
	bigI := blockchain.ShaHashToBig(&hash)
	return bigI.Cmp(&difficulty) <= 0
}

