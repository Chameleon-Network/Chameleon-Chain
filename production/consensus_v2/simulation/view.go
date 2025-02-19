// chain/production/consensus_v2/simulation/view.go
package main

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type State struct {
	block     types.BlockInterface
	committee []incognitokey.CommitteePublicKey
}

func (s *State) CalculateTimeSlot(timestamp int64) int64 {
	// Implement the logic for calculating the time slot
	// based on the timestamp
	return timestamp
}

// GetHash returns the hash of the block in the state.
func (s *State) GetHash() *common.Hash {
	// Ensure that the block is not nil
	if s.block == nil {
		return nil
	}
	return s.block.Hash()
}

// GetPreviousHash returns the hash of the previous block.
func (s *State) GetPreviousHash() *common.Hash {
	// Ensure that the block is not nil
	if s.block == nil {
		return nil
	}
	hash := s.block.GetPrevHash()
	return &hash
}

// GetHeight returns the height of the block.
func (s *State) GetHeight() uint64 {
	// Ensure that the block is not nil
	if s.block == nil {
		return 0
	}
	return s.block.GetHeight()
}

// GetCommittee returns the list of committee public keys.
func (s *State) GetCommittee() []incognitokey.CommitteePublicKey {
	// Return the committee; can be an empty slice if not set
	return s.committee
}

// GetProposerByTimeSlot returns the proposer for a specific time slot and version.
func (s *State) GetProposerByTimeSlot(ts int64, version int) (incognitokey.CommitteePublicKey, int) {
	// Check if the committee is empty to avoid out-of-bounds access
	if len(s.committee) == 0 {
		return incognitokey.CommitteePublicKey{}, -1 // Return an empty key and an invalid index
	}

	id := blockchain.GetProposerByTimeSlot(ts, len(s.committee))
	// Ensure that the id is within bounds
	if id < 0 || id >= len(s.committee) {
		return incognitokey.CommitteePublicKey{}, -1 // Invalid index
	}
	return s.committee[id], id
}

// GetBlock returns the block in the state.
func (s *State) GetBlock() types.BlockInterface {
	return s.block
}
