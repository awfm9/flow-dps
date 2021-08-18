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

package feeder

import (
	"fmt"

	pwal "github.com/prometheus/tsdb/wal"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/complete/wal"

	"github.com/optakt/flow-dps/models/dps"
)

type Disk struct {
	reader *pwal.Reader
}

// FromDisk creates a trie update feeder that sources state deltas
// directly from an execution node's trie directory.
func FromDisk(reader *pwal.Reader) (*Disk, error) {

	l := Disk{
		reader: reader,
	}

	return &l, nil
}

func (d *Disk) Update() (*ledger.TrieUpdate, error) {

	// We read in a loop because the WAL contains entries that are not trie
	// updates; we don't really need to care about them, so we can just skip
	// them until we find a trie update.
	for {

		// This part reads the next entry from the WAL, makes sure we didn't
		// encounter an error when reading or decoding and ensures that it's a
		// trie update.
		next := d.reader.Next()
		err := d.reader.Err()
		if !next && err != nil {
			return nil, fmt.Errorf("could not read next record: %w", err)
		}
		if !next {
			return nil, dps.ErrUnavailable
		}
		record := d.reader.Record()
		operation, _, update, err := wal.Decode(record)
		if err != nil {
			return nil, fmt.Errorf("could not decode record: %w", err)
		}
		if operation != wal.WALUpdate {
			continue
		}

		// For older versions, we need to verify the length of types that are aliased
		// to the hash.Hash type from Flow Go, because it is a slice instead of
		// a fixed-length byte array.
		if len(update.RootHash) != 32 {
			return nil, fmt.Errorf("invalid ledger root hash length in trie update: got %d want 32", len(update.RootHash))
		}
		for _, path := range update.Paths {
			if len(path) != 32 {
				return nil, fmt.Errorf("invalid ledger path length in trie update: got %d want 32", len(path))
			}
		}

		// However, we need to make sure that all slices are copied, because the
		// decode function will reuse the underlying slices later.
		update = clone(update)

		return update, nil
	}
}
