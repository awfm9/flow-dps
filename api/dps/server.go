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

package dps

import (
	"context"
	"fmt"

	"github.com/fxamacker/cbor/v2"

	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/models/convert"
	"github.com/optakt/flow-dps/models/index"
)

// Server is a simple implementation of the generated APIServer interface. It
// uses an index reader interface as the backend to retrieve the desired data.
// This is generally an on-disk interface, but could be a GRPC-based index as
// well, in which case there is a double redirection.
type Server struct {
	index index.Reader
	codec index.Codec
}

// NewServer creates a new server, using the provided index reader as a backend
// for data retrieval.
func NewServer(index index.Reader, codec index.Codec) *Server {

	s := Server{
		index: index,
		codec: codec,
	}

	return &s
}

// GetFirst implements the `GetFirst` method of the generated GRPC server.
func (s *Server) GetFirst(_ context.Context, _ *GetFirstRequest) (*GetFirstResponse, error) {

	height, err := s.index.First()
	if err != nil {
		return nil, fmt.Errorf("could not get first height: %w", err)
	}

	res := GetFirstResponse{
		Height: height,
	}

	return &res, nil
}

// GetLast implements the `GetLast` method of the generated GRPC server.
func (s *Server) GetLast(_ context.Context, _ *GetLastRequest) (*GetLastResponse, error) {

	height, err := s.index.Last()
	if err != nil {
		return nil, fmt.Errorf("could not get last height: %w", err)
	}

	res := GetLastResponse{
		Height: height,
	}

	return &res, nil
}

// GetHeader implements the `GetHeader` method of the generated GRPC server.
func (s *Server) GetHeader(_ context.Context, req *GetHeaderRequest) (*GetHeaderResponse, error) {

	header, err := s.index.Header(req.Height)
	if err != nil {
		return nil, fmt.Errorf("could not get header: %w", err)
	}

	// The header is encoded using CBOR with canonical encoding options.
	data, err := s.codec.Marshal(header)
	if err != nil {
		return nil, fmt.Errorf("could not encode header: %w", err)
	}

	res := GetHeaderResponse{
		Height: req.Height,
		Data:   data,
	}

	return &res, nil
}

// GetCommit implements the `GetCommit` method of the generated GRPC server.
func (s *Server) GetCommit(_ context.Context, req *GetCommitRequest) (*GetCommitResponse, error) {

	commit, err := s.index.Commit(req.Height)
	if err != nil {
		return nil, fmt.Errorf("could not get commit: %w", err)
	}

	res := GetCommitResponse{
		Height: req.Height,
		Commit: commit[:],
	}

	return &res, nil
}

// GetEvents implements the `GetEvents` method of the generated GRPC server.
func (s *Server) GetEvents(_ context.Context, req *GetEventsRequest) (*GetEventsResponse, error) {

	types := make([]flow.EventType, 0, len(req.Types))
	for _, typ := range req.Types {
		types = append(types, flow.EventType(typ))
	}

	events, err := s.index.Events(req.Height, types...)
	if err != nil {
		return nil, fmt.Errorf("could not get events: %w", err)
	}

	// The events are CBOR-encoded with canonical encoding options.
	data, err := s.codec.Marshal(events)
	if err != nil {
		return nil, fmt.Errorf("could not encode events: %w", err)
	}

	res := GetEventsResponse{
		Height: req.Height,
		Types:  req.Types,
		Data:   data,
	}

	return &res, nil
}

// GetRegisters implements the `GetRegisters` function of the generated GRPC
// server.
func (s *Server) GetRegisters(_ context.Context, req *GetRegistersRequest) (*GetRegistersResponse, error) {

	paths, err := convert.BytesToPaths(req.Paths)
	if err != nil {
		return nil, fmt.Errorf("could not convert paths: %w", err)
	}

	values, err := s.index.Registers(req.Height, paths)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve registers: %w", err)
	}

	res := GetRegistersResponse{
		Height: req.Height,
		Paths:  req.Paths,
		Values: convert.ValuesToBytes(values),
	}

	return &res, nil
}

// GetHeight implements the `GetHeight` function of the generated GRPC
// server.
func (s *Server) GetHeight(_ context.Context, req *GetHeightRequest) (*GetHeightResponse, error) {
	var blockID flow.Identifier
	copy(blockID[:], req.BlockID)

	height, err := s.index.Height(blockID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve height: %w", err)
	}

	res := GetHeightResponse{
		BlockID: req.BlockID,
		Height:  height,
	}

	return &res, nil
}

// GetTransaction implements the `GetTransaction` function of the generated GRPC
// server.
func (s *Server) GetTransaction(_ context.Context, req *GetTransactionRequest) (*GetTransactionResponse, error) {
	transactionID := flow.HashToID(req.TransactionID)

	transaction, err := s.index.Transaction(transactionID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve transaction: %w", err)
	}

	transactionData, err := cbor.Marshal(transaction)
	if err != nil {
		return nil, fmt.Errorf("could not encode transaction: %w", err)
	}

	res := GetTransactionResponse{
		TransactionID: req.TransactionID,
		Data:          transactionData,
	}

	return &res, nil
}

// ListTransactionsForBlock implements the `ListTransactionsForBlock` function of the generated GRPC
// server.
func (s *Server) ListTransactionsForBlock(_ context.Context, req *ListTransactionsForBlockRequest) (*ListTransactionsForBlockResponse, error) {
	var blockID flow.Identifier
	copy(blockID[:], req.BlockID)

	tt, err := s.index.Transactions(blockID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve transactions: %w", err)
	}

	var transactions [][]byte
	for _, t := range tt {
		transactions = append(transactions, t[:])
	}

	res := ListTransactionsForBlockResponse{
		BlockID:        req.BlockID,
		TransactionIDs: transactions,
	}

	return &res, nil
}

// ListTransactionsForCollection implements the `ListTransactionsForCollection` function of the generated GRPC
// server.
func (s *Server) ListTransactionsForCollection(_ context.Context, req *ListTransactionsForCollectionRequest) (*ListTransactionsForCollectionResponse, error) {
	collectionID := flow.HashToID(req.CollectionID)

	collection, err := s.index.Collection(collectionID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve collection: %w", err)
	}

	var transactionIDs [][]byte
	for _, tr := range collection.Transactions {
		transactionIDs = append(transactionIDs, tr[:])
	}

	res := ListTransactionsForCollectionResponse{
		CollectionID:   req.CollectionID,
		TransactionIDs: transactionIDs,
	}

	return &res, nil
}

// ListCollectionsForBlock implements the `ListCollectionsForBlock` function of the generated GRPC
// server.
func (s *Server) ListCollectionsForBlock(_ context.Context, req *ListCollectionsForBlockRequest) (*ListCollectionsForBlockResponse, error) {
	var blockID flow.Identifier
	copy(blockID[:], req.BlockID)

	cc, err := s.index.Collections(blockID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve collections: %w", err)
	}

	var collections [][]byte
	for _, c := range cc {
		collections = append(collections, c[:])
	}

	res := ListCollectionsForBlockResponse{
		BlockID:       req.BlockID,
		CollectionIDs: collections,
	}

	return &res, nil
}
