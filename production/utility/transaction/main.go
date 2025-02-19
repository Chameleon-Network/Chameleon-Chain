// chain/production/utility/transaction/main.go
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver1"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	// Specify the correct path to the LevelDB data folder
	dbPath := filepath.Join("..", "..", "..", "productiondata")
	db, err := incdb.Open("leveldb", dbPath)
	if err != nil {
		fmt.Println("Could not open connection to LevelDB:", err)
		panic(err)
	}

	// Proceed with initializing transactions
	initGenesisTx(db)
	//initThankTx(db)
}

func initGenesisTx(db incdb.Database) {
	var initTxs []string
	testUserkeyList := map[string]uint64{
		"112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5JqxDAX6sjMoUT6": uint64(5000000000000000),
	}

	for privateKey, initAmount := range testUserkeyList {
		// Deserialize private key and initialize the user's keyset
		testUserKey, _ := wallet.Base58CheckDeserialize(privateKey)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := tx_ver1.Tx{}

		// Create a StateDB instance
		stateDB, err := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
		if err != nil {
			fmt.Println("Error creating StateDB instance:", err)
			return
		}

		// Initialize the transaction with the necessary parameters
		err = testSalaryTX.InitTxSalary(
			initAmount,
			&testUserKey.KeySet.PaymentAddress,
			&testUserKey.KeySet.PrivateKey,
			stateDB,
			nil, // metadata is not needed for now
		)
		if err != nil {
			fmt.Println("Error initializing transaction:", err)
			continue
		}

		// Marshal the transaction into JSON for printing
		initTx, _ := json.MarshalIndent(testSalaryTX, "", "  ")
		initTxs = append(initTxs, string(initTx))
	}

	// Print the initialized transactions
	fmt.Println(initTxs)
}

func initThankTx(db incdb.Database) {
	var initTxs []string
	testUserkeyList := map[string]string{
		"112t8rnXBS7jJ4iqFon5rM66ex1Fc7sstNrJA9iMKgNURMUf3rywYfJ4c5Kpxw1BgL1frj9Nu5uL5vpemn9mLUW25CD1w7khX88WdauTVyKa": "@abc",
	}

	for privateKey, info := range testUserkeyList {
		testUserKey, _ := wallet.Base58CheckDeserialize(privateKey)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := tx_ver1.Tx{}

		// Update InitTxSalary call to match the required parameters
		err := testSalaryTX.InitTxSalary(
			0,
			&testUserKey.KeySet.PaymentAddress,
			&testUserKey.KeySet.PrivateKey,
			nil, // Replace with actual StateDB if required
			nil, // Replace with actual Metadata if required
		)
		if err != nil {
			fmt.Println("Error initializing transaction:", err)
			continue
		}

		testSalaryTX.Info = []byte(info)
		initTx, _ := json.MarshalIndent(testSalaryTX, "", "  ")
		initTxs = append(initTxs, string(initTx))
	}
	fmt.Println(initTxs)
}
