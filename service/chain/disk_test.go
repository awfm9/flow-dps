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

package chain_test

import (
	"math"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/storage/badger/operation"
	"github.com/optakt/flow-dps/testing/helpers"

	"github.com/optakt/flow-dps/service/chain"
)

const (
	testHeight  = 42
	testChainID = flow.ChainID("flow-testnet")
)

var (
	testCommit  = flow.StateCommitment{132, 131, 130, 129, 128, 127, 126, 125, 124, 123, 122, 121, 120, 119, 118, 117, 116, 115, 114, 113, 112, 111, 110, 19, 18, 17, 16, 15, 14, 13, 12, 11}
	testBlockID = flow.Identifier{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
)

func TestDisk_Root(t *testing.T) {
	db := populatedDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	root, err := c.Root()
	assert.NoError(t, err)
	assert.Equal(t, uint64(testHeight), root)
}

func TestDisk_Header(t *testing.T) {
	db := populatedDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	header, err := c.Header(testHeight)
	assert.NoError(t, err)

	require.NotNil(t, header)
	assert.Equal(t, testChainID, header.ChainID)

	_, err = c.Header(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Commit(t *testing.T) {
	db := populatedDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	commit, err := c.Commit(testHeight)
	assert.NoError(t, err)
	assert.Equal(t, testCommit, commit)

	_, err = c.Commit(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Events(t *testing.T) {
	db := populatedDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	events, err := c.Events(testHeight)
	assert.NoError(t, err)
	assert.Len(t, events, 2)

	_, err = c.Events(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Transactions(t *testing.T) {
	db := populatedDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	tt, err := c.Transactions(testHeight)
	assert.NoError(t, err)
	assert.Len(t, tt, 2)

	_, err = c.Transactions(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Collections(t *testing.T) {
	db := populatedDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	tt, err := c.Collections(testHeight)
	assert.NoError(t, err)
	assert.Len(t, tt, 2)

	_, err = c.Collections(math.MaxUint64)
	assert.Error(t, err)
}

func populatedDB(t *testing.T) *badger.DB {
	t.Helper()

	db := helpers.InMemoryDB(t)

	err := db.Update(func(tx *badger.Txn) error {
		err := operation.InsertRootHeight(testHeight)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertHeader(testBlockID, &flow.Header{ChainID: testChainID})(tx)
		if err != nil {
			return err
		}

		err = operation.IndexBlockHeight(testHeight, testBlockID)(tx)
		if err != nil {
			return err
		}

		err = operation.IndexStateCommitment(testBlockID, testCommit)(tx)
		if err != nil {
			return err
		}

		events := []flow.Event{
			{
				Type:             "test",
				TransactionIndex: 1,
				EventIndex:       2,
			},
			{
				Type:             "test",
				TransactionIndex: 3,
				EventIndex:       4,
			},
		}
		err = operation.InsertEvent(testBlockID, events[0])(tx)
		if err != nil {
			return err
		}
		err = operation.InsertEvent(testBlockID, events[1])(tx)
		if err != nil {
			return err
		}

		tb1 := flow.TransactionBody{
			ReferenceBlockID: testBlockID,
			GasLimit:         42,
			Payer:            flow.Address{0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12},
		}
		tb2 := flow.TransactionBody{
			ReferenceBlockID: testBlockID,
			GasLimit:         84,
			Payer:            flow.Address{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		}

		err = operation.InsertTransactionResult(testBlockID, &flow.TransactionResult{TransactionID: tb1.ID()})(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransactionResult(testBlockID, &flow.TransactionResult{TransactionID: tb2.ID()})(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransaction(tb1.ID(), &tb1)(tx)
		if err != nil {
			return err
		}
		err = operation.InsertTransaction(tb2.ID(), &tb2)(tx)
		if err != nil {
			return err
		}

		tb3 := flow.TransactionBody{
			ReferenceBlockID: testBlockID,
			GasLimit:         21,
			Payer:            flow.Address{0xb0, 0x20, 0xe8, 0x58, 0x72, 0xc8, 0x12, 0x59},
		}
		tb4 := flow.TransactionBody{
			ReferenceBlockID: testBlockID,
			GasLimit:         168,
			Payer:            flow.Address{0x94, 0x2f, 0x2f, 0xf3, 0x50, 0x6b, 0xa8, 0xde},
		}

		collection1 := flow.LightCollection{Transactions: []flow.Identifier{tb1.ID(), tb2.ID()}}
		collection2 := flow.LightCollection{Transactions: []flow.Identifier{tb3.ID(), tb4.ID()}}

		err = operation.IndexCollectionByTransaction(tb1.ID(), collection1.ID())(tx)
		if err != nil {
			return err
		}

		err = operation.IndexCollectionByTransaction(tb2.ID(), collection1.ID())(tx)
		if err != nil {
			return err
		}

		err = operation.IndexCollectionByTransaction(tb3.ID(), collection2.ID())(tx)
		if err != nil {
			return err
		}

		err = operation.IndexCollectionByTransaction(tb4.ID(), collection2.ID())(tx)
		if err != nil {
			return err
		}

		err = operation.InsertCollection(&collection1)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertCollection(&collection2)(tx)
		if err != nil {
			return err
		}

		return nil
	})
	require.NoError(t, err)

	return db
}
