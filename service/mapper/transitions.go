// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package mapper

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/complete/mtrie/trie"
	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/models/dps"
)

// TransitionFunc is a function that is applied onto the state machine's
// state.
type TransitionFunc func(*State) error

// Transitions is what applies transitions to the state of an FSM.
type Transitions struct {
	cfg   Config
	log   zerolog.Logger
	chain dps.Chain
	feed  Feeder
	index dps.Writer
	once  *sync.Once
}

// NewTransitions returns a Transitions component using the given dependencies and using the given options
func NewTransitions(log zerolog.Logger, chain dps.Chain, feed Feeder, index dps.Writer, options ...Option) *Transitions {

	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}

	t := Transitions{
		log:   log.With().Str("component", "mapper_transitions").Logger(),
		cfg:   cfg,
		chain: chain,
		feed:  feed,
		index: index,
		once:  &sync.Once{},
	}

	return &t
}

// InitializeMapper initializes the mapper and determines whether we resume from
// a previous indexing operation, or whether we bootstrap the state.
func (t *Transitions) InitializeMapper(s *State) error {
	if s.status != StatusInitialize {
		return fmt.Errorf("invalid states for initializing mapper (%s)", s.status)
	}

	// For now, we are assuming that we need to bootstrap if the root checkpoint
	// is given, and that we need to resume if no root checkpoint is given.
	if t.cfg.RootTrie != nil {
		t.log.Info().Msg("root trie found, bootstrapping index")
		s.status = StatusBootstrap
		return nil
	}

	t.log.Info().Msg("no root trie found, resuming indexing")

	s.status = StatusResume
	return nil
}

// BootstrapState bootstraps the state by loading the checkpoint if there is one
// and initializing the elements subsequently used by the FSM.
func (t *Transitions) BootstrapState(s *State) error {
	if s.status != StatusBootstrap {
		return fmt.Errorf("invalid status for bootstrapping state (%s)", s.status)
	}

	// If we are bootstrapping, we need the root trie to extract the ledger
	// registers that need to be indexed for the root block.
	if t.cfg.RootTrie == nil {
		return fmt.Errorf("no root trie available for bootstrapping index")
	}
	tree := t.cfg.RootTrie

	// We always need at least one step in our forest, which is used as the
	// stopping point when indexing the payloads since the last finalized
	// block. We thus introduce an empty tree, with no paths and an
	// irrelevant previous commit.
	empty := trie.NewEmptyMTrie()
	s.forest.Save(empty, nil, flow.DummyStateCommitment)

	// The chain indexing will forward last to next and next to current height,
	// which will be the one for the checkpoint.
	first := flow.StateCommitment(empty.RootHash())
	s.last = flow.DummyStateCommitment
	s.next = first

	t.log.Info().Hex("commit", first[:]).Msg("added empty tree to forest")

	// Then, we can load the root height and apply it to the state. That
	// will allow us to load the root blockchain data in the next step.
	height, err := t.chain.Root()
	if err != nil {
		return fmt.Errorf("could not get root height: %w", err)
	}
	s.height = height

	// Here, we store all the paths so we can index the payloads, if wanted.
	paths := allPaths(tree)
	s.forest.Save(tree, paths, first)

	second := tree.RootHash()
	t.log.Info().Uint64("height", s.height).Hex("commit", second[:]).Int("registers", len(paths)).Msg("added checkpoint tree to forest")

	// We have successfully bootstrapped. However, no chain data for the root
	// block has been indexed yet. This is why we "pretend" that we just
	// forwarded the state to this height, so we go straight to the chain data
	// indexing.
	s.status = StatusUpdate
	return nil
}

// ResumeIndexing resumes indexing the data from a previous run.
func (t *Transitions) ResumeIndexing(s *State) error {
	if s.status != StatusResume {
		return fmt.Errorf("invalid status for resuming indexing (%s)", s.status)
	}

	return nil
}

