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

package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/models/dps/mock"
)

func TestNewController(t *testing.T) {
	c := NewController(nil)
	assert.NotNil(t, c)
}

func TestController_GetRegister(t *testing.T) {
	var (
		testHeight uint64 = 128
		lastHeight uint64 = 256

		testKey   = []byte(`testKey`)
		testValue = []byte(`testValue`)
	)

	tests := []struct {
		desc string

		reqHeight *uint64
		reqKey    []byte

		mockValue []byte
		mockErr   error

		wantResp *GetRegisterResponse
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			desc: "nominal case, height given",

			reqHeight: &testHeight,
			reqKey:    testKey,

			mockValue: testValue,

			wantResp: &GetRegisterResponse{
				Height: testHeight,
				Key:    testKey,
				Value:  testValue,
			},
			wantErr: assert.NoError,
		},
		{
			desc: "nominal case, no height given",

			reqKey: testKey,

			mockValue: testValue,

			wantResp: &GetRegisterResponse{
				Height: lastHeight,
				Key:    testKey,
				Value:  testValue,
			},
			wantErr: assert.NoError,
		},
		{
			desc: "state error",

			reqKey: testKey,

			mockErr: errors.New("dummy error"),

			wantResp: nil,
			wantErr:  assert.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			m := mock.NewState()

			if test.reqHeight != nil {
				m.RawState.On("WithHeight", *test.reqHeight).Return(m.RawState).Once()
			} else {
				m.RawState.On("WithHeight", lastHeight).Return(m.RawState).Once()
			}

			m.LastState.On("Height").Return(lastHeight).Once()
			m.RawState.On("Get", test.reqKey).Return(test.mockValue, test.mockErr)

			c := &Controller{
				state: m,
			}

			req := &GetRegisterRequest{
				Height: test.reqHeight,
				Key:    test.reqKey,
			}

			got, err := c.GetRegister(context.Background(), req)
			test.wantErr(t, err)

			if test.wantResp != nil {
				assert.Equal(t, test.wantResp, got)
			}

			m.AssertExpectations(t)
		})
	}
}

func TestController_GetValues(t *testing.T) {
	var testVersion uint64 = 42
	var (
		testKeys = []*Key{
			{
				Parts: []*KeyPart{
					{
						Type:  0,
						Value: []byte(`testOwner`),
					},
					{
						Type:  1,
						Value: []byte(`testController`),
					},
					{
						Type:  2,
						Value: []byte(`testKey`),
					},
				},
			},
		}
		testValue  = []byte(`testValue`)
		testValues = []ledger.Value{ledger.Value(testValue)}
		testCommit = flow.StateCommitment{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
		// testCommitHex = hex.EncodeToString(testCommit[:])
		lastCommit = flow.StateCommitment{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2}
		// lastCommitHex = hex.EncodeToString(lastCommit[:])
	)

	tests := []struct {
		desc string

		reqCommit  []byte
		reqVersion *uint64
		reqKeys    []*Key

		mockValues []ledger.Value
		mockErr    error

		wantResp *GetValuesResponse
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			desc: "nominal case, version and commit hash given",

			reqKeys:    testKeys,
			reqCommit:  testCommit[:],
			reqVersion: &testVersion,

			mockValues: testValues,

			wantResp: &GetValuesResponse{
				Values: [][]byte{testValue},
			},
			wantErr: assert.NoError,
		},
		{
			desc: "nominal case, version given, using latest commit",

			reqKeys:    testKeys,
			reqVersion: &testVersion,

			mockValues: testValues,

			wantResp: &GetValuesResponse{
				Values: [][]byte{testValue},
			},
			wantErr: assert.NoError,
		},
		{
			desc: "nominal case, no version or commit hash given",

			reqKeys: testKeys,

			mockValues: testValues,

			wantResp: &GetValuesResponse{
				Values: [][]byte{testValue},
			},
			wantErr: assert.NoError,
		},
		{
			desc: "invalid commit hash in request",

			reqKeys:   testKeys,
			reqCommit: []byte(`not a hexadecimal value`),

			wantErr: assert.Error,
		},
		{
			desc: "state get returns an error",

			reqKeys: testKeys,

			mockErr: errors.New("dummy error"),

			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			m := mock.NewState()
			m.LastState.On("Commit").Return(lastCommit).Once()
			m.LedgerState.On("Get", tmock.Anything).Return(test.mockValues, test.mockErr).Once()

			if test.reqVersion != nil {
				m.LedgerState.On("WithVersion", uint8(*test.reqVersion)).Return(m.LedgerState).Once()
			}

			c := &Controller{
				state: m,
			}

			req := &GetValuesRequest{
				Keys:    test.reqKeys,
				Hash:    test.reqCommit,
				Version: test.reqVersion,
			}

			got, err := c.GetValues(context.Background(), req)
			test.wantErr(t, err)

			if test.wantResp != nil {
				assert.Equal(t, test.wantResp, got)
			}

			m.AssertExpectations(t)
		})
	}
}
