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
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/optakt/flow-dps/rosetta/fail"
	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/meta"
)

type OptionsRequest struct {
	NetworkID identifier.Network `json:"network_identifier"`
}

type OptionsResponse struct {
	Version meta.Version `json:"version"`
	Allow   Allow        `json:"allow"`
}

type Allow struct {
	OperationStatuses       []meta.StatusDefinition `json:"operation_statuses"`
	OperationTypes          []string                `json:"operation_types"`
	Errors                  []meta.ErrorDefinition  `json:"errors"`
	HistoricalBalanceLookup bool                    `json:"historical_balance_lookup"`
}

func (d *Data) Options(ctx echo.Context) error {

	// Decode the network list request from the HTTP request JSON body.
	var req OptionsRequest
	err := ctx.Bind(&req)
	if err != nil {
		return httpError(http.StatusBadRequest, fail.InvalidFormat("could not unmarshal request", fail.WithError(err)))
	}

	if req.NetworkID.Blockchain == "" {
		return httpError(http.StatusBadRequest, fail.InvalidFormat("blockchain identifier: blockchain field is missing"))
	}
	if req.NetworkID.Network == "" {
		return httpError(http.StatusBadRequest, fail.InvalidFormat("blockchain identifier: network field is missing"))
	}

	err = d.config.Check(req.NetworkID)
	var netErr fail.InvalidNetwork
	if errors.As(err, &netErr) {
		return httpError(http.StatusUnprocessableEntity, netErr.RosettaError())
	}
	if err != nil {
		return httpError(http.StatusInternalServerError, fail.Internal("could not validate network", fail.WithError(err)))
	}

	// Create the allow object, which is native to the response.
	allow := Allow{
		OperationStatuses:       d.config.Statuses(),
		OperationTypes:          d.config.Operations(),
		Errors:                  d.config.Errors(),
		HistoricalBalanceLookup: true,
	}

	res := OptionsResponse{
		Version: d.config.Version(),
		Allow:   allow,
	}

	return ctx.JSON(http.StatusOK, res)
}
