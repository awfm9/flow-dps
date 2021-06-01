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

package height

import (
	"fmt"

	"github.com/onflow/flow-go/model/flow"
	"github.com/optakt/flow-dps/api/dps"
)

func FromDPS(client dps.APIClient) func(flow.Identifier) (uint64, error) {
	return func(blockID flow.Identifier) (uint64, error) {
		return 0, fmt.Errorf("not implemented")
	}
}
