// Copyright (c) 2021 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package ha

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReady(t *testing.T) {
	r := require.New(t)

	ready := Readiness{}
	r.True(ready.IsNotReady())
	r.False(ready.IsReady())
	r.False(ready.CannotStart())
	r.True(ready.CannotStop())

	// it is now starting
	r.False(ready.IsNotReady())
	r.False(ready.IsReady())
	r.True(ready.CannotStart())
	r.True(ready.CannotStop())

	// it is now ready
	ready.Ready()
	r.False(ready.IsNotReady())
	r.True(ready.IsReady())
	r.True(ready.CannotStart())
	r.False(ready.CannotStop())

	// it is now stopping
	r.False(ready.IsNotReady())
	r.False(ready.IsReady())
	r.True(ready.CannotStart())
	r.True(ready.CannotStop())

	// go back to not ready
	ready.NotReady()
	r.True(ready.IsNotReady())
	r.False(ready.IsReady())
	r.True(ready.CannotStop())
}
