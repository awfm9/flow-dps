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
	"github.com/onflow/flow-go/ledger/complete/mtrie/trie"
)

// Config contains optional parameters we can set for the mapper.
type Config struct {
	CheckpointFile string
	PostProcessing func(*trie.MTrie)
}

// WithCheckpointFile will initialize the mapper's internal trie with the trie
// from the provided checkpoint file.
func WithCheckpointFile(file string) func(*Config) {
	return func(cfg *Config) {
		cfg.CheckpointFile = file
	}
}

// WithPostProcessing will provide a callback that allows post-processing of the
// final state trie.
func WithPostProcessing(post func(*trie.MTrie)) func(*Config) {
	return func(cfg *Config) {
		cfg.PostProcessing = post
	}
}
