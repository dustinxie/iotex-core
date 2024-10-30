// Copyright (c) 2021 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package db

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/stretchr/testify/require"
)

func TestPb(t *testing.T) {
	r := require.New(t)

	vn := &versionedNamespace{
		keyLen: 5}
	data := vn.serialize()
	r.Equal("1005", hex.EncodeToString(data))
	vn1, err := deserializeVersionedNamespace(data)
	r.NoError(err)
	r.Equal(vn, vn1)
}

func TestSlice(t *testing.T) {
	r := require.New(t)

	h := hash.Hash160b([]byte("testing slice"))
	fmt.Printf("hash = %p, %x\n", h[:4], h)
	key := append(h[:4], 0)
	fmt.Printf("key = %p, %x\n", key, key)
	r.Equal(5, len(key))
	fmt.Printf("hash = %p, %x\n", h[:4], h)
}
