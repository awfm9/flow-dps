// Copyright 2021 Alvalor S.A.
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

package dps

import (
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"
)

type IndexReader interface {
	Last() (uint64, error)
	Header(height uint64) (*flow.Header, error)
	Commit(height uint64) (flow.StateCommitment, error)
	Events(height uint64, types ...flow.EventType) ([]flow.Event, error)
	Register(height uint64, path ledger.Path) (ledger.Value, error)
}

type IndexWriter interface {
	Header(height uint64, header *flow.Header) error
	Commit(height uint64, commit flow.StateCommitment) error
	Events(height uint64, events []flow.Event) error
	Payloads(height uint64, path []ledger.Path, value []*ledger.Payload) error
	Last(height uint64) error
}
