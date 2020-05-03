// Copyright (c) 2020 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package staking

import (
	"context"
	"math/big"
	"time"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-core/action"
	"github.com/iotexproject/iotex-core/action/protocol"
	accountutil "github.com/iotexproject/iotex-core/action/protocol/account/util"
	"github.com/iotexproject/iotex-core/pkg/util/byteutil"
	"github.com/iotexproject/iotex-core/state"
)

const (
	// HandleCreateStake is the handler name of createStake
	HandleCreateStake = "createStake"
	// HandleUnstake is the handler name of unstake
	HandleUnstake = "unstake"
	// HandleWithdrawStake is the handler name of withdrawStake
	HandleWithdrawStake = "withdrawStake"
	// HandleChangeCandidate is the handler name of changeCandidate
	HandleChangeCandidate = "changeCandidate"
	// HandleTransferStake is the handler name of transferStake
	HandleTransferStake = "transferStake"
	// HandleDepositToStake is the handler name of depositToStake
	HandleDepositToStake = "depositToStake"
	// HandleRestake is the handler name of restake
	HandleRestake = "restake"
	// HandleCandidateRegister is the handler name of candidateRegister
	HandleCandidateRegister = "candidateRegister"
	// HandleCandidateUpdate is the handler name of candidateUpdate
	HandleCandidateUpdate = "candidateUpdate"
)

// Errors and vars
var (
	ErrNilParameters = errors.New("parameter is nil")
	errCandNotExist  = &handleError{
		err:           ErrInvalidOwner,
		failureStatus: iotextypes.ReceiptStatus_ErrCandidateNotExist,
	}
	handleSuccess = &handleError{
		err:           nil,
		failureStatus: iotextypes.ReceiptStatus_Success,
	}
)

type handleError struct {
	err           error
	failureStatus iotextypes.ReceiptStatus
}

func (h *handleError) Error() error {
	return h.err
}

func (h *handleError) ReceiptStatus() uint64 {
	return uint64(h.failureStatus)
}

func (h *handleError) CanContinueToCommitBlock() bool {
	return h.failureStatus != iotextypes.ReceiptStatus_ErrWriteAccount &&
		h.failureStatus != iotextypes.ReceiptStatus_ErrWriteBucket &&
		h.failureStatus != iotextypes.ReceiptStatus_ErrWriteCandidate
}

func (p *Protocol) handleCreateStake(ctx context.Context, act *action.CreateStake, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleCreateStake))}
	actionCtx := protocol.MustGetActionCtx(ctx)
	blkCtx := protocol.MustGetBlockCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, act.Amount()); fetchErr != nil {
		return topics, fetchErr
	}

	// Create new bucket and bucket index
	candidate := csm.GetByName(act.Candidate())
	if candidate == nil {
		return topics, errCandNotExist
	}
	bucket := NewVoteBucket(candidate.Owner, actionCtx.Caller, act.Amount(), act.Duration(), blkCtx.BlockTimeStamp, act.AutoStake())
	bucketIdx, err := putBucketAndIndex(csm, bucket)
	if err != nil {
		return topics, &handleError{
			err:           err,
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucketIdx)), hash.BytesToHash256(candidate.Owner.Bytes()))

	// update candidate
	if updateError := updateCandidate(csm, candidate, p.calculateVoteWeight(bucket, false), _addVote); updateError != nil {
		return topics, updateError
	}

	// update staker balance
	if updateError := updateCaller(csm, actionCtx.Caller, act.Amount(), _subBalance); updateError != nil {
		return topics, updateError
	}
	return topics, handleSuccess
}

