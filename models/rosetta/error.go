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

package rosetta

// Error is used to return rich errors from the API instead of utilizing HTTP
// status codes (which often do not have a good analog). Both the code and the
// message fields can be individually used to correctly identify an error.
// Implementations must use unique values for both fields.
//
// Example for detail fields given in the Rosetta API documentation are
// `address` and `error`.
type Error struct {
	Code        uint                   `json:"code"`
	Message     string                 `json:"message"`
	Description string                 `json:"description"`
	Retriable   bool                   `json:"retriable"`
	Details     map[string]interface{} `json:"details"`
}
