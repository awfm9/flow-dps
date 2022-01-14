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

package trie_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/common/hash"
	"github.com/onflow/flow-go/ledger/common/utils"

	"github.com/optakt/flow-dps/ledger/trie"
	"github.com/optakt/flow-dps/testing/helpers"
)

func Test_DecodeLegacyFormat(t *testing.T) {
	path := utils.PathByUint8(3)
	h := ledger.GetDefaultHashForHeight(241)

	lightNode := &trie.LightNode{
		LIndex:    1,
		RIndex:    2,
		Height:    241,
		Path:      path[:],
		HashValue: h[:],
	}
	wantPayload := utils.LightPayload8('A', 'a')

	store := helpers.InMemoryStore(t)
	defer store.Close()

	// Version 0.
	encoded := []byte{
		0x00, 0x00, // Encoding version
		0x00, 0xf1, // Height
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // LIndex
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // RIndex
		0x00, 0x07, // Max depth
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x13, 0x88, // Register count
		0x00, 0x20, // Path data len
		0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Path data
		0x00, 0x00, 0x00, 0x19, // Payload data len
		0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x09, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x61, // Payload data
		0x00, 0x20, // HashValue length
		0x2e, 0xe8, 0x22, 0x96, 0x86, 0xc1, 0xd0, 0xc8, 0x72, 0x80, 0x58, 0x67, 0x3d, 0x50, 0xc8, 0x68, 0xe4, 0xb0, 0x37, 0x1a, 0x1, 0x28, 0x55, 0x23, 0x49, 0x3b, 0x1e, 0x9f, 0xf2, 0x47, 0xc0, 0x50, // Hash data
	}

	reader := bytes.NewReader(encoded)
	newStorableNode, err := trie.DecodeLightNode(reader, store)
	require.NoError(t, err)

	assert.Equal(t, lightNode, newStorableNode)

	key, err := hash.ToHash(lightNode.HashValue)
	require.NoError(t, err)

	got, err := store.Retrieve(key)
	assert.Equal(t, wantPayload.Value[:], got.Value[:])
}
