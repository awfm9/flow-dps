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

package index

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/models/dps"
)

// Writer implements the `index.Writer` interface to write indexing data to
// an underlying Badger database.
type Writer struct {
	sync.RWMutex
	db   *badger.DB
	lib  dps.WriteLibrary
	cfg  Config
	tx   *badger.Txn
	sema *semaphore.Weighted
	err  chan error
}

// NewWriter creates a new index writer that writes new indexing data to the
// given Badger database.
func NewWriter(db *badger.DB, lib dps.WriteLibrary, options ...func(*Config)) *Writer {

	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}

	w := Writer{
		db:   db,
		lib:  lib,
		cfg:  cfg,
		tx:   db.NewTransaction(true),
		sema: semaphore.NewWeighted(int64(cfg.ConcurrentTransactions)),
		err:  make(chan error, cfg.ConcurrentTransactions),
	}

	return &w
}

// First indexes the height of the first finalized block.
func (w *Writer) First(height uint64) error {
	return w.apply(w.lib.SaveFirst(height))
}

// Last indexes the height of the last finalized block.
func (w *Writer) Last(height uint64) error {
	return w.apply(w.lib.SaveLast(height))
}

// Height indexes the height for the given block ID.
func (w *Writer) Height(blockID flow.Identifier, height uint64) error {
	return w.apply(w.lib.IndexHeightForBlock(blockID, height))
}

// Commit indexes the given commitment of the execution state as it was after
// the execution of the finalized block at the given height.
func (w *Writer) Commit(height uint64, commit flow.StateCommitment) error {
	return w.apply(w.lib.SaveCommit(height, commit))
}

// Header indexes the given header of a finalized block at the given height.
func (w *Writer) Header(height uint64, header *flow.Header) error {
	return w.apply(w.lib.SaveHeader(height, header))
}

// Payloads indexes the given payloads, which should represent a trie update
// of the execution state contained within the finalized block at the given
// height.
func (w *Writer) Payloads(height uint64, paths []ledger.Path, payloads []*ledger.Payload) error {
	if len(paths) != len(payloads) {
		return fmt.Errorf("mismatch between paths and payloads counts")
	}
	return w.apply(func(tx *badger.Txn) error {
		for i, path := range paths {
			payload := payloads[i]
			err := w.lib.SavePayload(height, path, payload)(tx)
			if err != nil {
				return fmt.Errorf("could not save payload (path: %x): %w", path, err)
			}
		}
		return nil
	})
}

func (w *Writer) Collections(height uint64, collections []*flow.LightCollection) error {
	var collIDs []flow.Identifier
	return w.apply(func(tx *badger.Txn) error {
		for _, collection := range collections {
			err := w.lib.SaveCollection(collection)(tx)
			if err != nil {
				return fmt.Errorf("could not store collection (coll: %x): %w", collection.ID(), err)
			}
			collID := collection.ID()
			err = w.lib.IndexTransactionsForCollection(collID, collection.Transactions)(tx)
			if err != nil {
				return fmt.Errorf("could not index transactions for collection (coll: %x): %w", collID, err)
			}
			collIDs = append(collIDs, collID)
		}
		err := w.lib.IndexCollectionsForHeight(height, collIDs)(tx)
		if err != nil {
			return fmt.Errorf("could not index collections for height: %w", err)
		}
		return nil
	})
}

func (w *Writer) Guarantees(_ uint64, guarantees []*flow.CollectionGuarantee) error {
	return w.apply(func(tx *badger.Txn) error {
		for _, guarantee := range guarantees {
			err := w.lib.SaveGuarantee(guarantee)(tx)
			if err != nil {
				return fmt.Errorf("could not store guarantee (tx: %x): %w", guarantee.ID(), err)
			}
		}
		return nil
	})
}

func (w *Writer) Transactions(height uint64, transactions []*flow.TransactionBody) error {
	var txIDs []flow.Identifier
	return w.apply(func(tx *badger.Txn) error {
		for _, transaction := range transactions {
			txID := transaction.ID()
			err := w.lib.SaveTransaction(transaction)(tx)
			if err != nil {
				return fmt.Errorf("could not save transaction (tx: %x): %w", txID, err)
			}
			err = w.lib.IndexHeightForTransaction(txID, height)(tx)
			if err != nil {
				return fmt.Errorf("could not save transaction height (tx: %x): %w", txID, err)
			}
			txIDs = append(txIDs, txID)
		}

		err := w.lib.IndexTransactionsForHeight(height, txIDs)(tx)
		if err != nil {
			return fmt.Errorf("could not index transactions for height: %w", err)
		}

		return nil
	})
}

