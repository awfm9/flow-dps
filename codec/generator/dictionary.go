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

package generator

import (
	"time"
)

// DictionaryKind represents the resources that the dictionary is trained to compress efficiently.
type DictionaryKind string

// Supported dictionary kinds.
const (
	KindPayloads     DictionaryKind = "payloads"
	KindEvents       DictionaryKind = "events"
	KindTransactions DictionaryKind = "transactions"
)

func (k DictionaryKind) String() string {
	return string(k)
}

type dictionary struct {
	kind DictionaryKind
	raw  []byte
	size int

	ratio    float64
	duration time.Duration
}
