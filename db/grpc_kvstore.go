// Copyright (c) 2024 IoTeX
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package db

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iotexproject/iotex-core/v2/db/batch"
	"github.com/iotexproject/iotex-core/v2/db/kvstorepb"
)

type GrpcKVStore struct {
	url      string
	insecure bool
	conn     *grpc.ClientConn
	client   kvstorepb.KVStoreServiceClient
}

func NewGrpcKVStore(url string, insecure bool) *GrpcKVStore {
	return &GrpcKVStore{
		url:      url,
		insecure: insecure,
	}
}

func (gkv *GrpcKVStore) Start(ctx context.Context) error {
	opts := []grpc.DialOption{}
	if gkv.insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(gkv.url, opts...)
	if err != nil {
		return err
	}
	gkv.conn = conn
	gkv.client = kvstorepb.NewKVStoreServiceClient(conn)
	return nil
}

func (gkv *GrpcKVStore) Stop(ctx context.Context) error {
	return gkv.conn.Close()
}

func (gkv *GrpcKVStore) Put(ns string, key, value []byte) error {
	panic("GrpcKVStore does not implement Put() method")
}

func (gkv *GrpcKVStore) Get(ns string, key []byte) ([]byte, error) {
	res, err := gkv.client.Get(context.Background(), &kvstorepb.GetRequest{
		Ns:  ns,
		Key: key,
	})
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func (gkv *GrpcKVStore) Delete(string, []byte) error {
	panic("GrpcKVStore does not implement Delete() method")
}

func (gkv *GrpcKVStore) WriteBatch(batch.KVStoreBatch) error {
	panic("GrpcKVStore does not implement WriteBatch() method")
}

func (gkv *GrpcKVStore) Filter(string, Condition, []byte, []byte) ([][]byte, [][]byte, error) {
	panic("GrpcKVStore does not implement Filter() method")
}