func (p *Protocol) handleUnstake(ctx context.Context, act *action.Unstake, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleUnstake))}
	actionCtx := protocol.MustGetActionCtx(ctx)
	blkCtx := protocol.MustGetBlockCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, big.NewInt(0)); fetchErr != nil {
		return topics, fetchErr
	}

	bucket, fetchErr := p.fetchBucket(csm, actionCtx.Caller, act.BucketIndex(), true, true)
	if fetchErr != nil {
		return topics, fetchErr
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucket.Index)), hash.BytesToHash256(bucket.Candidate.Bytes()))

	candidate := csm.GetByOwner(bucket.Candidate)
	if candidate == nil {
		return nil, errCandNotExist
	}

	if bucket.AutoStake {
		return topics, &handleError{
			err:           errors.New("AutoStake should be disabled first in order to unstake"),
			failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketType,
		}
	}

	if blkCtx.BlockTimeStamp.Before(bucket.StakeStartTime.Add(bucket.StakedDuration)) {
		return topics, &handleError{
			err:           errors.New("bucket is not ready to be unstaked"),
			failureStatus: iotextypes.ReceiptStatus_ErrUnstakeBeforeMaturity,
		}
	}

	// update bucket
	bucket.UnstakeStartTime = blkCtx.BlockTimeStamp.UTC()
	if err := updateBucket(csm, act.BucketIndex(), bucket); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to update bucket for voter %s", bucket.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}

	// clear candidate's self stake if the bucket is self staking
	if csm.ContainsSelfStakingBucket(act.BucketIndex()) {
		candidate.SelfStake = big.NewInt(0)
	}
	if updateError := updateCandidate(csm, candidate, p.calculateVoteWeight(
		bucket, csm.ContainsSelfStakingBucket(act.BucketIndex())), _subVote); updateError != nil {
		return topics, updateError
	}
	return topics, handleSuccess
}

func (p *Protocol) handleWithdrawStake(ctx context.Context, act *action.WithdrawStake, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleWithdrawStake))}
	actionCtx := protocol.MustGetActionCtx(ctx)
	blkCtx := protocol.MustGetBlockCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, big.NewInt(0)); fetchErr != nil {
		return topics, fetchErr
	}

	bucket, fetchErr := p.fetchBucket(csm, actionCtx.Caller, act.BucketIndex(), true, true)
	if fetchErr != nil {
		return topics, fetchErr
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucket.Index)), hash.BytesToHash256(bucket.Candidate.Bytes()))

	// check unstake time
	if bucket.UnstakeStartTime.Unix() == 0 {
		return topics, &handleError{
			err:           errors.New("bucket has not been unstaked"),
			failureStatus: iotextypes.ReceiptStatus_ErrWithdrawBeforeUnstake,
		}
	}

	if blkCtx.BlockTimeStamp.Before(bucket.UnstakeStartTime.Add(p.config.WithdrawWaitingPeriod)) {
		return topics, &handleError{
			err:           errors.New("stake is not ready to withdraw"),
			failureStatus: iotextypes.ReceiptStatus_ErrWithdrawBeforeMaturity,
		}
	}

	// delete bucket and bucket index
	if err := delBucketAndIndex(csm, bucket.Owner, bucket.Candidate, act.BucketIndex()); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to delete bucket for candidate %s", bucket.Candidate.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}

	// update withdrawer balance
	if updateError := updateCaller(csm, actionCtx.Caller, bucket.StakedAmount, _addBalance); updateError != nil {
		return topics, updateError
	}
	return topics, handleSuccess
}

func (p *Protocol) handleChangeCandidate(ctx context.Context, act *action.ChangeCandidate, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleChangeCandidate))}
	actionCtx := protocol.MustGetActionCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, big.NewInt(0)); fetchErr != nil {
		return topics, fetchErr
	}

	candidate := csm.GetByName(act.Candidate())
	if candidate == nil {
		return topics, errCandNotExist
	}

	bucket, fetchErr := p.fetchBucket(csm, actionCtx.Caller, act.BucketIndex(), true, false)
	if fetchErr != nil {
		return topics, fetchErr
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucket.Index)),
		hash.BytesToHash256(bucket.Candidate.Bytes()),
		hash.BytesToHash256(candidate.Owner.Bytes()))

	prevCandidate := csm.GetByOwner(bucket.Candidate)
	if prevCandidate == nil {
		return topics, errCandNotExist
	}

	// update bucket index
	if err := delCandBucketIndex(csm, bucket.Candidate, act.BucketIndex()); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to delete candidate bucket index for candidate %s", bucket.Candidate.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}
	if err := putCandBucketIndex(csm, candidate.Owner, act.BucketIndex()); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to put candidate bucket index for candidate %s", candidate.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}
	// update bucket
	bucket.Candidate = candidate.Owner
	if err := updateBucket(csm, act.BucketIndex(), bucket); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to update bucket for voter %s", bucket.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}

	// update previous candidate
	weightedVotes := p.calculateVoteWeight(bucket, false)
	if updateError := updateCandidate(csm, prevCandidate, weightedVotes, _subVote); updateError != nil {
		return topics, updateError
	}

	// update current candidate
	if updateError := updateCandidate(csm, candidate, weightedVotes, _addVote); updateError != nil {
		return topics, updateError
	}
	return topics, handleSuccess
}

