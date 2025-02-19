// chain/production/incognito.go
package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"path/filepath"
	"runtime/debug"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"github.com/incognitochain/incognito-chain/databasemp"
	_ "github.com/incognitochain/incognito-chain/databasemp/lvdb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/pruner"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/wallet"
)

// handleError handles errors gracefully and provides detailed logs with stack traces.
func handleError(err error, message string) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
		debug.PrintStack()
		panic(err)
	}
}

// recoverPanic recovers from panics, logs stack traces, and prevents crashes.
func recoverPanic() {
	if r := recover(); r != nil {
		log.Printf("[PANIC] Recovered: %v", r)
		debug.PrintStack()
	}
}

func getBTCRelayingChain(btcRelayingChainID, btcDataFolderName string) (*btcrelaying.BlockChain, error) {
    defer recoverPanic()

    log.Printf("[DEBUG] Initializing BTC relaying chain: ID=%s, Folder=%s", btcRelayingChainID, btcDataFolderName)
    relayingChainParams := map[string]*chaincfg.Params{
        portal.TestnetBTCChainID:  btcrelaying.GetTestNet3Params(),
        portal.Testnet2BTCChainID: btcrelaying.GetTestNet3ParamsForInc2(),
        portal.MainnetBTCChainID:  btcrelaying.GetMainNetParams(),
    }
    relayingChainGenesisBlkHeight := map[string]int32{
        portal.TestnetBTCChainID:  int32(2063133),
        portal.Testnet2BTCChainID: int32(2064989),
        portal.MainnetBTCChainID:  int32(697298),
    }
    chain, err := btcrelaying.GetChainV2(
        filepath.Join(config.Config().DataDir, btcDataFolderName),
        relayingChainParams[btcRelayingChainID],
        relayingChainGenesisBlkHeight[btcRelayingChainID],
    )
    if err != nil {
        log.Printf("[ERROR] Failed to initialize BTC relaying chain: %v", err)
        return nil, err
    }
    if chain == nil {
        log.Printf("[ERROR] BTC relaying chain is nil")
        return nil, fmt.Errorf("BTC relaying chain is nil")
    }
    return chain, nil
}
func getBNBRelayingChainState(bnbRelayingChainID string) (*bnbrelaying.BNBChainState, error) {
	defer recoverPanic()

	log.Printf("[DEBUG] Loading BNB relaying chain state: ID=%s", bnbRelayingChainID)
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join(config.Config().DataDir, "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to load BNB relaying chain state: %v", err)
		return nil, err
	}
	if bnbChainState == nil {
		log.Printf("[ERROR] BNB relaying chain state is nil")
		return nil, fmt.Errorf("BNB relaying chain state is nil")
	}
	return bnbChainState, nil
}

func mainMaster(serverChan chan<- *Server) error {
	defer recoverPanic()

	cfg := config.LoadConfig()
	log.Printf("[DEBUG] Config loaded: %+v", cfg)

	initLogRotator(cfg.LogFileName)

	if err := parseAndSetDebugLevels(cfg.LogLevel); err != nil {
		handleError(err, "Invalid log levels")
	}

	config.LoadParam()
	portal.SetupParam()

	err := wallet.InitPublicKeyBurningAddressByte()
	handleError(err, "Wallet initialization failed")

	blockchain.CreateGenesisBlocks()

	interrupt := interruptListener()

	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join(cfg.DataDir, cfg.DatabaseDir))
	handleError(err, "Database initialization failed")

	p := pruner.NewPrunerManager(db)
	if config.Config().OfflinePrune {
		p.OfflinePrune()
		return nil
	}

	consensusDB, err := incdb.Open("leveldb", filepath.Join(cfg.DataDir, "consensus"))
	handleError(err, "Consensus database initialization failed")
	rawdb_consensus.SetConsensusDatabase(consensusDB)

	dbmp, err := databasemp.Open("leveldbmempool", filepath.Join(cfg.DataDir, cfg.MempoolDir))
	handleError(err, "Mempool database initialization failed")

	btcChain, err := getBTCRelayingChain(
		portal.GetPortalParams().RelayingParam.BTCRelayingHeaderChainID,
		portal.GetPortalParams().RelayingParam.BTCDataFolderName,
	)
	handleError(err, "BTC relaying chain setup failed")

	bnbChainState, err := getBNBRelayingChainState(portal.GetPortalParams().RelayingParam.BNBRelayingHeaderChainID)
	handleError(err, "BNB relaying chain state setup failed")

	server := Server{}
	err = server.NewServer(cfg.Listener, db, dbmp, nil, cfg.NumIndexerWorkers, cfg.IndexerAccessTokens, version(), btcChain, bnbChainState, p, interrupt)
	handleError(err, "Server setup failed")
	server.Start()

	<-interrupt
	return nil
}

func main() {
	defer recoverPanic()
	err := mainMaster(nil)
	if err != nil {
		log.Fatalf("[ERROR] Main master failed: %v", err)
	}
}
