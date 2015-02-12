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
		difficulty = convertDifficulty(template.Bits) //[ convertDifficulty returns template.Bits in some useful form - could be bigint ] 
		
		// height of the next block (number of blocks between genesis block and next block)
		height = template.Height

		
		block, err = createBlock(prevHash, difficulty, height)
		if err != nil {
			log.Fatal(err)
		}
		
		txs = getNetworkTransactions(???) // returns the transactions from the network
		a := "PsVSrUSQf72X6GWFQXJPxR7WSAPVRb1gWx"
		coinbaseTx, err := CreateCoinbaseTx(height, a, msg) // address conversion moved into CreateCoinbaseTx
		if err != nil {
			return nil, err
		}
		
		txs.insert(coinbaseTx) // probably not valid go, but whatever is needed
		
		// we'll provide hints for this in the instructions
		store := blockchain.BuildMerkleTreeStore(txs)
		// Create a merkleroot from a list of 1 transaction.
		merkleRoot := store[len(store)-1]
		nonce := rand.Uint32()
		block = CreateBlock(prevHash, merkleRoot, difficulty, nonce)

		for attempts := 0; attempts < 10000; attempts++ {
			// Hash the header (BlockSha defined in btcwire/blockheader.go)
			hash := block.Header.BlockSha()
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
			
			// Increment the nonce in the block's header. It might overflow, but that's
			// no big deal.
			block.Header.Nonce += 1
			// log.Printf("Trying nonce: %d\n", block.Header.Nonce)
		}
	}
}

