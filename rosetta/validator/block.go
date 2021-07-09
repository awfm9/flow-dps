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

package validator

import (
	"fmt"

	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/rosetta/failure"
	"github.com/optakt/flow-dps/rosetta/identifier"
)

// Block identifier tries to extrapolate the block identifier to a full version
// of itself. For now, we will always need a height.
// NOTE: We always pass a block identifier that in principle at least could be
// valid, so we will have at least a height or a hash.
func (v *Validator) Block(block identifier.Block) (identifier.Block, error) {

	// We currently only support retrieval by height, until we start indexing
	// the block IDs as part of the DPS index.
	if block.Index == nil {
		return identifier.Block{}, fmt.Errorf("block retrieval by hash only is currently not supported")
	}

	// If a block hash is present, it should be a valid block ID for Flow.
	if block.Hash != "" {
		_, err := flow.HexStringToIdentifier(block.Hash)
		if err != nil {
			return identifier.Block{}, failure.InvalidBlock{
				Index:       *block.Index,
				Hash:        block.Hash,
				Description: failure.NewDescription("block hash is not a valid hex-encoded string"),
			}
		}
	}

	// The block index can't be below the first indexed height.
	first, err := v.index.First()
	if err != nil {
		return identifier.Block{}, fmt.Errorf("could not get first: %w", err)
	}
	if *block.Index < first {
		return identifier.Block{}, failure.InvalidBlock{
			Index: *block.Index,
			Hash:  block.Hash,
			Description: failure.NewDescription("block index is below first indexed height",
				failure.WithUint64("first_index", first),
			),
		}
	}

	// The block index can't be above the last indexed height.
	last, err := v.index.Last()
	if err != nil {
		return identifier.Block{}, fmt.Errorf("could not get last: %w", err)
	}
	if *block.Index > last {
		return identifier.Block{}, failure.UnknownBlock{
			Index: *block.Index,
			Hash:  block.Hash,
			Description: failure.NewDescription("block index is above last indexed height",
				failure.WithUint64("last_index", last),
			),
		}
	}

	// The given block ID should match the block ID at the given height.
	header, err := v.index.Header(*block.Index)
	if err != nil {
		return identifier.Block{}, fmt.Errorf("could not get header: %w", err)
	}
	if block.Hash != "" && block.Hash != header.ID().String() {
		return identifier.Block{}, failure.InvalidBlock{
			Index: *block.Index,
			Hash:  block.Hash,
			Description: failure.NewDescription("block hash mismatches with authoritative hash for index",
				failure.WithString("want_hash", header.ID().String()),
			),
		}
	}

	// At this point, they either matched, or the block ID is empty, so we
	// should insert it.
	block.Hash = header.ID().String()

	return block, nil
}
