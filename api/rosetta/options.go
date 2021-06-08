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

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/object"
)

type OptionsRequest struct {
	NetworkID identifier.Network `json:"network_identifier"`
}

type OptionsResponse struct {
	Version object.Version
	Allow   object.Allow
}

func (d *Data) Options(ctx echo.Context) error {

	// Decode the network list request from the HTTP request JSON body.
	var req OptionsRequest
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, object.AnyError(err))
	}

	// Get our network and check it's correct.
	err = d.validate.Network(req.NetworkID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, object.AnyError(err))
	}

	// Get the current status.
	version, err := d.retrieve.Version()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, object.AnyError(err))
	}

	// Get the allowed operations.
	allow, err := d.retrieve.Allow()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, object.AnyError(err))
	}

	res := OptionsResponse{
		Version: version,
		Allow:   allow,
	}

	return ctx.JSON(http.StatusOK, res)
}
