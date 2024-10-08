// Copyright (c) 2024 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package db

import (
	"context"
	"syscall"

	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/iotexproject/iotex-core/db/batch"
	"github.com/iotexproject/iotex-core/pkg/lifecycle"
	"github.com/iotexproject/iotex-core/pkg/log"
)

// PebbleVersioned is KVStore implementation based on pebble DB
type PebbleVersioned struct {
	lifecycle.Readiness
	db     *pebble.DB
	path   string
	config Config
}

// NewPebbleDB creates a new PebbleDB instance
func NewPebbleVersioned(cfg Config) *PebbleVersioned {
	return &PebbleVersioned{
		db:     nil,
		path:   cfg.DbPath,
		config: cfg,
	}
}

// Start opens the DB (creates new file if not existing yet)
func (b *PebbleVersioned) Start(_ context.Context) error {
	comparer := pebble.DefaultComparer
	comparer.Split = func(a []byte) int {
		return prefixLength
	}
	db, err := pebble.Open(b.path, &pebble.Options{
		Comparer:           comparer,
		FormatMajorVersion: pebble.FormatPrePebblev1MarkedCompacted,
		ReadOnly:           b.config.ReadOnly,
	})
	if err != nil {
		return errors.Wrap(ErrIO, err.Error())
	}
	b.db = db
	return b.TurnOn()
}

// Stop closes the DB
func (b *PebbleVersioned) Stop(_ context.Context) error {
	if err := b.TurnOff(); err != nil {
		return err
	}
	if err := b.db.Close(); err != nil {
		return errors.Wrap(ErrIO, err.Error())
	}
	return nil
}

// Get retrieves a record
func (b *PebbleVersioned) Get(ns string, key []byte) ([]byte, error) {
	if !b.IsReady() {
		return nil, ErrDBNotStarted
	}
	v, closer, err := b.db.Get(nsKey(ns, key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, errors.Wrapf(ErrNotExist, "ns %s key = %x doesn't exist, %s", ns, key, err.Error())
		}
		return nil, err
	}
	val := make([]byte, len(v))
	copy(val, v)
	return val, closer.Close()
}

// Put inserts a <key, value> record
func (b *PebbleVersioned) Put(ns string, key, value []byte) (err error) {
	if !b.IsReady() {
		return ErrDBNotStarted
	}
	err = b.db.Set(nsKey(ns, key), value, nil)
	if err != nil {
		if errors.Is(err, syscall.ENOSPC) {
			log.L().Fatal("Failed to put db.", zap.Error(err))
		}
		err = errors.Wrap(ErrIO, err.Error())
	}
	return
}

// Delete deletes a record,if key is nil,this will delete the whole bucket
func (b *PebbleVersioned) Delete(ns string, key []byte) (err error) {
	if !b.IsReady() {
		return ErrDBNotStarted
	}
	if key == nil {
		panic("delete whole ns not supported by PebbleDB")
	}
	err = b.db.Delete(nsKey(ns, key), nil)
	if err != nil {
		if errors.Is(err, syscall.ENOSPC) {
			log.L().Fatal("Failed to delete db.", zap.Error(err))
		}
		err = errors.Wrap(ErrIO, err.Error())
	}
	return
}

// WriteBatch commits a batch
func (b *PebbleVersioned) WriteBatch(kvsb batch.KVStoreBatch) error {
	if !b.IsReady() {
		return ErrDBNotStarted
	}

	batch, err := b.dedup(kvsb)
	if err != nil {
		return nil
	}
	err = batch.Commit(nil)
	if err != nil {
		if errors.Is(err, syscall.ENOSPC) {
			log.L().Fatal("Failed to write batch db.", zap.Error(err))
		}
		err = errors.Wrap(ErrIO, err.Error())
	}
	return err
}

func (b *PebbleVersioned) dedup(kvsb batch.KVStoreBatch) (*pebble.Batch, error) {
	kvsb.Lock()
	defer kvsb.Unlock()

	type doubleKey struct {
		ns  string
		key string
	}
	// remove duplicate keys, only keep the last write for each key
	var (
		entryKeySet = make(map[doubleKey]struct{})
		ch          = b.db.NewBatch()
	)
	for i := kvsb.Size() - 1; i >= 0; i-- {
		write, e := kvsb.Entry(i)
		if e != nil {
			return nil, e
		}
		// only handle Put and Delete
		if write.WriteType() != batch.Put && write.WriteType() != batch.Delete {
			continue
		}
		key := write.Key()
		k := doubleKey{ns: write.Namespace(), key: string(key)}
		if _, ok := entryKeySet[k]; !ok {
			entryKeySet[k] = struct{}{}
			// add into batch
			if write.WriteType() == batch.Put {
				ch.Set(nsKey(write.Namespace(), key), write.Value(), nil)
			} else {
				ch.Delete(nsKey(write.Namespace(), key), nil)
			}
		}
	}
	return ch, nil
}
