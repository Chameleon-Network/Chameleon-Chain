package consensus_v2

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

type EngineConfig struct {
	Node          NodeInterface
	Blockchain    *blockchain.BlockChain
	PubSubManager *pubsub.PubSubManager
}

type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	IsEnableMining() bool
	GetMiningKeys() string
	GetPrivateKey() string
	GetUserMiningState() (role string, chainID int)
	GetPubkeyMiningState(*incognitokey.CommitteePublicKey) (role string, chainID int)
	RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error)
	GetSelfPeerID() peer.ID
	PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error
	IsBeaconFullnode(*incognitokey.CommitteePublicKey) (bool, string)
}

type ConsensusInterface interface {
	// GetConsensusName - retrieve consensus name
	GetConsensusName() string
	GetChainKey() string
	GetChainID() int

	// Start - start consensus
	Start() error
	// Stop - stop consensus
	Stop() error
	// IsOngoing - check whether consensus is currently voting on a block
	IsOngoing() bool
	IsStarted() bool
	// ProcessBFTMsg - process incoming BFT message
	ProcessBFTMsg(msg *wire.MessageBFT)
	// ValidateProducerSig - validate a block producer signature
	//ValidateProducerSig(block common.BlockInterface) error
	// ValidateCommitteeSig - validate a block committee signature
	//ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error

	// LoadUserKey - load user mining key
	LoadUserKeys(miningKey []signatureschemes2.MiningKey) error
	// GetUserPublicKey - get user public key of loaded mining key
	GetUserPublicKey() *incognitokey.CommitteePublicKey
	// ValidateData - validate data with this consensus signature scheme
	ValidateData(data []byte, sig string, publicKey string) error
	// SignData - sign data with this consensus signature scheme
	SignData(data []byte) (string, error)
}
