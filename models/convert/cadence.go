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

package convert

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/onflow/cadence"

	"github.com/optakt/flow-dps/models/dps"
)

func ParseCadenceArgument(param string) (cadence.Value, error) {

	// Cadence values should be provided in the form of Type(Value), so that we
	// can unambiguously determine the type.
	re := regexp.MustCompile(`(\w+)\((.+)\)`)
	parts := re.FindStringSubmatch(param)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid parameter format (%s)", param)
	}
	typ := parts[1]
	val := parts[2]

	// Now, we can switch on the type and parse accordingly.
	switch typ {
	case "Bool":
		b, err := strconv.ParseBool(val)
		if err != nil {
			return nil, fmt.Errorf("could not parse boolean: %w", err)
		}
		return cadence.NewBool(b), nil

	case "Int":
		v, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("could not parse integer: %w", err)
		}
		return cadence.NewInt(int(v)), nil

	case "Int8":
		v, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("could not parse integer: %w", err)
		}
		return cadence.NewInt8(int8(v)), nil

	case "Int16":
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("could not parse integer: %w", err)
		}
		return cadence.NewInt16(int16(v)), nil

	case "Int32":
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("could not parse integer: %w", err)
		}
		return cadence.NewInt32(int32(v)), nil

	case "Int64":
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse integer: %w", err)
		}
		return cadence.NewInt64(v), nil

	case "Int128":
		v, ok := big.NewInt(0).SetString(val, 10)
		if !ok {
			return nil, fmt.Errorf("could not parse big integer (%s)", val)
		}
		return cadence.NewInt128FromBig(v), nil

	case "Int256":
		v, ok := big.NewInt(0).SetString(val, 10)
		if !ok {
			return nil, fmt.Errorf("could not parse big integer (%s)", val)
		}
		return cadence.NewInt256FromBig(v), nil

	case "UInt":
		v, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("could not parse unsigned integer: %w", err)
		}
		return cadence.NewUInt(uint(v)), nil

	case "UInt8":
		v, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("could not parse unsigned integer: %w", err)
		}
		return cadence.NewUInt8(uint8(v)), nil

	case "UInt16":
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("could not parse unsigned integer: %w", err)
		}
		return cadence.NewUInt16(uint16(v)), nil

	case "UInt32":
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("could not parse integer: %w", err)
		}
		return cadence.NewUInt32(uint32(v)), nil

	case "UInt64":
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse unsigned integer: %w", err)
		}
		return cadence.NewUInt64(v), nil

	case "UInt128":
		v, ok := big.NewInt(0).SetString(val, 10)
		if !ok {
			return nil, fmt.Errorf("could not parse big integer (%s)", val)
		}
		return cadence.NewUInt128FromBig(v), nil

	case "UInt256":
		v, ok := big.NewInt(0).SetString(val, 10)
		if !ok {
			return nil, fmt.Errorf("could not parse big integer (%s)", val)
		}
		return cadence.NewUInt256FromBig(v), nil

	case "UFix64":
		v, err := cadence.NewUFix64(val)
		if err != nil {
			return nil, fmt.Errorf("could not parse unsigned fixed point integer: %w", err)
		}
		return v, nil

	case "Fix64":
		v, err := cadence.NewFix64(val)
		if err != nil {
			return nil, fmt.Errorf("could not parse fixed point integer: %w", err)
		}
		return v, nil

	case "Address":
		bytes, err := hex.DecodeString(val)
		if err != nil {
			return nil, fmt.Errorf("could not decode hex string: %w", err)
		}
		return cadence.BytesToAddress(bytes), nil

	case "Bytes":
		bytes, err := hex.DecodeString(val)
		if err != nil {
			return nil, fmt.Errorf("could not decode hex string: %w", err)
		}
		return cadence.NewBytes(bytes), nil

	case "String":
		return cadence.NewString(val), nil

	default:
		return nil, fmt.Errorf("unknown type for Cadence conversion (%s)", typ)
	}
}

// ParseRosettaValue parses the input string, which should be the Rosetta value, and
// produces the Cadence UFix64 number. The function will assume that the input string
// has dps.FlowDecimals decimal places.
func ParseRosettaValue(value string) (cadence.UFix64, error) {

	var decimal int64
	var fraction uint64
	var err error

	// If the number is longer than dps.FlowDecimals (8) characters, it means
	// that we will have both the decimal and the fractional part.
	// The last 8 characters will represent the fraction, while anything before
	// that is the decimal part.
	// If the number is shorter, it means that the whole string is the fraction.
	if len(value) > dps.FlowDecimals {

		decimalsIndex := len(value) - dps.FlowDecimals

		// Get all characters except the last 8 and convert that string to an integer.
		decimal, err = strconv.ParseInt(value[:decimalsIndex], 10, 32)
		if err != nil {
			return 0, fmt.Errorf("could not parse decimal part: %w", err)
		}

		// Get the last 8 characters and convert them to an uint.
		fraction, err = strconv.ParseUint(value[decimalsIndex:], 10, 32)
		if err != nil {
			return 0, fmt.Errorf("could not parse fraction part: %w", err)
		}

	} else {
		// Parse the whole input string as the fraction.
		fraction, err = strconv.ParseUint(value, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("could not parse fraction part: %w", err)
		}
	}

	// Create UFix64 number from parsed parts.
	val, err := cadence.NewUFix64FromParts(int(decimal), uint(fraction))
	if err != nil {
		return 0, fmt.Errorf("could not create cadence UFix64 number: %w", err)
	}

	return val, nil
}
