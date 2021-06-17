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

package storage

import (
	"encoding/binary"
	"fmt"

	"github.com/OneOfOne/xxhash"
	"github.com/dgraph-io/badger/v2"
	"github.com/fxamacker/cbor/v2"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/common/pathfinder"
	"github.com/onflow/flow-go/model/flow"
)

func SaveFirst(height uint64) func(*badger.Txn) error {
	return save(encodeKey(prefixFirst), height)
}

func SaveLast(height uint64) func(*badger.Txn) error {
	return save(encodeKey(prefixLast), height)
}

func SaveCommit(height uint64, commit flow.StateCommitment) func(*badger.Txn) error {
	return save(encodeKey(prefixCommit, height), commit)
}

func SaveHeader(height uint64, header *flow.Header) func(*badger.Txn) error {
	return save(encodeKey(prefixHeader, height), header)
}

func SaveEvents(height uint64, typ flow.EventType, events []flow.Event) func(*badger.Txn) error {
	hash := xxhash.ChecksumString64(string(typ))
	return save(encodeKey(prefixEvents, height, hash), events)
}

func SavePayload(height uint64, path ledger.Path, payload *ledger.Payload) func(*badger.Txn) error {
	return save(encodeKey(prefixPayload, path, height), payload)
}

func SaveHeight(blockID flow.Identifier, height uint64) func(*badger.Txn) error {
	return save(encodeKey(prefixHeight, blockID), height)
}

func SaveTransaction(transaction flow.Transaction) func(*badger.Txn) error {
	return save(encodeKey(prefixTransaction, transaction.ID()), transaction)
}

func SaveTransactions(blockID flow.Identifier, transactions []flow.Identifier) func(*badger.Txn) error {
	return save(encodeKey(prefixTransactions, blockID), transactions)
}

func RetrieveFirst(height *uint64) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixFirst), height)
}

func RetrieveLast(height *uint64) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixLast), height)
}

func RetrieveHeader(height uint64, header *flow.Header) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixHeader, height), header)
}

func RetrieveCommit(height uint64, commit *flow.StateCommitment) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixCommit, height), commit)
}

func RetrieveEvents(height uint64, types []flow.EventType, events *[]flow.Event) func(*badger.Txn) error {
	return func(tx *badger.Txn) error {
		lookup := make(map[uint64]struct{})
		for _, typ := range types {
			hash := xxhash.ChecksumString64(string(typ))
			lookup[hash] = struct{}{}
		}

		prefix := encodeKey(prefixEvents, height)
		opts := badger.DefaultIteratorOptions
		// NOTE: this is an optimization only, it does not enforce that all
		// results in the iteration have this prefix.
		opts.Prefix = prefix

		it := tx.NewIterator(opts)
		defer it.Close()

		// Iterate on all keys with the right prefix.
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// If types were given for filtering, discard events which should not be included.
			hash := binary.BigEndian.Uint64(it.Item().Key()[1+8:])
			_, ok := lookup[hash]
			if len(lookup) != 0 && !ok {
				continue
			}

			// Unmarshal event batch and append them to result slice.
			var evts []flow.Event
			err := it.Item().Value(func(val []byte) error {
				val, err := decompressor.DecodeAll(val, nil)
				if err != nil {
					return fmt.Errorf("could not decompress events: %w", err)
				}
				err = cbor.Unmarshal(val, &evts)
				if err != nil {
					return fmt.Errorf("could not decode events: %w", err)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("could not unmarshal events: %w", err)
			}

			*events = append(*events, evts...)
		}

		return nil
	}
}

func RetrievePayload(height uint64, path ledger.Path, payload *ledger.Payload) func(*badger.Txn) error {
	return func(tx *badger.Txn) error {
		key := encodeKey(prefixPayload, path, height)
		it := tx.NewIterator(badger.IteratorOptions{
			PrefetchSize:   0,
			PrefetchValues: false,
			Reverse:        true,
			AllVersions:    false,
			InternalAccess: false,
			Prefix:         key[:1+pathfinder.PathByteSize],
		})
		defer it.Close()

		it.Seek(key)
		if !it.Valid() {
			return badger.ErrKeyNotFound
		}

		err := it.Item().Value(func(val []byte) error {
			val, err := decompressor.DecodeAll(val, nil)
			if err != nil {
				return fmt.Errorf("could not decompress payload: %w", err)
			}
			err = cbor.Unmarshal(val, &payload)
			if err != nil {
				return fmt.Errorf("could not decode payload: %w", err)
			}
			return nil
		})
		return err
	}
}

func RetrieveHeight(blockID flow.Identifier, height *uint64) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixHeight, blockID), height)
}

func RetrieveTransaction(transactionID flow.Identifier, transaction *flow.Transaction) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixTransaction, transactionID), transaction)
}

func RetrieveTransactions(blockID flow.Identifier, transactions *[]flow.Identifier) func(*badger.Txn) error {
	return retrieve(encodeKey(prefixTransactions, blockID), transactions)
}
