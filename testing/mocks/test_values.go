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

package mocks

import (
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/common/hash"
	"github.com/onflow/flow-go/ledger/complete/mtrie/node"
	"github.com/onflow/flow-go/model/flow"
	"github.com/optakt/flow-dps/models/dps"
)

// Global variables that can be used for testing. They are non-nil valid values for the types commonly needed
// test DPS components.
var (
	GenericHeight = uint64(42)

	GenericCommits = []flow.StateCommitment{
		{
			0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a,
			0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a,
			0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a,
			0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a,
		},
		{
			0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b,
			0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b,
			0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b,
			0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b, 0x4b,
		},
	}

	GenericHeader = &flow.Header{
		ChainID: dps.FlowTestnet,
		Height:  GenericHeight,
	}

	GenericIdentifiers = []flow.Identifier{
		{
			0xd5, 0xf5, 0x0b, 0xc1, 0x7b, 0xa1, 0xea, 0xad,
			0x83, 0x0c, 0x86, 0xac, 0xce, 0x64, 0x5c, 0xa6,
			0xc0, 0x9f, 0xf0, 0xfe, 0xc5, 0x1c, 0x76, 0x10,
			0x03, 0x1c, 0xb9, 0x99, 0xa5, 0xb0, 0xb3, 0x22,
		},
		{
			0x2a, 0x04, 0x51, 0x3c, 0xc3, 0xc9, 0xa7, 0xf2,
			0xec, 0x08, 0x93, 0x56, 0x5f, 0x52, 0xc2, 0x9e,
			0x19, 0xf5, 0x58, 0x88, 0x10, 0x11, 0xe1, 0x13,
			0x60, 0x43, 0x9e, 0x57, 0x60, 0x18, 0xe3, 0xde,
		},
		{
			0xd6, 0x41, 0x58, 0x46, 0x8d, 0x94, 0x04, 0x77,
			0x61, 0x69, 0xfb, 0x50, 0x95, 0x64, 0xa4, 0xca,
			0x1c, 0xc1, 0x0d, 0x7e, 0xbb, 0x10, 0x8c, 0xbc,
			0xfc, 0x41, 0x94, 0xa6, 0x0e, 0x39, 0xda, 0xe1,
		},
		{
			0xd0, 0x90, 0xe0, 0x9b, 0x6d, 0x20, 0xd6, 0xd5,
			0x60, 0xa8, 0x05, 0x57, 0x71, 0x4a, 0xac, 0x1d,
			0x31, 0xfd, 0xa1, 0x6f, 0xb2, 0xdd, 0xe2, 0x58,
			0x4c, 0x06, 0x61, 0x4b, 0x6b, 0xde, 0x8c, 0xd5,
		},
		{
			0xf9, 0xa8, 0xc7, 0xbe, 0x46, 0xdd, 0x1a, 0x3c,
			0x4d, 0x66, 0xec, 0x19, 0xe2, 0x43, 0xf0, 0x15,
			0x7c, 0x90, 0xa5, 0x32, 0xe1, 0x1e, 0x15, 0xad,
			0x6b, 0xbe, 0xa0, 0x64, 0x3b, 0x83, 0xf4, 0xcc,
		},
	}

	GenericEvents = []flow.Event{
		{TransactionID: GenericIdentifiers[0]},
		{TransactionID: GenericIdentifiers[1]},
		{TransactionID: GenericIdentifiers[2]},
		{TransactionID: GenericIdentifiers[3]},
		{TransactionID: GenericIdentifiers[4]},
	}

	GenericTransactions = []*flow.TransactionBody{
		{ReferenceBlockID: GenericIdentifiers[0]},
		{ReferenceBlockID: GenericIdentifiers[1]},
		{ReferenceBlockID: GenericIdentifiers[2]},
		{ReferenceBlockID: GenericIdentifiers[3]},
		{ReferenceBlockID: GenericIdentifiers[4]},
	}

	GenericCollections = []*flow.LightCollection{
		{Transactions: GenericIdentifiers[:2]},
		{Transactions: GenericIdentifiers[2:]},
	}

	GenericLedgerPaths = []ledger.Path{
		{
			0xaa, 0xc5, 0x13, 0xeb, 0x1a, 0x04, 0x57, 0x70,
			0x0a, 0xc3, 0xfa, 0x8d, 0x29, 0x25, 0x13, 0xe1,
			0xaa, 0xc5, 0x13, 0xeb, 0x1a, 0x04, 0x57, 0x70,
			0x0a, 0xc3, 0xfa, 0x8d, 0x29, 0x25, 0x13, 0xe1,
		},
		{
			0xd5, 0x08, 0x44, 0x13, 0xdb, 0xe5, 0x2b, 0xd2,
			0x3a, 0x66, 0x7f, 0xc4, 0x08, 0xe0, 0x54, 0x60,
			0xd5, 0x08, 0x44, 0x13, 0xdb, 0xe5, 0x2b, 0xd2,
			0x3a, 0x66, 0x7f, 0xc4, 0x08, 0xe0, 0x54, 0x60,
		},
		{
			0x60, 0x0a, 0xd8, 0xa4, 0xf1, 0x6b, 0xce, 0x2e,
			0x57, 0x59, 0xfd, 0x6e, 0x45, 0xcf, 0xa9, 0xa0,
			0x60, 0x0a, 0xd8, 0xa4, 0xf1, 0x6b, 0xce, 0x2e,
			0x57, 0x59, 0xfd, 0x6e, 0x45, 0xcf, 0xa9, 0xa0,
		},
		{
			0xa5, 0x68, 0x7b, 0x2d, 0x95, 0x18, 0x7b, 0xc7,
			0xce, 0xd0, 0xe1, 0x02, 0xd6, 0xce, 0xfe, 0x93,
			0xa5, 0x68, 0x7b, 0x2d, 0x95, 0x18, 0x7b, 0xc7,
			0xce, 0xd0, 0xe1, 0x02, 0xd6, 0xce, 0xfe, 0x93,
		},
		{
			0x60, 0x0a, 0xd8, 0xa4, 0xf1, 0x6b, 0xce, 0x2e,
			0x57, 0x59, 0xfd, 0x6e, 0x45, 0xcf, 0xa9, 0xa0,
			0x60, 0x0a, 0xd8, 0xa4, 0xf1, 0x6b, 0xce, 0x2e,
			0x57, 0x59, 0xfd, 0x6e, 0x45, 0xcf, 0xa9, 0xa0,
		},
		{
			0xfd, 0x84, 0xc0, 0xa7, 0xb2, 0x35, 0xc9, 0x89,
			0xc1, 0x8e, 0x6a, 0xa2, 0x69, 0x04, 0xfe, 0xba,
			0xfd, 0x84, 0xc0, 0xa7, 0xb2, 0x35, 0xc9, 0x89,
			0xc1, 0x8e, 0x6a, 0xa2, 0x69, 0x04, 0xfe, 0xba,
		},
	}

	GenericLedgerKey = ledger.NewKey([]ledger.KeyPart{
		ledger.NewKeyPart(0, []byte(`owner`)),
		ledger.NewKeyPart(1, []byte(`controller`)),
		ledger.NewKeyPart(2, []byte(`key`)),
	})

	GenericLedgerValues = []ledger.Value{
		{0x01},
		{0x02},
		{0x03},
		{0x04},
		{0x05},
		{0x06},
	}

	GenericLedgerPayloads = []*ledger.Payload{
		ledger.NewPayload(GenericLedgerKey, GenericLedgerValues[0]),
		ledger.NewPayload(GenericLedgerKey, GenericLedgerValues[1]),
		ledger.NewPayload(GenericLedgerKey, GenericLedgerValues[2]),
		ledger.NewPayload(GenericLedgerKey, GenericLedgerValues[3]),
		ledger.NewPayload(GenericLedgerKey, GenericLedgerValues[4]),
		ledger.NewPayload(GenericLedgerKey, GenericLedgerValues[5]),
	}

	GenericTrieUpdate = &ledger.TrieUpdate{
		RootHash: ledger.RootHash{
			0xb4, 0xcd, 0xba, 0x9f, 0xcf, 0x45, 0xae, 0xff,
			0x39, 0x0a, 0x18, 0xd2, 0xb4, 0x25, 0x43, 0xfc,
			0xf6, 0xae, 0x9a, 0xc9, 0x05, 0xec, 0xd9, 0x1b,
			0x89, 0x21, 0xf4, 0x6d, 0x74, 0x02, 0x7f, 0xba,
		},
		Paths:    GenericLedgerPaths,
		Payloads: GenericLedgerPayloads,
	}

	// GenericRootNode Visual Representation:
	//           6 (root)
	//          / \
	//         3   5
	//        / \   \
	//       1   2   4
	GenericRootNode = node.NewNode(
		42,
		node.NewNode(
			42,
			node.NewLeaf(GenericLedgerPaths[0], GenericLedgerPayloads[0], 42),
			node.NewLeaf(GenericLedgerPaths[1], GenericLedgerPayloads[1], 42),
			GenericLedgerPaths[2],
			GenericLedgerPayloads[2],
			hash.DummyHash,
			64,
			64,
		),
		node.NewNode(
			42,
			node.NewLeaf(GenericLedgerPaths[3], GenericLedgerPayloads[3], 42),
			nil,
			GenericLedgerPaths[4],
			GenericLedgerPayloads[4],
			hash.DummyHash,
			64,
			64,
		),
		GenericLedgerPaths[5],
		GenericLedgerPayloads[5],
		hash.DummyHash,
		64,
		64,
	)
)
