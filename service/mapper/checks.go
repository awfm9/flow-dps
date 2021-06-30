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

const (
	stateEmpty = iota + 1
	stateActive
	stateIndexed
	stateForwarded
)

type CheckFunc func(*State) bool

func Empty(s *State) bool {
	return s.state == stateEmpty
}

func Ready(s *State) bool {
	return s.state == stateActive && !s.forest.Has(s.next)
}

func Matched(s *State) bool {
	return s.state == stateActive && s.forest.Has(s.next)
}

func Indexed(s *State) bool {
	return s.state == stateIndexed
}

func Forwarded(s *State) bool {
	return s.state == stateForwarded
}
