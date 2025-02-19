// chain/production/multiview/multiview.go
package multiview

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incdb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type View interface {
	CalculateTimeSlot(int64) int64
	GetCurrentTimeSlot() int64
	GetHash() *common.Hash
	GetPreviousHash() *common.Hash
	GetHeight() uint64
	GetCommittee() []incognitokey.CommitteePublicKey
	GetPreviousBlockCommittee(db incdb.Database) ([]incognitokey.CommitteePublicKey, error)
	CommitteeStateVersion() int
	GetBlock() types.BlockInterface
	ReplaceBlock(blk types.BlockInterface)
	GetBeaconHeight() uint64
	GetProposerByTimeSlot(ts int64, version int) (incognitokey.CommitteePublicKey, int)
	GetProposerLength() int
	CompareCommitteeFromBlock(View) int
	PastHalfTimeslot(int64) bool
}

type MultiView interface {
	ReplaceBlockIfImproveFinality(block types.BlockInterface) (bool, error)
	IsInstantFinality() bool
	GetViewByHash(hash common.Hash) View
	SimulateAddView(view View) (cloneMultiview MultiView)
	GetBestView() View
	GetFinalView() View
	FinalizeView(hashToFinalize common.Hash) error
	GetExpectedFinalView() View
	GetAllViewsWithBFS() []View
	RunCleanProcess()
	Clone() MultiView
	Reset()
	AddView(v View) (int, error)
}

type multiView struct {
	viewByHash     map[common.Hash]View   // Mapping from block hash to views
	viewByPrevHash map[common.Hash][]View // Mapping from previous block hash to views
	lock           *sync.RWMutex

	// State tracking views
	finalView         View // View that must not be reverted (finalized)
	expectedFinalView View // Expected final view
	bestView          View // Best view for the current state
}

func NewMultiView() *multiView {
	s := &multiView{
		viewByHash:     make(map[common.Hash]View),
		viewByPrevHash: make(map[common.Hash][]View),
		lock:           new(sync.RWMutex),
	}
	return s
}

// RunCleanProcess runs a periodic cleanup of outdated views.
func (s *multiView) RunCleanProcess() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				if len(s.viewByHash) > 1 {
					s.removeOutdatedView()
				}
			}
		}
	}()
}

// Clone creates a deep copy of the current multi-view instance.
func (s *multiView) Clone() MultiView {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cloneMV := NewMultiView()
	for h, v := range s.viewByHash {
		cloneMV.viewByHash[h] = v
	}
	for h, v := range s.viewByPrevHash {
		cloneMV.viewByPrevHash[h] = v
	}
	cloneMV.expectedFinalView = s.expectedFinalView
	cloneMV.bestView = s.bestView
	cloneMV.finalView = s.finalView
	return cloneMV
}

// Reset clears all the views and resets the state.
func (s *multiView) Reset() {
	s.viewByHash = make(map[common.Hash]View)
	s.viewByPrevHash = make(map[common.Hash][]View)
	s.expectedFinalView = nil
	s.bestView = nil
	s.finalView = nil
}

// removeOutdatedView cleans up views that are outdated based on the final view height.
func (s *multiView) removeOutdatedView() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for h, v := range s.viewByHash {
		if v.GetHeight() < s.finalView.GetHeight()-1 {
			delete(s.viewByHash, h)
			delete(s.viewByPrevHash, h)
			delete(s.viewByPrevHash, *v.GetPreviousHash())
		}
	}
}

// GetViewByHash retrieves a view by its hash if it exists and is still valid.
func (s *multiView) GetViewByHash(hash common.Hash) View {
	s.lock.RLock()
	defer s.lock.RUnlock()
	view, _ := s.viewByHash[hash]
	if view == nil || view.GetHeight() < s.finalView.GetHeight() {
		return nil
	}
	return view
}