func (p *Protocol) handleTransferStake(ctx context.Context, act *action.TransferStake, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleTransferStake))}
	actionCtx := protocol.MustGetActionCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, big.NewInt(0)); fetchErr != nil {
		return topics, fetchErr
	}

	bucket, fetchErr := p.fetchBucket(csm, actionCtx.Caller, act.BucketIndex(), true, false)
	if fetchErr != nil {
		return topics, fetchErr
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucket.Index)),
		hash.BytesToHash256(act.VoterAddress().Bytes()),
		hash.BytesToHash256(bucket.Candidate.Bytes()))

	// update bucket index
	if err := delVoterBucketIndex(csm, bucket.Owner, act.BucketIndex()); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to delete voter bucket index for voter %s", bucket.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}
	if err := putVoterBucketIndex(csm, act.VoterAddress(), act.BucketIndex()); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to put candidate bucket index for voter %s", act.VoterAddress().String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}

	// update bucket
	bucket.Owner = act.VoterAddress()
	if err := updateBucket(csm, act.BucketIndex(), bucket); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to update bucket for voter %s", bucket.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}
	return topics, handleSuccess
}

func (p *Protocol) handleDepositToStake(ctx context.Context, act *action.DepositToStake, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleDepositToStake))}
	actionCtx := protocol.MustGetActionCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, act.Amount()); fetchErr != nil {
		return topics, fetchErr
	}

	bucket, fetchErr := p.fetchBucket(csm, actionCtx.Caller, act.BucketIndex(), false, true)
	if fetchErr != nil {
		return topics, fetchErr
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucket.Index)),
		hash.BytesToHash256(bucket.Owner.Bytes()),
		hash.BytesToHash256(bucket.Candidate.Bytes()))
	if !bucket.AutoStake {
		return topics, &handleError{
			err:           errors.New("deposit is only allowed on auto-stake bucket"),
			failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketType,
		}
	}
	candidate := csm.GetByOwner(bucket.Candidate)
	if candidate == nil {
		return topics, errCandNotExist
	}

	prevWeightedVotes := p.calculateVoteWeight(bucket, csm.ContainsSelfStakingBucket(act.BucketIndex()))
	// update bucket
	bucket.StakedAmount.Add(bucket.StakedAmount, act.Amount())
	if err := updateBucket(csm, act.BucketIndex(), bucket); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to update bucket for voter %s", bucket.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}

	// update candidate
	if err := candidate.UpdateVote(prevWeightedVotes, _subVote); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to subtract vote for candidate %s", bucket.Candidate.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrNotEnoughBalance,
		}
	}
	if csm.ContainsSelfStakingBucket(act.BucketIndex()) {
		if err := candidate.AddSelfStake(act.Amount()); err != nil {
			return topics, &handleError{
				err:           errors.Wrapf(err, "failed to add self stake for candidate %s", candidate.Owner.String()),
				failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketAmount,
			}
		}
	}
	if updateError := updateCandidate(csm, candidate, p.calculateVoteWeight(
		bucket, csm.ContainsSelfStakingBucket(act.BucketIndex())), _addVote); updateError != nil {
		return topics, updateError
	}

	// update depositor balance
	if updateError := updateCaller(csm, actionCtx.Caller, act.Amount(), _subBalance); updateError != nil {
		return topics, updateError
	}
	return topics, handleSuccess
}

