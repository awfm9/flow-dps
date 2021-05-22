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

package storage

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/klauspost/compress/zstd"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"

	"github.com/awfm9/flow-dps/service/dictionaries"
)

var (
	codec             cbor.EncMode
	defaultCompressor *zstd.Encoder
	headerCompressor  *zstd.Encoder
	payloadCompressor *zstd.Encoder
	eventsCompressor  *zstd.Encoder
	decompressor      *zstd.Decoder
)

func init() {

	var err error

	codec, err = cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		panic(fmt.Errorf("could not initialize codec: %w", err))
	}

	defaultCompressor, err = zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
	)
	if err != nil {
		panic(fmt.Errorf("could not initialize default compressor: %w", err))
	}

	headerDict, err := hex.DecodeString(dictionaries.Headers)
	if err != nil {
		panic(fmt.Errorf("could not decode header dictionary: %w", err))
	}

	headerCompressor, err = zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
		zstd.WithEncoderDict(headerDict),
	)
	if err != nil {
		panic(fmt.Errorf("could not initialize header compressor: %w", err))
	}

	payloadDict, err := hex.DecodeString(dictionaries.Payloads)
	if err != nil {
		panic(fmt.Errorf("could not decode payload dictionary: %w", err))
	}

	payloadCompressor, err = zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
		zstd.WithEncoderDict(payloadDict),
	)
	if err != nil {
		panic(fmt.Errorf("could not initialize payload compressor: %w", err))
	}

	eventsDict, err := hex.DecodeString(dictionaries.Events)
	if err != nil {
		panic(fmt.Errorf("could not decode events dictionary: %w", err))
	}

	eventsCompressor, err = zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
		zstd.WithEncoderDict(eventsDict),
	)
	if err != nil {
		panic(fmt.Errorf("could not initialize events compressor: %w", err))
	}

	decompressor, err = zstd.NewReader(nil,
		zstd.WithDecoderDicts(payloadDict),
	)
	if err != nil {
		panic(fmt.Errorf("could not initialize decompressor: %w", err))
	}
}

func encodeKey(prefix uint8, segments ...interface{}) []byte {
	key := []byte{prefix}
	var val []byte
	for _, segment := range segments {
		switch s := segment.(type) {
		case uint64:
			val = make([]byte, 8)
			binary.BigEndian.PutUint64(val, s)
		case []byte:
			val = s
		case flow.Identifier:
			val = s[:]
		case ledger.Path:
			val = []byte(s)
		default:
			panic(fmt.Sprintf("unknown type (%T)", segment))
		}
		key = append(key, val...)
	}

	return key
}