// FinalizeView finalizes the view corresponding to the given hash.
func (s *multiView) FinalizeView(hashToFinalize common.Hash) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	viewToFinalize, ok := s.viewByHash[hashToFinalize]
	if !ok {
		return errors.New("Cannot find view by hash " + hashToFinalize.String())
	}
	if s.finalView.GetHeight() >= viewToFinalize.GetHeight() {
		return nil
	}

	// Ensure the view to finalize is on the same branch as the best view.
	notLink := true
	prevView := s.bestView
	for {
		if hashToFinalize.String() == prevView.GetHash().String() {
			notLink = false
			break
		}
		prevView = s.viewByHash[*prevView.GetPreviousHash()]
		if prevView == nil {
			break
		}
	}

	if notLink {
		// Reorganize multi-view if not linked to the current chain
		newOrgMultiView := s.getAllViewsWithBFS(viewToFinalize)
		s.Reset()
		for _, view := range newOrgMultiView {
			s.updateViewState(view)
		}
	} else {
		// If the final view is older, update it to the new one
		if s.finalView.GetHeight() < viewToFinalize.GetHeight() {
			s.finalView = viewToFinalize
		}
	}
	return nil
}

// addView adds a view to the multi-view if it's validated and properly linked.
func (s *multiView) addView(view View) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Initialize the multi-view if empty
	if len(s.viewByHash) == 0 {
		s.viewByHash[*view.GetHash()] = view
		s.updateViewState(view)
		return true
	}

	// Add new view only if it links to a previous valid view
	if _, exists := s.viewByHash[*view.GetHash()]; !exists {
		if _, exists := s.viewByHash[*view.GetPreviousHash()]; exists {
			s.viewByHash[*view.GetHash()] = view
			s.viewByPrevHash[*view.GetPreviousHash()] = append(s.viewByPrevHash[*view.GetPreviousHash()], view)
			s.updateViewState(view)
			return true
		}
	}
	return false
}

// GetBestView retrieves the best view (most advanced view).
func (s *multiView) GetBestView() View {
	return s.bestView
}

// GetFinalView retrieves the final view that cannot be reverted.
func (s *multiView) GetFinalView() View {
	return s.finalView
}

// GetExpectedFinalView retrieves the expected final view at this point.
func (s *multiView) GetExpectedFinalView() View {
	return s.expectedFinalView
}

// updateViewState updates the state of the multi-view system when a new view is added.
func (s *multiView) updateViewState(newView View) {
	if s.expectedFinalView == nil {
		s.finalView = newView
		s.bestView = newView
		s.expectedFinalView = newView
		return
	}

	// Update the best view based on the block version and other criteria
	if newView.GetBlock().GetVersion() < types.INSTANT_FINALITY_VERSION {
		if newView.GetHeight() > s.bestView.GetHeight() {
			s.bestView = newView
		}

		if newView.GetHeight() == s.bestView.GetHeight() {
			if newView.GetBlock().GetProduceTime() < s.bestView.GetBlock().GetProduceTime() {
				s.bestView = newView
			}
		}

		// Update the expected final view based on the version
		if newView.GetBlock().GetVersion() == types.BFT_VERSION {
			prev1Hash := s.bestView.GetPreviousHash()
			if prev1Hash == nil {
				return
			}
			prev1View := s.viewByHash[*prev1Hash]
			if prev1View == nil {
				return
			}
			s.expectedFinalView = prev1View
		} else if newView.GetBlock().GetVersion() >= types.MULTI_VIEW_VERSION {
			prev1Hash := s.bestView.GetPreviousHash()
			prev1View := s.viewByHash[*prev1Hash]
			if prev1View == nil || s.expectedFinalView.GetHeight() == prev1View.GetHeight() {
				return
			}
			bestViewTimeSlot := s.bestView.CalculateTimeSlot(s.bestView.GetBlock().GetProposeTime())
			prev1TimeSlot := s.bestView.CalculateTimeSlot(prev1View.GetBlock().GetProposeTime())
			if prev1TimeSlot+1 == bestViewTimeSlot {
				s.expectedFinalView = prev1View
			}
			if newView.GetBlock().GetVersion() >= types.LEMMA2_VERSION {
				if newView.GetBlock().GetHeight()-1 == newView.GetBlock().GetFinalityHeight() {
					s.expectedFinalView = prev1View
				}
			}
		} else {
			fmt.Println("Block version is not correct")
		}
	}

	// Handle instant finality version updates
	if newView.GetBlock().GetVersion() >= types.INSTANT_FINALITY_VERSION {
		compareCommittee := newView.CompareCommitteeFromBlock(s.bestView)
		if compareCommittee == 0 || newView.GetPreviousHash().String() == s.bestView.GetHash().String() {
			if newView.GetHeight() > s.bestView.GetHeight() {
				s.bestView = newView
			}
			if newView.GetHeight() == s.bestView.GetHeight() {
				if newView.GetBlock().GetProduceTime() < s.bestView.GetBlock().GetProduceTime() {
					s.bestView = newView
				}
			}
		} else {
			newViewCreateNewFinal := false
			bestViewCreateNewFinal := false

			expectedFinalViewOfNewView := s.findExpectFinalView(newView)
			if s.finalView != expectedFinalViewOfNewView {
				newViewCreateNewFinal = true
			}

			if s.finalView != s.expectedFinalView {
				bestViewCreateNewFinal = true
			}

			if !newViewCreateNewFinal && !bestViewCreateNewFinal {
				if compareCommittee == 1 {
					s.bestView = newView
				}
			} else if newViewCreateNewFinal && bestViewCreateNewFinal {
				if compareCommittee == 1 {
					s.bestView = newView
				}
			} else if newViewCreateNewFinal || bestViewCreateNewFinal {
				if newViewCreateNewFinal {
					s.bestView = newView
				}
			}
		}
		s.expectedFinalView = s.findExpectFinalView(s.bestView)
	}
}

