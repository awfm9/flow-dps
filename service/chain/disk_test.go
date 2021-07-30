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

	"github.com/optakt/flow-dps/models/dps"
	"github.com/optakt/flow-dps/service/chain"
	"github.com/optakt/flow-dps/testing/helpers"
	"github.com/optakt/flow-dps/testing/mocks"
)

func TestDisk_Root(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	root, err := c.Root()
	assert.NoError(t, err)
	assert.Equal(t, mocks.GenericHeight, root)
}

func TestDisk_Header(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	header, err := c.Header(mocks.GenericHeight)
	assert.NoError(t, err)

	require.NotNil(t, header)
	assert.Equal(t, dps.FlowMainnet, header.ChainID)

	_, err = c.Header(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Commit(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	commit, err := c.Commit(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Equal(t, mocks.GenericCommit(0), commit)

	_, err = c.Commit(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Events(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	events, err := c.Events(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Len(t, events, 2)

	_, err = c.Events(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Collections(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	tt, err := c.Collections(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Len(t, tt, 2)

	_, err = c.Collections(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Guarantees(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	tt, err := c.Guarantees(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Len(t, tt, 2)

	_, err = c.Guarantees(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Transactions(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	tt, err := c.Transactions(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Len(t, tt, 4)

	_, err = c.Transactions(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_TransactionResults(t *testing.T) {
	db := populateDB(t)
	defer db.Close()
	c := chain.FromDisk(db)

	tr, err := c.Results(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Len(t, tr, 4)

	_, err = c.Results(math.MaxUint64)
	assert.Error(t, err)
}

func TestDisk_Seals(t *testing.T) {
	db := populateDB(t)
	defer db.Close()

	c := chain.FromDisk(db)

	seals, err := c.Seals(mocks.GenericHeight)
	assert.NoError(t, err)
	assert.Len(t, seals, 4)

	_, err = c.Seals(math.MaxUint64)
	assert.Error(t, err)
}

func populateDB(t *testing.T) *badger.DB {
	t.Helper()

	db := helpers.InMemoryDB(t)

	err := db.Update(func(tx *badger.Txn) error {
		err := operation.InsertRootHeight(mocks.GenericHeight)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertHeader(mocks.GenericIdentifier(0), &flow.Header{ChainID: dps.FlowMainnet})(tx)
		if err != nil {
			return err
		}

		err = operation.IndexBlockHeight(mocks.GenericHeight, mocks.GenericIdentifier(0))(tx)
		if err != nil {
			return err
		}

		err = operation.IndexStateCommitment(mocks.GenericIdentifier(0), mocks.GenericCommit(0))(tx)
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
		err = operation.InsertEvent(mocks.GenericIdentifier(0), events[0])(tx)
		if err != nil {
			return err
		}
		err = operation.InsertEvent(mocks.GenericIdentifier(0), events[1])(tx)
		if err != nil {
			return err
		}

		tb1 := flow.TransactionBody{
			ReferenceBlockID: mocks.GenericIdentifier(0),
			GasLimit:         42,
			Payer:            flow.Address{0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12, 0x12},
		}
		tb2 := flow.TransactionBody{
			ReferenceBlockID: mocks.GenericIdentifier(0),
			GasLimit:         84,
			Payer:            flow.Address{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		}
		tb3 := flow.TransactionBody{
			ReferenceBlockID: mocks.GenericIdentifier(0),
			GasLimit:         21,
			Payer:            flow.Address{0xb0, 0x20, 0xe8, 0x58, 0x72, 0xc8, 0x12, 0x59},
		}
		tb4 := flow.TransactionBody{
			ReferenceBlockID: mocks.GenericIdentifier(0),
			GasLimit:         168,
			Payer:            flow.Address{0x94, 0x2f, 0x2f, 0xf3, 0x50, 0x6b, 0xa8, 0xde},
		}

		err = operation.InsertTransaction(tb1.ID(), &tb1)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransaction(tb2.ID(), &tb2)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransaction(tb3.ID(), &tb3)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransaction(tb4.ID(), &tb4)(tx)
		if err != nil {
			return err
		}

		collection1 := flow.LightCollection{Transactions: []flow.Identifier{tb1.ID(), tb2.ID()}}
		collection2 := flow.LightCollection{Transactions: []flow.Identifier{tb3.ID(), tb4.ID()}}

		err = operation.InsertCollection(&collection1)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertCollection(&collection2)(tx)
		if err != nil {
			return err
		}

		guarantee1 := &flow.CollectionGuarantee{
			CollectionID: collection1.ID(),
			Signature:    mocks.GenericBytes,
		}
		guarantee2 := &flow.CollectionGuarantee{
			CollectionID: collection2.ID(),
			Signature:    mocks.GenericBytes,
		}

		err = operation.InsertGuarantee(guarantee1.CollectionID, guarantee1)(tx)
		if err != nil {
			return err
		}

		err = operation.InsertGuarantee(guarantee2.CollectionID, guarantee2)(tx)
		if err != nil {
			return err
		}

		err = operation.IndexPayloadGuarantees(mocks.GenericIdentifier(0), []flow.Identifier{collection1.ID(), collection2.ID()})(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransactionResult(mocks.GenericIdentifier(0), &flow.TransactionResult{TransactionID: tb1.ID()})(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransactionResult(mocks.GenericIdentifier(0), &flow.TransactionResult{TransactionID: tb2.ID()})(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransactionResult(mocks.GenericIdentifier(0), &flow.TransactionResult{TransactionID: tb3.ID()})(tx)
		if err != nil {
			return err
		}

		err = operation.InsertTransactionResult(mocks.GenericIdentifier(0), &flow.TransactionResult{TransactionID: tb4.ID()})(tx)
		if err != nil {
			return err
		}

		seals := mocks.GenericSeals(4)
		var sealIDs []flow.Identifier
		for _, seal := range seals {
			err = operation.InsertSeal(seal.ID(), seal)(tx)
			if err != nil {
				return err
			}

			sealIDs = append(sealIDs, seal.ID())
		}

		err = operation.IndexPayloadSeals(mocks.GenericIdentifier(0), sealIDs)(tx)
		if err != nil {
			return err
		}

		return nil
	})
	require.NoError(t, err)

	return db
}
