package blsbftv2

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

type NodeInterface interface {
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	//GetMiningKeys() string
	//GetPrivateKey() string
	//GetUserMiningState() (role string, chainID int)
	RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error)
	GetSelfPeerID() peer.ID
}

type ChainInterface interface {
	GetFinalView() multiview.View
	GetBestView() multiview.View
	//GetEpoch() uint64
	GetChainName() string
	//GetConsensusType() string
	//GetLastBlockTimeStamp() int64
	//GetMinBlkInterval() time.Duration
	//GetMaxBlkCreateTime() time.Duration
	IsReady() bool
	//SetReady(bool)
	//GetActiveShardNumber() int
	//CurrentHeight() uint64
	//GetCommitteeSize() int
	//GetCommittee() []incognitokey.CommitteePublicKey
	//GetPendingCommittee() []incognitokey.CommitteePublicKey
	//GetPubKeyCommitteeIndex(string) int
	//GetLastProposerIndex() int
	UnmarshalBlock(blockString []byte) (types.BlockInterface, error)
	CreateNewBlock(version int, proposer string, round int, startTime int64) (types.BlockInterface, error)
	CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64) (types.BlockInterface, error)
	InsertAndBroadcastBlock(block types.BlockInterface) error
	// ValidateAndInsertBlock(block types.BlockInterface) error
	//ValidateBlockSignatures(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	ValidatePreSignBlock(block types.BlockInterface) error
	GetShardID() int

	//for new syncker
	//GetBestViewHeight() uint64
	//GetFinalViewHeight() uint64
	//GetBestViewHash() string
	//GetFinalViewHash() string

	GetViewByHash(hash common.Hash) multiview.View
}
