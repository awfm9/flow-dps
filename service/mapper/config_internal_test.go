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

package mapper

import (
	"testing"
	"time"

	"github.com/onflow/flow-go/ledger/complete/mtrie/trie"
	"github.com/stretchr/testify/assert"
)

func TestWithRootCheckpoint(t *testing.T) {
	c := &Config{}
	trie := trie.NewEmptyMTrie()

	WithRootTrie(trie)(c)

	assert.Equal(t, trie, c.RootTrie)
}

func TestWithSkipRegisters(t *testing.T) {
	c := &Config{}

	WithSkipRegisters(true)(c)

	assert.Equal(t, true, c.SkipRegisters)
}

func TestWithWaitInterval(t *testing.T) {
	c := &Config{}
	dur := 13 * time.Second

	WithWaitInterval(dur)(c)

	assert.Equal(t, dur, c.WaitInterval)
}
