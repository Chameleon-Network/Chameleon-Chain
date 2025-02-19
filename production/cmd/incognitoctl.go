package main

import (
	"log"
	"os"

	"github.com/incognitochain/incognito-chain/common"
)

var (
	cfg *params
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "0.0.1")

	// load component
	tcfg, err := loadParams()
	if err != nil {
		log.Println("Parse component error", err.Error())
		return
	}
	cfg = tcfg

	log.Printf("Process cmd: %s", cfg.Command)
	if ok, err := common.SliceExists(CmdList, cfg.Command); ok || err == nil {
		processCmd()
	} else {
		log.Println("Parse component error", err.Error())
		os.Exit(-1)
	}
}
