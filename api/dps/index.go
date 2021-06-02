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
	"fmt"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"
)

type Index struct {
	client APIClient
}

func IndexFromAPI(client APIClient) *Index {

	i := Index{
		client: client,
	}

	return &i
}

func (i *Index) Last() (uint64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (i *Index) Header(height uint64) (*flow.Header, error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *Index) Commit(height uint64) (flow.StateCommitment, error) {
	return flow.StateCommitment{}, fmt.Errorf("not implemented")
}

func (i *Index) Events(height uint64, types ...flow.EventType) ([]flow.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *Index) Register(height uint64, path ledger.Path) (ledger.Value, error) {
	return nil, fmt.Errorf("not implemented")
}