// IndexChain indexes chain data for the current height.
func (t *Transitions) IndexChain(s *State) error {
	if s.status != StatusIndex {
		return fmt.Errorf("invalid status for indexing chain (%s)", s.status)
	}

	log := t.log.With().Uint64("height", s.height).Logger()

	// We try to retrieve the next header until it becomes available, which
	// means all data coming from the protocol state is available after this
	// point.
	header, err := t.chain.Header(s.height)
	if errors.Is(err, dps.ErrUnavailable) {
		log.Debug().Msg("waiting for next header")
		time.Sleep(t.cfg.WaitInterval)
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not get header: %w", err)
	}

	// At this point, we can retrieve the data from the consensus state. This is
	// a slight optimization for the live indexer, as it allows us to process
	// some data before the full execution data becomes available.
	guarantees, err := t.chain.Guarantees(s.height)
	if err != nil {
		return fmt.Errorf("could not get guarantees: %w", err)
	}
	seals, err := t.chain.Seals(s.height)
	if err != nil {
		return fmt.Errorf("could not get seals: %w", err)
	}

	// We can also proceed to already indexing the data related to the consensus
	// state, before dealing with anything related to execution data, which
	// might go into the wait state.
	blockID := header.ID()
	err = t.index.Height(blockID, s.height)
	if err != nil {
		return fmt.Errorf("could not index height: %w", err)
	}
	err = t.index.Header(s.height, header)
	if err != nil {
		return fmt.Errorf("could not index header: %w", err)
	}
	err = t.index.Guarantees(s.height, guarantees)
	if err != nil {
		return fmt.Errorf("could not index guarantees: %w", err)
	}
	err = t.index.Seals(s.height, seals)
	if err != nil {
		return fmt.Errorf("could not index seals: %w", err)
	}

	// Next, we try to retrieve the next commit until it becomes available,
	// at which point all the data coming from the execution data should be
	// available.
	commit, err := t.chain.Commit(s.height)
	if errors.Is(err, dps.ErrUnavailable) {
		log.Debug().Msg("waiting for next state commitment")
		time.Sleep(t.cfg.WaitInterval)
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not get commit: %w", err)
	}
	collections, err := t.chain.Collections(s.height)
	if err != nil {
		return fmt.Errorf("could not get collections: %w", err)
	}
	transactions, err := t.chain.Transactions(s.height)
	if err != nil {
		return fmt.Errorf("could not get transactions: %w", err)
	}
	results, err := t.chain.Results(s.height)
	if err != nil {
		return fmt.Errorf("could not get transaction results: %w", err)
	}
	events, err := t.chain.Events(s.height)
	if err != nil {
		return fmt.Errorf("could not get events: %w", err)
	}

	// Next, all we need to do is index the remaining data and we have fully
	// processed indexing for this block height.
	err = t.index.Commit(s.height, commit)
	if err != nil {
		return fmt.Errorf("could not index commit: %w", err)
	}
	err = t.index.Collections(s.height, collections)
	if err != nil {
		return fmt.Errorf("could not index collections: %w", err)
	}
	err = t.index.Transactions(s.height, transactions)
	if err != nil {
		return fmt.Errorf("could not index transactions: %w", err)
	}
	err = t.index.Results(results)
	if err != nil {
		return fmt.Errorf("could not index transaction results: %w", err)
	}
	err = t.index.Events(s.height, events)
	if err != nil {
		return fmt.Errorf("could not index events: %w", err)
	}

	// At this point, we need to forward the the `last` state commitment to
	// `next`, so we know what the state commitment was at the last finalized
	// block we processed. This will allow us to know when to stop when
	// walking back through the forest to collect trie updates.
	s.last = s.next

	// Last but not least, we need to update `next` to point to the commit we
	// have just retrieved for the new block height. This is the sentinel that
	// tells us when we have collected enough trie updates for the forest to
	// have reached the next finalized block.
	s.next = commit

	log.Info().Msg("indexed blockchain data for finalized block")

	// After indexing the blockchain data, we can go back to updating the state
	// tree until we find the commit of the finalized block. This will allow us
	// to index the payloads then.
	s.status = StatusUpdate
	return nil
}

// UpdateTree updates the state's tree. If the state's forest already matches with the next block's state commitment,
// it immediately returns and sets the state's status to StatusMatched.
func (t *Transitions) UpdateTree(s *State) error {
	if s.status != StatusUpdate {
		return fmt.Errorf("invalid status for updating tree (%s)", s.status)
	}

	log := t.log.With().Uint64("height", s.height).Hex("last", s.last[:]).Hex("next", s.next[:]).Logger()

	// If the forest contains a tree for the commit of the next finalized block,
	// we have reached our goal, and we can go to the next step in order to
	// collect the register payloads we want to index for that block.
	ok := s.forest.Has(s.next)
	if ok {
		log.Info().Hex("commit", s.next[:]).Msg("matched commit of finalized block")
		s.status = StatusCollect
		return nil
	}

	// First, we get the next tree update from the feeder. We can skip it if
	// it doesn't have any updated paths, or if we can't find the tree to apply
	// it to in the forest. This usually means that it was meant for a pruned
	// branch of the execution forest.
	update, err := t.feed.Update()
	if errors.Is(err, dps.ErrUnavailable) {
		time.Sleep(t.cfg.WaitInterval)
		log.Debug().Msg("waiting for next trie update")
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not feed update: %w", err)
	}
	parent := flow.StateCommitment(update.RootHash)
	tree, ok := s.forest.Tree(parent)
	if !ok {
		log.Warn().Msg("state commitment mismatch, retrieving next trie update")
		return nil
	}

	// We then apply the update to the relevant tree, as retrieved from the
	// forest, and save the updated tree in the forest. If the tree is not new,
	// we should error, as that should not happen.
	paths, payloads := pathsPayloads(update)
	tree, err = trie.NewTrieWithUpdatedRegisters(tree, paths, payloads)
	if err != nil {
		return fmt.Errorf("could not update tree: %w", err)
	}
	s.forest.Save(tree, paths, parent)

	hash := tree.RootHash()
	log.Info().Hex("commit", hash[:]).Int("registers", len(paths)).Msg("updated tree with register payloads")

	return nil
}

