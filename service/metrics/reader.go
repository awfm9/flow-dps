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

package metrics

import (
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/models/dps"
)

type Reader struct {
	read dps.Reader
}

func NewReader(read dps.Reader) *Reader {
	r := Reader{
		read: read,
	}
	return &r
}

func (r *Reader) First() (uint64, error) {
	return r.read.First()
}

func (r *Reader) Last() (uint64, error) {
	return r.read.Last()
}

func (r *Reader) HeightForBlock(blockID flow.Identifier) (uint64, error) {
	return r.read.HeightForBlock(blockID)
}

func (r *Reader) Commit(height uint64) (flow.StateCommitment, error) {
	return r.read.Commit(height)
}

func (r *Reader) Header(height uint64) (*flow.Header, error) {
	return r.read.Header(height)
}

func (r *Reader) Values(height uint64, paths []ledger.Path) ([]ledger.Value, error) {
	return r.read.Values(height, paths)
}

func (r *Reader) Transaction(txID flow.Identifier) (*flow.TransactionBody, error) {
	return r.read.Transaction(txID)
}

func (r *Reader) TransactionsByHeight(height uint64) ([]flow.Identifier, error) {
	return r.read.TransactionsByHeight(height)
}

func (r *Reader) Events(height uint64, types ...flow.EventType) ([]flow.Event, error) {
	return r.read.Events(height, types...)
}
