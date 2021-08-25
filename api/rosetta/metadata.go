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

	"github.com/optakt/flow-dps/rosetta/failure"
	"github.com/optakt/flow-dps/rosetta/identifier"
	"github.com/optakt/flow-dps/rosetta/object"
)

// MetadataRequest implements the request schema for /construction/metadata.
// `Options` object in this request is generated by a call to `/construction/preprocess`,
// and should be sent unaltered as returned by that endpoint.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#request-3
type MetadataRequest struct {
	NetworkID identifier.Network `json:"network_identifier"`
	Options   object.Options     `json:"options"`
}

// MetadataResponse implements the response schema for /construction/metadata.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#response-3
type MetadataResponse struct {
	Metadata object.Metadata `json:"metadata"`
}

// Metadata implements the /construction/metadata endpoint of the Rosetta Construction API.
// Metadata endpoint returns information required for constructing the transaction.
// For Flow, that information includes the reference block and sequence number. Reference block
// is the last indexed block, and is used to track transaction expiration. Sequence number is
// the proposer account's public key sequence number. Sequence number is incremented for each
// transaction and is used to prevent replay attacks.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#constructionmetadata
func (c *Construction) Metadata(ctx echo.Context) error {

	var req MetadataRequest
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidEncoding(invalidJSON, err))
	}

	if req.NetworkID.Blockchain == "" {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(blockchainEmpty))
	}
	if req.NetworkID.Network == "" {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(networkEmpty))
	}

	if req.Options.AccountID.Address == "" {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(addressEmpty))
	}
	if len(req.Options.AccountID.Address) != hexAddressSize {
		return echo.NewHTTPError(http.StatusBadRequest, invalidFormat(addressLength,
			withDetail("have_length", len(req.Options.AccountID.Address)),
			withDetail("want_length", hexAddressSize),
		))
	}

	current, _, err := c.retrieve.Current()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, internal(referenceBlockRetrieval, err))
	}

	// TODO: Allow arbitrary proposal key index
	// => https://github.com/optakt/flow-dps/issues/369
	sequenceNr, err := c.retrieve.SequenceNumber(req.Options.AccountID, 0)
	var iaErr failure.InvalidAccount
	if errors.As(err, &iaErr) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, invalidAccount(iaErr))
	}
	var ipErr failure.InvalidProposalKey
	if errors.As(err, &ipErr) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, invalidProposalKey(ipErr))
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, internal(sequenceNumberRetrieval, err))
	}

	res := MetadataResponse{
		Metadata: object.Metadata{
			ReferenceBlockID: current,
			SequenceNumber:   sequenceNr,
		},
	}

	return ctx.JSON(http.StatusOK, res)
}
