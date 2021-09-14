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

package rosetta

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/object"
)

// BalanceRequest implements the request schema for /account/balance.
// See https://www.rosetta-api.org/docs/AccountApi.html#request
type BalanceRequest struct {
	NetworkID  identifier.Network    `json:"network_identifier"`
	BlockID    identifier.Block      `json:"block_identifier"`
	AccountID  identifier.Account    `json:"account_identifier"`
	Currencies []identifier.Currency `json:"currencies"`
}

// BalanceResponse implements the successful response schema for /account/balance.
// See https://www.rosetta-api.org/docs/AccountApi.html#200---ok
type BalanceResponse struct {
	BlockID  identifier.Block `json:"block_identifier"`
	Balances []object.Amount  `json:"balances"`
}

// Balance implements the /account/balance endpoint of the Rosetta Data API.
// See https://www.rosetta-api.org/docs/AccountApi.html#accountbalance
func (d *Data) Balance(ctx echo.Context) error {

	var req BalanceRequest
	err := ctx.Bind(&req)
	if err != nil {
		return unpackError(err)
	}

	err = d.validate.Request(req)
	if err != nil {
		return formatError(err)
	}

	err = d.config.Check(req.NetworkID)
	if err != nil {
		return apiError(networkCheck, err)
	}

	rosBlockID, balances, err := d.retrieve.Balances(req.BlockID, req.AccountID, req.Currencies)
	if err != nil {
		return apiError(balancesRetrieval, err)
	}

	res := BalanceResponse{
		BlockID:  rosBlockID,
		Balances: balances,
	}

	return ctx.JSON(http.StatusOK, res)
}