func (w *Writer) Results(results []*flow.TransactionResult) error {
	return w.apply(func(tx *badger.Txn) error {
		for _, result := range results {
			err := w.lib.SaveResult(result)(tx)
			if err != nil {
				return fmt.Errorf("could not index transaction result (tx: %x): %w", result.TransactionID, err)
			}
		}
		return nil
	})
}

// Events indexes the events, which should represent all events of the finalized
// block at the given height.
func (w *Writer) Events(height uint64, events []flow.Event) error {
	buckets := make(map[flow.EventType][]flow.Event)
	for _, event := range events {
		buckets[event.Type] = append(buckets[event.Type], event)
	}
	err := w.apply(func(tx *badger.Txn) error {
		for typ, evts := range buckets {
			err := w.lib.SaveEvents(height, typ, evts)(tx)
			if err != nil {
				return fmt.Errorf("could not persist events bucket (type: %s): %w", typ, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not index events: %w", err)
	}
	return nil
}

// Seals indexes the seals, which should represent all seals in the finalized
// block at the given height.
func (w *Writer) Seals(height uint64, seals []*flow.Seal) error {
	sealIDs := make([]flow.Identifier, 0, len(seals))
	return w.apply(func(tx *badger.Txn) error {
		for _, seal := range seals {
			err := w.lib.SaveSeal(seal)(tx)
			if err != nil {
				return fmt.Errorf("could not save seal (seal: %x): %w", seal.ID(), err)
			}

			sealIDs = append(sealIDs, seal.ID())
		}

		err := w.lib.IndexSealsForHeight(height, sealIDs)(tx)
		if err != nil {
			return fmt.Errorf("could not index seals for height: %w", err)
		}

		return nil
	})
}

func (w *Writer) apply(op func(*badger.Txn) error) error {

	// Before applying an additional operation to the transaction we are
	// currently building, we want to see if there was an error committing any
	// previous transaction.
	select {
	case err := <-w.err:
		return fmt.Errorf("could not commit transaction: %w", err)
	default:
		// skip
	}

	// If we had no error in a previous transaction, we try applying the
	// operation to the current transaction. If the transaction is already too
	// big, we simply commit it with our callback and start a new transaction.
	// Transaction creation is guarded by a semaphore that limits it to the
	// configured number of inflight transactions.
	err := op(w.tx)
	if errors.Is(err, badger.ErrTxnTooBig) {
		w.tx.CommitWith(w.done)
		_ = w.sema.Acquire(context.Background(), 1)
		w.tx = w.db.NewTransaction(true)
		err = op(w.tx)
	}
	if err != nil {
		return fmt.Errorf("could not apply operation: %w", err)
	}

	return nil
}

func (w *Writer) done(err error) {

	// When a transaction is fully committed, we get the result in this
	// callback. In case of an error, we pipe it to the apply function through
	// the error channel.
	if err != nil {
		w.err <- err
	}

	// Releasing one resource on the semaphore will free up one slot for
	// inflight transactions.
	w.sema.Release(1)
}

func (w *Writer) Close() error {

	// The first transaction we created did not claim a slot on the semaphore.
	// This makes sense, because we only want to limit in-flight (committing)
	// transactions. The currently building transaction is not in-progress.
	// However, we still need to make sure that the currently building
	// transaction is properly committed. We assume that we are no longer
	// appyling new operations when we call `Close`, so we can explicitly do so
	// here, without using the callback.
	err := w.tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit final transaction: %w", err)
	}

	// Once we acquire all semaphore resources, it means all transactions have
	// been committed. We can now close the error channel and drain any
	// remaining errors.
	_ = w.sema.Acquire(context.Background(), int64(w.cfg.ConcurrentTransactions))
	close(w.err)
	var merr *multierror.Error
	for err := range w.err {
		merr = multierror.Append(merr, err)
	}

	return merr.ErrorOrNil()
}
