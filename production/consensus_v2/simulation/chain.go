// chain/production/consensus_v2/simulation/chain.go
package main

import (
	"encoding/json"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
)

type Chain struct {
	multiview multiview.MultiView
	chainID   int
	chainName string
}

func NewChain(chainID int, chainName string, committee []incognitokey.CommitteePublicKey) *Chain {
	c := new(Chain)
	c.chainID = chainID
	c.chainName = chainName
	c.multiview = multiview.NewMultiView()
	state := &State{
		NewBlock(1, 1, "Genesis", common.Hash{}),
		committee,
	}
	c.multiview.AddView(state)
	return c
}

func NewBlock(height int, round int, proposer string, hash common.Hash) types.BlockInterface {
	return &types.ShardBlock{
		Header: types.ShardHeader{
			Height:   uint64(height),
			Round:    round,
			
			Proposer: proposer,
		},
	}
}

func (c *Chain) GetFinalView() multiview.View {
	return c.multiview.GetFinalView()
}

func (c *Chain) GetBestView() multiview.View {
	return c.multiview.GetBestView()
}

func (c *Chain) GetChainName() string {
	return c.chainName
}

func (s *Chain) IsReady() bool {
	return true
}

func (s *Chain) UnmarshalBlock(blockString []byte) (types.BlockInterface, error) {
	blk := &types.ShardBlock{}
	json.Unmarshal(blockString, blk)
	return blk, nil
}

func (c *Chain) CreateNewBlock(version int, proposer string, round int, startTime int64) (types.BlockInterface, error) {
	newBlock := NewBlock(int(c.GetBestView().GetHeight()+1), int(time.Now().Unix()), proposer, *c.GetBestView().GetHash())
	return newBlock, nil
}

func (c *Chain) CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64) (types.BlockInterface, error) {
	//TODO: must using the old block data, and timestamp
	oldBlock.(*types.ShardBlock).Header.Proposer = proposer
	oldBlock.(*types.ShardBlock).Header.ProposeTime = time.Now().Unix()
	// newBlock := NewBlock(c.GetBestView().GetHeight()+1, oldBlock.GetProduceTime(), proposer, *c.GetBestView().GetHash())
	return oldBlock, nil
}

func (s *Chain) InsertAndBroadcastBlock(block types.BlockInterface) error {
	state := &State{
		block,
		s.multiview.GetBestView().GetCommittee(),
	}
	s.multiview.AddView(state)
	return nil
}

func (s *Chain) ValidatePreSignBlock(block types.BlockInterface) error {
	return nil
}

func (s *Chain) GetShardID() int {
	return s.chainID
}

func (c Chain) GetViewByHash(hash common.Hash) multiview.View {
	return c.multiview.GetViewByHash(hash)
}
