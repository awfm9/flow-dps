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

// +build integration

package rosetta_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/klauspost/compress/zstd"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/optakt/flow-dps/api/rosetta"
	"github.com/optakt/flow-dps/models/dps"
	"github.com/optakt/flow-dps/rosetta/configuration"
	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/invoker"
	"github.com/optakt/flow-dps/rosetta/retriever"
	"github.com/optakt/flow-dps/rosetta/scripts"
	"github.com/optakt/flow-dps/rosetta/validator"
	"github.com/optakt/flow-dps/service/dictionaries"
	"github.com/optakt/flow-dps/service/index"
	"github.com/optakt/flow-dps/testing/snapshots"
)

func setupDB(t *testing.T) *badger.DB {
	t.Helper()

	opts := badger.DefaultOptions("").
		WithInMemory(true).
		WithReadOnly(true).
		WithLogger(nil)

	db, err := badger.Open(opts)
	require.NoError(t, err)

	reader := hex.NewDecoder(strings.NewReader(snapshots.Rosetta))
	dict, _ := hex.DecodeString(dictionaries.Payload)

	decompressor, err := zstd.NewReader(reader,
		zstd.WithDecoderDicts(dict),
	)
	require.NoError(t, err)

	err = db.Load(decompressor, runtime.GOMAXPROCS(0))
	require.NoError(t, err)

	return db
}

func setupAPI(t *testing.T, db *badger.DB) *rosetta.Data {
	t.Helper()

	index := index.NewReader(db)

	params := dps.FlowParams[dps.FlowTestnet]
	config := configuration.New(params.ChainID)
	validate := validator.New(params, index)
	generate := scripts.NewGenerator(params)
	invoke := invoker.New(index)
	retrieve := retriever.New(params, index, validate, generate, invoke)
	controller := rosetta.NewData(config, retrieve)

	return controller
}
func TestGetBalance(t *testing.T) {

	db := setupDB(t)
	api := setupAPI(t, db)

	tests := []struct {
		name string

		request rosetta.BalanceRequest

		wantStatusCode int
		wantBalance    string
		wantHandlerErr assert.ErrorAssertionFunc
	}{
		{
			name:           "valid balance request - first occurrence of the account",
			request:        balanceRequest("754aed9de6197641", 13, "af528bb047d6cd1400a326bb127d689607a096f5ccd81d8903dfebbac26afb23"),
			wantBalance:    "10000100000",
			wantStatusCode: http.StatusOK,
			wantHandlerErr: assert.NoError,
		},
		{
			name:           "valid balance request - mid chain",
			request:        balanceRequest("754aed9de6197641", 50, "d99888d47dc326fed91087796865316ac71863616f38fa0f735bf1dfab1dc1df"),
			wantBalance:    "10000099999",
			wantStatusCode: http.StatusOK,
			wantHandlerErr: assert.NoError,
		},
		{
			name:           "valid balance request - last indexed block",
			request:        balanceRequest("754aed9de6197641", 425, "594d59b2e61bb18b149ffaac2b27b0efe1854f6795cd3bb96a443c3676d78683"),
			wantBalance:    "10000100002",
			wantStatusCode: http.StatusOK,
			wantHandlerErr: assert.NoError,
		},
		{
			name:           "empty balance request",
			request:        rosetta.BalanceRequest{},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantHandlerErr: assert.Error,
		},
		/*
			TODO: when we re-add block validation, we should reintroduce this test case - block height and hash mismatch should result in an error
				=> https://github.com/optakt/flow-dps/issues/51
			{
				name:           "invalid request - block hash and height mismatch",
				request:        balanceRequest("754aed9de6197641", 50, "594d59b2e61bb18b149ffaac2b27b0efe1854f6795cd3bb96a443c3676d78683"),
				wantStatusCode: http.StatusUnprocessableEntity,
				wantHandlerErr: assert.Error,
			},
		*/
		{
			name:           "invalid account address",
			request:        balanceRequest("invalid_address", 0, "d47b1bf7f37e192cf83d2bee3f6332b0d9b15c0aa7660d1e5322ea964667b333"),
			wantStatusCode: http.StatusUnprocessableEntity,
			wantHandlerErr: assert.Error,
		},
		{
			name:           "invalid balance request - account does not exist yet",
			request:        balanceRequest("754aed9de6197641", 12, "9035c558379b208eba11130c928537fe50ad93cdee314980fccb695aa31df7fc"),
			wantStatusCode: http.StatusInternalServerError,
			wantHandlerErr: assert.Error,
		},
		{
			name:           "invalid balance request - unknown block",
			request:        balanceRequest("754aed9de6197641", 426, "xyz"),
			wantStatusCode: http.StatusInternalServerError,
			wantHandlerErr: assert.Error,
		},
	}

	for _, test := range tests {

		test := test
		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			// prepare request payload (JSON)
			enc, err := json.Marshal(test.request)
			require.NoError(t, err)

			// create request
			req := httptest.NewRequest(http.MethodPost, "/account/balance", bytes.NewReader(enc))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()

			ctx := echo.New().NewContext(req, rec)

			// execute the request
			err = api.Balance(ctx)
			test.wantHandlerErr(t, err)

			// validate response data

			// validate status code
			if test.wantStatusCode != http.StatusOK {

				// ugly workaround for HTTP status code set by router vs set on context
				e, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				assert.Equal(t, test.wantStatusCode, e.Code)

				// nothing more to do, response validation should only be done for '200 OK' responses
				return
			}

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)

			// unpack response
			var balanceResponse rosetta.BalanceResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &balanceResponse))

			// verify that the block data matches the input data
			assert.Equal(t, test.request.BlockID.Index, balanceResponse.BlockID.Index)
			assert.Equal(t, test.request.BlockID.Hash, balanceResponse.BlockID.Hash)

			// verify that we have at least one balance in the response
			if assert.Len(t, balanceResponse.Balances, 1) {

				// verify the balance data - both the value and that the output matches the input spec
				balance := balanceResponse.Balances[0]
				assert.Equal(t, test.request.Currencies[0].Symbol, balance.Currency.Symbol)
				assert.Equal(t, test.request.Currencies[0].Decimals, balance.Currency.Decimals)
				assert.Equal(t, test.wantBalance, balance.Value)
			}
		})
	}
}

