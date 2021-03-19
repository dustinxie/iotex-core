// Copyright (c) 2021 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package action

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-core/pkg/version"
)

// RlpTx defines the struct of RLP-encoded transaction
type RlpTx struct {
	AbstractAction
	chainID   uint32
	to        string
	amount    *big.Int
	hash      hash.Hash256
	rawHash   hash.Hash256
	data, sig []byte
}

// Amount returns the amount
func (rx *RlpTx) Amount() *big.Int { return rx.amount }

// Recipient returns the recipient address
func (rx *RlpTx) Recipient() string { return rx.to }

// TotalSize returns the total size of this Transfer
func (rx *RlpTx) TotalSize() uint32 {
	size := rx.BasicActionSize()
	if rx.amount != nil && len(rx.amount.Bytes()) > 0 {
		size += uint32(len(rx.amount.Bytes()))
	}
	return size + uint32(len(rx.data))
}

// Contract() returns the target contract address
func (rx *RlpTx) Contract() string { return rx.to }

// Data() returns the data payload
func (rx *RlpTx) Data() []byte { return rx.data }

// Proto converts Execution to protobuf's Execution
func (rx *RlpTx) Proto() *iotextypes.RlpTransaction {
	return &iotextypes.RlpTransaction{
		ChainID: rx.chainID,
		Data:    rx.data,
	}
}

// LoadProto converts a protobuf's RlpTransaction to RlpTx
func (rx *RlpTx) LoadProto(pbAct *iotextypes.RlpTransaction) error {
	if pbAct == nil {
		return errors.New("empty action proto to load")
	}
	if rx == nil {
		return errors.New("nil action to load proto")
	}

	rx.chainID = pbAct.GetChainID()
	rx.data = pbAct.GetData()
	return rx.decode()
}

func (rx *RlpTx) decode() error {
	var err error
	tx := types.Transaction{}
	if err = tx.DecodeRLP(
		rlp.NewStream(bytes.NewReader(rx.data), uint64(len(rx.data)))); err != nil {
		return err
	}

	rx.version = version.ProtocolVersion
	rx.nonce = tx.Nonce()
	rx.gasLimit = tx.Gas()
	rx.gasPrice = tx.GasPrice()
	rx.amount = tx.Value()
	rx.data = tx.Data()
	rx.hash = hash.BytesToHash256(tx.Hash().Bytes())
	rx.to = EmptyAddress
	if to := tx.To(); to != nil {
		addr, err := address.FromBytes(to.Bytes())
		if err != nil {
			return err
		}
		rx.to = addr.String()
	}

	// extract signature
	v, r, s := tx.RawSignatureValues()
	rx.sig = make([]byte, 64, 65)
	r.FillBytes(rx.sig[:32])
	s.FillBytes(rx.sig[32:])
	recId := uint32(v.Uint64()) - 2*rx.chainID - 8
	if recId >= 27 {
		recId -= 27
	}
	rx.sig = append(rx.sig, byte(recId))

	// recover public key
	rawHash := types.NewEIP155Signer(big.NewInt(int64(rx.chainID))).Hash(&tx)
	rx.rawHash = hash.BytesToHash256(rawHash[:])
	if rx.srcPubkey, err = crypto.RecoverPubkey(rawHash[:], rx.sig); err != nil {
		return err
	}
	return nil
}

// backfill fills RlpTx's data back to Envelope
func (rx *RlpTx) backfill(elp *Envelope) {
	elp.nonce = rx.nonce
	elp.gasLimit = rx.gasLimit
	elp.gasPrice = rx.gasPrice
	elp.srcPubkey = rx.srcPubkey
	elp.signature = rx.sig
}

// IsRLP returns true
func (rx *RlpTx) IsRLP() bool {
	return true
}

// SetEnvelopeContext overrides AbstractAction
func (rx *RlpTx) SetEnvelopeContext(seal SealedEnvelope) {
	panic("should not call SetEnvelopeContext for RlpTx")
}

// SanityCheck validates the variables in the action
func (rx *RlpTx) SanityCheck() error {
	// Reject transfer of negative amount
	if rx.amount.Sign() < 0 {
		return errors.Wrap(ErrBalance, "negative value")
	}
	// check if recipient's address is valid
	if rx.to != EmptyAddress {
		if _, err := address.FromString(rx.to); err != nil {
			return errors.Wrapf(err, "error when validating recipient's address %s", rx.to)
		}
	}

	return rx.AbstractAction.SanityCheck()
}

// RawHash returns the tx's raw hash to be signed
func (rx *RlpTx) RawHash() hash.Hash256 {
	return rx.rawHash
}

// Hash returns the RLP-encoded tx's hash
func (rx *RlpTx) Hash() hash.Hash256 {
	return rx.hash
}

// Cost returns the total cost
func (rx *RlpTx) Cost() (*big.Int, error) {
	maxExecFee := new(big.Int).SetUint64(rx.gasLimit)
	maxExecFee.Mul(maxExecFee, rx.GasPrice())
	return maxExecFee.Add(maxExecFee, rx.amount), nil
}

// IntrinsicGas returns the intrinsic gas
func (rx *RlpTx) IntrinsicGas() (uint64, error) {
	dataSize := uint64(len(rx.data))
	return calculateIntrinsicGas(ExecutionBaseIntrinsicGas, ExecutionDataGas, dataSize)
}
