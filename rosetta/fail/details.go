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

package fail

// Detail represents a function that can be used to provide more detailed information
// about a specific error instance.
type Detail func(*Error)

// WithInt adds an integer value to the error details.
func WithInt(key string, value int) Detail {
	return func(err *Error) {
		err.Details[key] = value
	}
}

// WithUint adds an unsigned integer value to the error details.
func WithUint(key string, value uint) Detail {
	return func(err *Error) {
		err.Details[key] = value
	}
}

// WithUint64 adds a 64-bit unsigned integer value to the error details.
func WithUint64(key string, value uint64) Detail {
	return func(err *Error) {
		err.Details[key] = value
	}
}

// WithString adds a textual value to the error details.
func WithString(key string, value string) Detail {
	return func(err *Error) {
		err.Details[key] = value
	}
}

// WithError adds the information about a specific error to the error details.
func WithError(e error) Detail {
	return func(err *Error) {
		err.Details["error"] = e.Error()
	}
}