func (p *Protocol) handleRestake(ctx context.Context, act *action.Restake, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleRestake))}
	actionCtx := protocol.MustGetActionCtx(ctx)
	blkCtx := protocol.MustGetBlockCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, big.NewInt(0)); fetchErr != nil {
		return topics, fetchErr
	}

	bucket, fetchErr := p.fetchBucket(csm, actionCtx.Caller, act.BucketIndex(), true, true)
	if fetchErr != nil {
		return topics, fetchErr
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucket.Index)), hash.BytesToHash256(bucket.Candidate.Bytes()))

	candidate := csm.GetByOwner(bucket.Candidate)
	if candidate == nil {
		return topics, errCandNotExist
	}

	prevWeightedVotes := p.calculateVoteWeight(bucket, csm.ContainsSelfStakingBucket(act.BucketIndex()))
	// update bucket
	actDuration := time.Duration(act.Duration()) * 24 * time.Hour
	if bucket.StakedDuration.Hours() > actDuration.Hours() {
		// in case of reducing the duration
		if bucket.AutoStake {
			// if auto-stake on, user can't reduce duration
			return topics, &handleError{
				err:           errors.New("AutoStake should be disabled first in order to reduce duration"),
				failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketType,
			}
		} else if blkCtx.BlockTimeStamp.Before(bucket.StakeStartTime.Add(bucket.StakedDuration)) {
			// if auto-stake off and maturity is not enough
			return topics, &handleError{
				err:           errors.New("bucket is not ready to be able to reduce duration"),
				failureStatus: iotextypes.ReceiptStatus_ErrReduceDurationBeforeMaturity,
			}
		}
	}
	bucket.StakedDuration = actDuration
	bucket.StakeStartTime = blkCtx.BlockTimeStamp.UTC()
	bucket.AutoStake = act.AutoStake()
	if err := updateBucket(csm, act.BucketIndex(), bucket); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to update bucket for voter %s", bucket.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}

	// update candidate
	if err := candidate.UpdateVote(prevWeightedVotes, _subVote); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to subtract vote for candidate %s", bucket.Candidate.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrNotEnoughBalance,
		}
	}
	if updateError := updateCandidate(csm, candidate, p.calculateVoteWeight(
		bucket, csm.ContainsSelfStakingBucket(act.BucketIndex())), _addVote); updateError != nil {
		return topics, updateError
	}
	return topics, handleSuccess
}

func (p *Protocol) handleCandidateRegister(ctx context.Context, act *action.CandidateRegister, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleCandidateRegister))}
	actCtx := protocol.MustGetActionCtx(ctx)
	blkCtx := protocol.MustGetBlockCtx(ctx)

	registrationFee := new(big.Int).Set(p.config.RegistrationConsts.Fee)

	if fetchErr := fetchCaller(ctx, csm, new(big.Int).Add(act.Amount(), registrationFee)); fetchErr != nil {
		return topics, fetchErr
	}

	owner := actCtx.Caller
	if act.OwnerAddress() != nil {
		owner = act.OwnerAddress()
	}

	c := csm.GetByOwner(owner)
	ownerExist := c != nil
	// cannot collide with existing owner (with selfstake != 0)
	if ownerExist && c.SelfStake.Cmp(big.NewInt(0)) != 0 {
		return topics, &handleError{
			err:           ErrInvalidOwner,
			failureStatus: iotextypes.ReceiptStatus_ErrCandidateAlreadyExist,
		}
	}
	// cannot collide with existing name
	if csm.ContainsName(act.Name()) && (!ownerExist || act.Name() != c.Name) {
		return topics, &handleError{
			err:           ErrInvalidCanName,
			failureStatus: iotextypes.ReceiptStatus_ErrCandidateConflict,
		}
	}
	// cannot collide with existing operator address
	if csm.ContainsOperator(act.OperatorAddress()) &&
		(!ownerExist || !address.Equal(act.OperatorAddress(), c.Operator)) {
		return topics, &handleError{
			err:           ErrInvalidOperator,
			failureStatus: iotextypes.ReceiptStatus_ErrCandidateConflict,
		}
	}

	bucket := NewVoteBucket(owner, owner, act.Amount(), act.Duration(), blkCtx.BlockTimeStamp, act.AutoStake())
	bucketIdx, err := putBucketAndIndex(csm, bucket)
	if err != nil {
		return topics, &handleError{
			err:           err,
			failureStatus: iotextypes.ReceiptStatus_ErrWriteBucket,
		}
	}
	topics = append(topics, hash.BytesToHash256(byteutil.Uint64ToBytesBigEndian(bucketIdx)), hash.BytesToHash256(owner.Bytes()))

	c = &Candidate{
		Owner:              owner,
		Operator:           act.OperatorAddress(),
		Reward:             act.RewardAddress(),
		Name:               act.Name(),
		Votes:              p.calculateVoteWeight(bucket, true),
		SelfStakeBucketIdx: bucketIdx,
		SelfStake:          act.Amount(),
	}

	if err := csm.Upsert(c); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to put candidate %s", owner.String()),
			failureStatus: csmErrorToReceiptStatus(err),
		}
	}

	// update caller balance
	if updateError := updateCaller(csm, actCtx.Caller, act.Amount(), _subBalance); updateError != nil {
		return topics, updateError
	}

	// put registrationFee to reward pool
	if err := p.depositGas(ctx, csm, registrationFee); err != nil {
		return topics, &handleError{
			err:           errors.Wrap(err, "failed to deposit gas"),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteAccount,
		}
	}
	return topics, handleSuccess
}