// TestGetBalanceBadRequest tests whether an improper JSON (e.g. wrong field types) will cause a '400 Bad Request' error
func TestGetBalanceBadRequest(t *testing.T) {

	db := setupDB(t)
	api := setupAPI(t, db)

	// JSON with an invalid structure (integer instead of string for network name)
	payload := `{ "network_identifier": { "blockchain": "flow", "network": 99} }`

	// create request
	req := httptest.NewRequest(http.MethodPost, "/account/balance", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	ctx := echo.New().NewContext(req, rec)

	// execute the request
	err := api.Balance(ctx)
	assert.Error(t, err)

	e, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, e.Code)
}

// balanceRequest generates a BalanceRequest with the specified parameters.
func balanceRequest(address string, blockIndex uint64, blockHash string) rosetta.BalanceRequest {

	return rosetta.BalanceRequest{
		NetworkID: defaultNetworkID(),
		AccountID: identifier.Account{
			Address: address,
		},
		BlockID: identifier.Block{
			Index: blockIndex,
			Hash:  blockHash,
		},
		Currencies: defaultCurrencySpec(),
	}
}

// defaultNetworkID returns the Network identifier common for all requests.
func defaultNetworkID() identifier.Network {
	return identifier.Network{
		Blockchain: dps.FlowBlockchain,
		Network:    dps.FlowTestnet.String(),
	}
}

// defaultCurrencySpec returns the Currency spec common for all requests.
// At the moment only get the FLOW tokens, perhaps in the future it will support multiple.
func defaultCurrencySpec() []identifier.Currency {
	return []identifier.Currency{
		{
			Symbol:   dps.FlowSymbol,
			Decimals: dps.FlowDecimals,
		},
	}
}
