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

// PreprocessRequest implements the request schema for /construction/preprocess.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#request-6
type PreprocessRequest struct {
	NetworkID  identifier.Network `json:"network_identifier"`
	Operations []object.Operation `json:"operations"`
}

// PreprocessResponse implements the response schema for /construction/preprocess.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#response-6
type PreprocessResponse struct {
	object.Options `json:"options,omitempty"`
}

// Preprocess implements the /construction/preprocess endpoint of the Rosetta Construction API.
// Preprocess receives a list of operations that should deterministically specify the
// intent of the transaction. Preprocess endpoint returns the `options` object that
// will be sent **unmodified** to /construction/metadata, effectively creating the metadata
// request.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#constructionpreprocess
func (c *Construction) Preprocess(ctx echo.Context) error {

	var req PreprocessRequest
	err := ctx.Bind(&req)
	if err != nil {
		return unpackError(err)
	}

	err = c.validate.Request(req)
	if err != nil {
		return formatError(err)
	}

	err = c.config.Check(req.NetworkID)
	if err != nil {
		return apiError(networkCheck, err)
	}

	intent, err := c.transact.DeriveIntent(req.Operations)
	if err != nil {
		return apiError(intentDetermination, err)
	}

	res := PreprocessResponse{
		Options: object.Options{
			AccountID: identifier.Account{
				Address: intent.From.Hex(),
			},
		},
	}

	return ctx.JSON(http.StatusOK, res)
}
