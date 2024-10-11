// Copyright (c) 2024 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package action

import (
	"context"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
)

const (
	LegacyTxType     = 0
	AccessListTxType = 1
	DynamicFeeTxType = 2
	BlobTxType       = 3
)

type (
	// Action is the action can be Executed in protocols. The method is added to avoid mistakenly used empty interface as action.
	Action interface {
		SanityCheck() error
	}

	// EthCompatibleAction is the action which is compatible to be converted to eth tx
	EthCompatibleAction interface {
		EthTo() (*common.Address, error)
		Value() *big.Int
		EthData() ([]byte, error)
	}

	TxContainer interface {
		Unfold(*SealedEnvelope, context.Context, func(context.Context, *common.Address) (bool, bool, bool, error)) error // unfold the tx inside the container
	}

	actionPayload interface {
		IntrinsicGas() (uint64, error)
		SanityCheck() error
		FillAction(*iotextypes.ActionCore)
	}

	hasDestination interface{ Destination() string }

	hasSize interface{ Size() uint32 }

	amountForCost interface{ Amount() *big.Int }

	gasLimitForCost interface{ GasLimitForCost() }

	validateSidecar interface{ ValidateSidecar() error }

	protoForRawHash interface{ ProtoForRawHash() *iotextypes.ActionCore }
)

// Sign signs the action using sender's private key
func Sign(act Envelope, sk crypto.PrivateKey) (*SealedEnvelope, error) {
	sealed := &SealedEnvelope{
		Envelope:  act,
		srcPubkey: sk.PublicKey(),
	}

	h, err := sealed.envelopeHash()
	if err != nil {
		return sealed, errors.Wrap(err, "failed to generate envelope hash")
	}
	sig, err := sk.Sign(h[:])
	if err != nil {
		return sealed, ErrInvalidSender
	}
	sealed.signature = sig
	return sealed, nil
}

// FakeSeal creates a SealedActionEnvelope without signature.
// This method should be only used in tests.
func FakeSeal(act Envelope, pubk crypto.PublicKey) *SealedEnvelope {
	sealed := &SealedEnvelope{
		Envelope:  act,
		srcPubkey: pubk,
	}
	return sealed
}

// AssembleSealedEnvelope assembles a SealedEnvelope use Envelope, Sender Address and Signature.
// This method should be only used in tests.
func AssembleSealedEnvelope(act Envelope, pk crypto.PublicKey, sig []byte) *SealedEnvelope {
	sealed := &SealedEnvelope{
		Envelope:  act,
		srcPubkey: pk,
		signature: sig,
	}
	return sealed
}

// CalculateIntrinsicGas returns the intrinsic gas of an action
func CalculateIntrinsicGas(baseIntrinsicGas uint64, payloadGas uint64, payloadSize uint64) (uint64, error) {
	if payloadGas == 0 && payloadSize == 0 {
		return baseIntrinsicGas, nil
	}
	if (math.MaxUint64-baseIntrinsicGas)/payloadGas < payloadSize {
		return 0, ErrInsufficientFunds
	}
	return payloadSize*payloadGas + baseIntrinsicGas, nil
}

// IsSystemAction determine whether input action belongs to system action
func IsSystemAction(act *SealedEnvelope) bool {
	switch act.Action().(type) {
	case *GrantReward, *PutPollResult:
		return true
	default:
		return false
	}
}

func CheckTransferAddress(act Action) error {
	switch act := act.(type) {
	case *Transfer:
		if _, err := address.FromString(act.recipient); err != nil {
			return errors.Wrapf(err, "invalid address %s", act.recipient)
		}
	default:
		// other actions have this check in their SanityCheck
	}
	return nil
}

func CheckSpecialAddress(act Action) bool {
	switch act := act.(type) {
	case *Transfer:
		return address.IsAddrV1Special(act.recipient)
	case *Execution:
		return address.IsAddrV1Special(act.contract)
	case *CandidateRegister:
		return address.IsAddrV1Special(act.rewardAddress.String()) ||
			address.IsAddrV1Special(act.operatorAddress.String()) ||
			(act.ownerAddress != nil && address.IsAddrV1Special(act.ownerAddress.String()))
	case *CandidateUpdate:
		return (act.rewardAddress != nil && address.IsAddrV1Special(act.rewardAddress.String())) ||
			(act.operatorAddress != nil && address.IsAddrV1Special(act.operatorAddress.String()))
	case *CandidateTransferOwnership:
		return address.IsAddrV1Special(act.newOwner.String())
	case *TransferStake:
		return address.IsAddrV1Special(act.voterAddress.String())
	case *ClaimFromRewardingFund:
		return act.address != nil && address.IsAddrV1Special(act.address.String())
	default:
		return false
	}
}