// findExpectFinalView traverses backward from the given view to find the expected final view.
func (s *multiView) findExpectFinalView(checkView View) View {
	currentViewPoint := checkView
	for {
		prev1Hash := currentViewPoint.GetPreviousHash()
		prev1View := s.viewByHash[*prev1Hash]

		if prev1View == nil {
			return currentViewPoint
		}

		if currentViewPoint.GetBlock().GetFinalityHeight() != 0 {
			break
		}
		currentViewPoint = prev1View
	}
	return currentViewPoint
}

// IsInstantFinality returns whether the system is in an instant finality state.
func (s *multiView) IsInstantFinality() bool {
	return s.expectedFinalView == s.bestView
}

// getAllViewsWithBFS retrieves all views in the system using BFS starting from the root view.
func (s *multiView) getAllViewsWithBFS(rootView View) []View {
	queue := []View{rootView}

	res := []View{}
	for len(queue) > 0 {
		firstItem := queue[0]
		if firstItem == nil {
			break
		}
		for _, v := range s.viewByPrevHash[*firstItem.GetHash()] {
			queue = append(queue, v)
		}
		res = append(res, firstItem)
		queue = queue[1:]
	}
	return res
}

// GetAllViewsWithBFS returns all views using BFS starting from the final view.
func (s *multiView) GetAllViewsWithBFS() []View {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.getAllViewsWithBFS(s.finalView)
}

// AddView adds a view to the system. This function should not be used directly.
func (s *multiView) AddView(v View) (int, error) {
	panic("must not use this")
}

// isFinalized checks whether the view corresponding to the given hash is finalized.
func (s *multiView) isFinalized(h common.Hash) bool {
	view, ok := s.viewByHash[h]
	if !ok {
		return false
	}

	view = s.expectedFinalView
	for {
		if view.GetHash().String() == h.String() {
			return true
		}
		previousViewHash := view.GetPreviousHash()
		view = s.viewByHash[*previousViewHash]
		if view == nil {
			return false
		}
	}
}

// ReplaceBlockIfImproveFinality checks if a block improves finality and replaces it if necessary.
func (s *multiView) ReplaceBlockIfImproveFinality(b types.BlockInterface) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if b.GetVersion() < types.INSTANT_FINALITY_VERSION {
		return false, nil
	}

	h := *b.Hash()
	replaceView, ok := s.viewByHash[h]
	if !ok {
		return false, errors.New("Cannot find view")
	}

	isImprove := false

	// Check if block improves finality with smaller propose time
	if s.isFinalized(h) && b.GetFinalityHeight() != 0 && b.GetProposeTime() < replaceView.GetBlock().GetProposeTime() {
		log.Println("improve finality")
		s.viewByHash[h] = replaceView
		isImprove = true
	}

	return isImprove, nil
}

func (s *multiView) SimulateAddView(view View) (cloneMultiview MultiView) {
	// Create a clone of the current multi-view
	cloneMV := s.Clone()

	// Add the new view to the cloned multi-view
	success := cloneMV.(*multiView).addView(view)
	if !success {
		return nil
	}

	// Return the cloned multi-view after adding the new view
	return cloneMV
}
