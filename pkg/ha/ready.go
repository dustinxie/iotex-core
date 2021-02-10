// Copyright (c) 2021 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package ha

import (
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	_notReady = 0
	_starting = 1
	_ready    = 2
	_stopping = 3
)

var (
	// ErrServiceNotReady is the error that the service is not ready
	ErrServiceNotReady = errors.New("service is not ready")

	// ErrServiceStarted is the error that the service is already started
	ErrServiceStarted = errors.New("service already started")
)

// Readiness is a thread-safe struct to indicate a service's status
type Readiness struct {
	ready int32
}

// NotReady sets the service to not ready (initial state)
func (r *Readiness) NotReady() { atomic.SwapInt32(&r.ready, _notReady) }

// CannotStart checks whether the service can start or not
func (r *Readiness) CannotStart() bool {
	return !atomic.CompareAndSwapInt32(&r.ready, _notReady, _starting)
}

// Ready sets the service to ready (can accept service request)
func (r *Readiness) Ready() { atomic.SwapInt32(&r.ready, _ready) }

// CannotStop checks whether the service can stop or not
func (r *Readiness) CannotStop() bool {
	return !atomic.CompareAndSwapInt32(&r.ready, _ready, _stopping)
}

// IsNotReady returns whether the service is not ready (initial state)
func (r *Readiness) IsNotReady() bool {
	return atomic.LoadInt32(&r.ready) == _notReady
}

// IsNotReady returns whether the service is ready (can accept service request)
func (r *Readiness) IsReady() bool {
	return atomic.LoadInt32(&r.ready) == _ready
}
