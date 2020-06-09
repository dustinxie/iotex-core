// Copyright (c) 2020 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package action

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	sigTests = []struct {
		signer    string
		recipient string
		msg       string
		hash      string
		sig       string
	}{
		// sample sig generated by keystore file at https://etherscan.io/verifySig/2059
		{
			"53fbc28faf9a52dfe5f591948a23189e900381b5",
			"io120au9ra0nffdle04jx2g5gccn6gq8qd4fy03l4",
			"{\"nonce\":\"136\",\"address\":\"io14s0vgnj0pjnazu4hsqlksdk7slah9vcfscn9ks\"}",
			"f93a97fae37fdadab6d49b74e3f3e4bee707ea2f007e08007bcc356cb283665b",
			"5595906a47dfc107a78cc48b500f89ab2dec545ba86578295aed4a260ce9a98b335924e86f683832e313f1a5dda7826d9b59caf40dd22ce92716420a367dfaec1c",
		},
		// sample sig created by Ledger Nano S at https://etherscan.io/verifySig/2060
		{
			"35cb41ec685a30bf5ebc3f935aebb1a8391e5fd6",
			"io1xh95rmrgtgct7h4u87f446a34qu3uh7kzqkx2a",
			"{\"nonce\":\"136\",\"address\":\"io14s0vgnj0pjnazu4hsqlksdk7slah9vcfscn9ks\"}",
			"f93a97fae37fdadab6d49b74e3f3e4bee707ea2f007e08007bcc356cb283665b",
			"e28fb729d1f8b841d36c2688571886a077e1f1529b239aa02d4ae1da84cdf5a37e7656e13d374d4fdf3010e75bcd3629a36765e941f2e427c8c4529c1f713caf01",
		},
		// sample sig https://etherscan.io/verifySig/2079
		{
			"ac1ec44e4f0ca7d172b7803f6836de87fb72b309",
			"io14s0vgnj0pjnazu4hsqlksdk7slah9vcfscn9ks",
			"{\"type\":\"Ethereum\",\"bucket\":47,\"nonce\":136,\"recipient\":\"io14s0vgnj0pjnazu4hsqlksdk7slah9vcfscn9ks\",\"reclaim\":\"This is to certify I am transferring the ownership of said bucket to said recipient on IoTeX blockchain\"}",
			"ce7937213ec491a14992345a3162556fff200dca972a317bc4cae88092e8a6f7",
			"7069f7159f484a73378f7992398d924700894f0640cff9c3eb980a327082ed91283d4d1a0a9eba572b022dd369ac73bc44a953db4849405507819a5ca5d0cb5e1c",
		},
	}
)

func TestVerifyEccSig(t *testing.T) {
	r := require.New(t)

	for _, v := range sigTests {
		sig, _ := hex.DecodeString(v.sig)
		pk, err := RecoverPubkeyFromEccSig("Ethereum", []byte(v.msg), sig)
		r.NoError(err)
		r.Equal(v.signer, hex.EncodeToString(pk.Hash()))
		h, _ := hex.DecodeString(v.hash)
		r.True(pk.Verify(h, sig))
	}

	// test with modified signature
	for _, v := range sigTests {
		sig, _ := hex.DecodeString(v.sig)
		sig[rand.Intn(len(sig))]++
		pk, err := RecoverPubkeyFromEccSig("Ethereum", []byte(v.msg), sig)
		if err == nil {
			r.NotEqual(v.signer, hex.EncodeToString(pk.Hash()))
		}
	}
}

func TestConsignmentTransfer(t *testing.T) {
	r := require.New(t)

	// generate payload from tests
	v := sigTests[2]
	msg := ConsignMsg{
		Type:      "Ethereum",
		Index:     47,
		Nonce:     136,
		Recipient: v.recipient,
		Reclaim:   _reclaim,
	}
	b, err := json.Marshal(msg)
	r.NoError(err)
	r.Equal(v.msg, string(b))

	c := &ConsignJSON{
		Msg: v.msg,
		Sig: v.sig,
	}
	b, err = json.Marshal(c)
	r.NoError(err)

	// process the payload as a consignment transfer
	con, err := NewConsignment(b)
	r.NoError(err)
	r.Equal(v.signer, hex.EncodeToString(con.Transferor().Bytes()))
	r.Equal(v.recipient, con.Transferee().String())
	r.EqualValues(47, con.AssetID())
	r.EqualValues(136, con.TransfereeNonce())
}