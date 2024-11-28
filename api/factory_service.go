// Copyright (c) 2024 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package api

import (
	"context"

	"github.com/iotexproject/iotex-core/v2/db"
	"github.com/iotexproject/iotex-core/v2/db/kvstorepb"
)

type factoryService struct {
	kvs db.KVStore
}

func newFactoryService(kvs db.KVStore) *factoryService {
	return &factoryService{
		kvs: kvs,
	}
}

func (service *factoryService) Get(_ context.Context, request *kvstorepb.GetRequest) (*kvstorepb.GetResponse, error) {
	v, err := service.kvs.Get(request.Ns, request.Key)
	if err != nil {
		return nil, err
	}
	return &kvstorepb.GetResponse{
		Value: v,
	}, nil
}