func (p *Protocol) handleCandidateUpdate(ctx context.Context, act *action.CandidateUpdate, csm CandidateStateManager,
) (action.Topics, protocol.HandleError) {
	topics := action.Topics{hash.BytesToHash256([]byte(HandleCandidateUpdate))}
	actCtx := protocol.MustGetActionCtx(ctx)

	if fetchErr := fetchCaller(ctx, csm, big.NewInt(0)); fetchErr != nil {
		return topics, fetchErr
	}

	// only owner can update candidate
	c := csm.GetByOwner(actCtx.Caller)
	if c == nil {
		return topics, errCandNotExist
	}

	if len(act.Name()) != 0 {
		c.Name = act.Name()
	}

	if act.OperatorAddress() != nil {
		c.Operator = act.OperatorAddress()
	}

	if act.RewardAddress() != nil {
		c.Reward = act.RewardAddress()
	}
	topics = append(topics, hash.BytesToHash256(c.Owner.Bytes()))

	if err := csm.Upsert(c); err != nil {
		return topics, &handleError{
			err:           errors.Wrapf(err, "failed to put candidate %s", c.Owner.String()),
			failureStatus: csmErrorToReceiptStatus(err),
		}
	}
	return topics, handleSuccess
}

func updateCandidate(csm CandidateStateManager, candidate *Candidate, vote *big.Int, update updateAmountType) protocol.HandleError {
	if err := candidate.UpdateVote(vote, update); err != nil {
		return &handleError{
			err:           errors.Wrapf(err, "failed to update vote for candidate %s", candidate.Owner.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketAmount,
		}
	}
	if err := csm.Upsert(candidate); err != nil {
		return &handleError{
			err:           errors.Wrapf(err, "failed to put candidate %s", candidate.Owner.String()),
			failureStatus: csmErrorToReceiptStatus(err),
		}
	}
	return nil
}

func updateCaller(csm CandidateStateManager, caller address.Address, balance *big.Int, update updateAmountType) protocol.HandleError {
	acc, _ := accountutil.LoadAccount(csm, hash.BytesToHash160(caller.Bytes()))
	switch update {
	case _addBalance:
		if err := acc.AddBalance(balance); err != nil {
			return &handleError{
				err:           errors.Wrapf(err, "failed to update the balance of caller %s", caller.String()),
				failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketAmount,
			}
		}
	case _subBalance:
		if err := acc.SubBalance(balance); err != nil {
			return &handleError{
				err:           errors.Wrapf(err, "failed to update the balance of caller %s", caller.String()),
				failureStatus: iotextypes.ReceiptStatus_ErrNotEnoughBalance,
			}
		}
	default:
		panic("wrong update request")
	}
	// put updated caller's account state
	if err := accountutil.StoreAccount(csm, caller.String(), acc); err != nil {
		return &handleError{
			err:           errors.Wrapf(err, "failed to store account %s", caller.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrWriteAccount,
		}
	}
	return nil
}

func (p *Protocol) fetchBucket(
	sr CandidateStateManager,
	caller address.Address,
	index uint64,
	checkOwner bool,
	allowSelfStaking bool,
) (*VoteBucket, protocol.HandleError) {
	bucket, err := getBucket(sr, index)
	if err != nil {
		fetchErr := &handleError{
			err:           errors.Wrapf(err, "failed to fetch bucket by index %d", index),
			failureStatus: iotextypes.ReceiptStatus_Failure,
		}
		if errors.Cause(err) == state.ErrStateNotExist {
			fetchErr.failureStatus = iotextypes.ReceiptStatus_ErrInvalidBucketIndex
		}
		return nil, fetchErr
	}
	if checkOwner && !address.Equal(bucket.Owner, caller) {
		return nil, &handleError{
			err:           errors.New("bucket owner does not match action caller"),
			failureStatus: iotextypes.ReceiptStatus_ErrUnauthorizedOperator,
		}
	}
	if !allowSelfStaking && sr.ContainsSelfStakingBucket(index) {
		return nil, &handleError{
			err:           errors.New("self staking bucket cannot be processed"),
			failureStatus: iotextypes.ReceiptStatus_ErrInvalidBucketType,
		}
	}
	return bucket, nil
}

func putBucketAndIndex(sm protocol.StateManager, bucket *VoteBucket) (uint64, error) {
	index, err := putBucket(sm, bucket)
	if err != nil {
		return 0, errors.Wrap(err, "failed to put bucket")
	}

	if err := putVoterBucketIndex(sm, bucket.Owner, index); err != nil {
		return 0, errors.Wrap(err, "failed to put bucket index")
	}

	if err := putCandBucketIndex(sm, bucket.Candidate, index); err != nil {
		return 0, errors.Wrap(err, "failed to put candidate index")
	}
	return index, nil
}

func delBucketAndIndex(sm protocol.StateManager, owner, cand address.Address, index uint64) error {
	if err := delBucket(sm, index); err != nil {
		return errors.Wrap(err, "failed to delete bucket")
	}

	if err := delVoterBucketIndex(sm, owner, index); err != nil {
		return errors.Wrap(err, "failed to delete bucket index")
	}

	if err := delCandBucketIndex(sm, cand, index); err != nil {
		return errors.Wrap(err, "failed to delete candidate index")
	}
	return nil
}

func fetchCaller(ctx context.Context, sr protocol.StateReader, amount *big.Int) protocol.HandleError {
	actionCtx := protocol.MustGetActionCtx(ctx)

	caller, err := accountutil.LoadAccount(sr, hash.BytesToHash160(actionCtx.Caller.Bytes()))
	if err != nil {
		return &handleError{
			err:           errors.Wrapf(err, "failed to load the account of caller %s", actionCtx.Caller.String()),
			failureStatus: iotextypes.ReceiptStatus_Failure,
		}
	}
	gasFee := big.NewInt(0).Mul(actionCtx.GasPrice, big.NewInt(0).SetUint64(actionCtx.IntrinsicGas))
	// check caller's balance
	if gasFee.Add(amount, gasFee).Cmp(caller.Balance) == 1 {
		return &handleError{
			err:           errors.Wrapf(state.ErrNotEnoughBalance, "caller %s balance not enough", actionCtx.Caller.String()),
			failureStatus: iotextypes.ReceiptStatus_ErrNotEnoughBalance,
		}
	}
	return nil
}

func csmErrorToReceiptStatus(err error) iotextypes.ReceiptStatus {
	switch errors.Cause(err) {
	case ErrInvalidCanName:
		return iotextypes.ReceiptStatus_ErrCandidateConflict
	case ErrInvalidOperator:
		return iotextypes.ReceiptStatus_ErrCandidateConflict
	case ErrInvalidSelfStkIndex:
		return iotextypes.ReceiptStatus_ErrCandidateConflict
	case ErrInvalidAmount:
		return iotextypes.ReceiptStatus_ErrCandidateNotExist
	case ErrInvalidOwner:
		return iotextypes.ReceiptStatus_ErrCandidateNotExist
	case ErrInvalidReward:
		return iotextypes.ReceiptStatus_ErrCandidateNotExist
	default:
		return iotextypes.ReceiptStatus_ErrWriteCandidate
	}
}