// CollectRegisters reads the payloads for the next block to be indexed from the state's forest, unless payload
// indexing is disabled.
func (t *Transitions) CollectRegisters(s *State) error {
	log := t.log.With().Uint64("height", s.height).Hex("commit", s.next[:]).Logger()
	if s.status != StatusCollect {
		return fmt.Errorf("invalid status for collecting registers (%s)", s.status)
	}

	// If indexing payloads is disabled, we can bypass collection and indexing
	// of payloads and just go straight to forwarding the height to the next
	// finalized block.
	if t.cfg.SkipRegisters {
		s.status = StatusForward
		return nil
	}

	// If we index payloads, we are basically stepping back from (and including)
	// the tree that corresponds to the next finalized block all the way up to
	// (and excluding) the tree for the last finalized block we indexed. To do
	// so, we will use the parent state commit to retrieve the parent trees from
	// the forest, and we use the paths we recorded changes on to retrieve the
	// changed payloads at each step.
	commit := s.next
	for commit != s.last {

		// We do this check only once, so that we don't need to do it for
		// each item we retrieve. The tree should always be there, but we
		// should check just to not fail silently.
		ok := s.forest.Has(commit)
		if !ok {
			return fmt.Errorf("could not load tree (commit: %x)", commit)
		}

		// For each path, we retrieve the payload and add it to the registers we
		// will index later. If we already have a payload for the path, it is
		// more recent as we iterate backwards in time, so we can skip the
		// outdated payload.
		// NOTE: We read from the tree one by one here, as the performance
		// overhead is minimal compared to the disk i/o for badger, and it
		// allows us to ignore sorting of paths.
		tree, _ := s.forest.Tree(commit)
		paths, _ := s.forest.Paths(commit)
		for _, path := range paths {
			_, ok := s.registers[path]
			if ok {
				continue
			}
			payloads := tree.UnsafeRead([]ledger.Path{path})
			s.registers[path] = payloads[0]
		}

		log.Debug().Int("batch", len(paths)).Msg("collected register batch for finalized block")

		// We now step back to the parent of the current state trie.
		parent, _ := s.forest.Parent(commit)
		commit = parent
	}

	log.Info().Int("registers", len(s.registers)).Msg("collected all registers for finalized block")

	// At this point, we have collected all the payloads, so we go to the next
	// step, where we will index them.
	s.status = StatusMap
	return nil
}

// IndexRegisters maps the collected registers to the current block.
func (t *Transitions) MapRegisters(s *State) error {
	if s.status != StatusMap {
		return fmt.Errorf("invalid status for indexing registers (%s)", s.status)
	}

	log := t.log.With().Uint64("height", s.height).Hex("commit", s.next[:]).Logger()

	// If there are no registers left to be indexed, we can go to the next step,
	// which is about forwarding the height to the next finalized block.
	if len(s.registers) == 0 {
		log.Info().Msg("indexed all registers for finalized block")
		s.status = StatusForward
		return nil
	}

	// We will now collect and index 1000 registers at a time. This gives the
	// FSM the chance to exit the loop between every 1000 payloads we index. It
	// doesn't really matter for badger if they are in random order, so this
	// way of iterating should be fine.
	n := 1000
	paths := make([]ledger.Path, 0, n)
	payloads := make([]*ledger.Payload, 0, n)
	for path, payload := range s.registers {
		paths = append(paths, path)
		payloads = append(payloads, payload)
		delete(s.registers, path)
		if len(paths) >= n {
			break
		}
	}

	// Then we store the (maximum) 1000 paths and payloads.
	err := t.index.Payloads(s.height, paths, payloads)
	if err != nil {
		return fmt.Errorf("could not index registers: %w", err)
	}

	log.Debug().Int("batch", len(paths)).Int("remaining", len(s.registers)).Msg("indexed register batch for finalized block")

	return nil
}

// ForwardHeight increments the height at which the mapping operates, and updates the last indexed height.
func (t *Transitions) ForwardHeight(s *State) error {
	if s.status != StatusForward {
		return fmt.Errorf("invalid status for forwarding height (%s)", s.status)
	}

	// After finishing the indexing of the payloads for a finalized block, or
	// skipping it, we should document the last indexed height. On the first
	// pass, we will also index the first indexed height here.
	var err error
	t.once.Do(func() { err = t.index.First(s.height) })
	if err != nil {
		return fmt.Errorf("could not index first height: %w", err)
	}
	err = t.index.Last(s.height)
	if err != nil {
		return fmt.Errorf("could not index last height: %w", err)
	}

	// Now that we have indexed the heights, we can forward to the next height,
	// and reset the forest to free up memory.
	s.height++
	s.forest.Reset(s.next)

	t.log.Info().Uint64("height", s.height).Msg("forwarded finalized block to next height")

	// Once the height is forwarded, we can set the status so that we index
	// the blockchain data next.
	s.status = StatusIndex
	return nil
}
