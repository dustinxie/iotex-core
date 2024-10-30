// Copyright (c) 2024 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package db

import (
	"bytes"
	"context"
	"math"
	"syscall"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-core/v2/db/batch"
	"github.com/iotexproject/iotex-core/v2/pkg/lifecycle"
	"github.com/iotexproject/iotex-core/v2/pkg/log"
)

const (
	_nsHashLen    = 4  // chance of namespace collision = 1/2^64
	_keyPrefixLen = 12 // first 12-bytes of key
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
		return _nsHashLen + _keyPrefixLen
	}
	db, err := pebble.Open(b.path, &pebble.Options{
		Comparer:           comparer,
		FormatMajorVersion: pebble.FormatPrePebblev1MarkedCompacted,
		ReadOnly:           b.config.ReadOnly,
		Levels:             []pebble.LevelOptions{{FilterPolicy: bloom.FilterPolicy(10)}},
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
func (b *PebbleVersioned) getNV(ns string, key []byte) ([]byte, error) {
	if !b.IsReady() {
		return nil, ErrDBNotStarted
	}
	v, closer, err := b.db.Get(concatNsKey(ns, key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, errors.Wrapf(ErrNotExist, "PebbleVersioned.getNV(): ns %s key = %x doesn't exist, %s", ns, key, err.Error())
		}
		return nil, err
	}
	val := make([]byte, len(v))
	copy(val, v)
	return val, closer.Close()
}

// Put inserts a <key, value> record
func (b *PebbleVersioned) Put(version uint64, ns string, key, value []byte) (err error) {
	if !b.IsReady() {
		return ErrDBNotStarted
	}
	// check namespace
	vn, err := b.checkNamespace(ns)
	if err != nil {
		return err
	}
	buf := b.db.NewBatch()
	if vn == nil {
		// namespace not yet created
		buf.Set(concatNsKey(ns, _minKey), (&versionedNamespace{
			keyLen: uint32(len(key)),
		}).serialize(), nil)
	} else {
		if len(key) != int(vn.keyLen) {
			return errors.Wrapf(ErrInvalid, "PebbleVersioned.Put(): invalid key length, expecting %d, got %d", vn.keyLen, len(key))
		}
		last, _, err := b.get(math.MaxUint64, ns, key)
		if !isNotExist(err) && version < last {
			// not allowed to perform write on an earlier version
			return ErrInvalid
		}
		buf.Delete(keyForDelete(concatNsKey(ns, key), version), nil)
	}
	buf.Set(keyForWrite(concatNsKey(ns, key), version), value, nil)

	if err = buf.Commit(nil); err != nil {
		if errors.Is(err, syscall.ENOSPC) {
			log.L().Fatal("PebbleVersioned.Put(): failed to write to db", zap.Error(err))
		}
		err = errors.Wrap(ErrIO, err.Error())
	}
	return err
}

func (b *PebbleVersioned) get(version uint64, ns string, key []byte) (uint64, []byte, error) {
	var (
		last     uint64
		isDelete bool
		value    []byte

		min = keyForDelete(key, 0)
		key = keyForWrite(key, version)
	)

	iter, err := b.db.NewIter(&pebble.IterOptions{})
	if err != nil {
		return 0, nil, errors.Wrap(err, "PebbleVersioned.get(): failed to create iterator")
	}
	if !iter.SeekPrefixGE(key) {
		// no valid key exists
		return 0, nil, errors.Wrapf(ErrNotExist, "PebbleVersioned.get(): failed to seek key = %x\n")
	}
	k, v := iter.Key(), iter.Value()

	err := b.db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(ns))
		var (
			c    = bucket.Cursor()
			k, v = c.Seek(key)
		)
		if k == nil || bytes.Compare(k, key) == 1 {
			k, v = c.Prev()
			if k == nil || bytes.Compare(k, min) <= 0 {
				// cursor is at the beginning/end of the bucket or smaller than minimum key
				return ErrNotExist
			}
		}
		isDelete, last = parseKey(k)
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return last, nil, err
	}
	if isDelete {
		return last, nil, ErrDeleted
	}
	return last, value, nil
}

// Delete deletes a record,if key is nil,this will delete the whole bucket
func (b *PebbleVersioned) Delete(ns string, key []byte) error {
	if !b.IsReady() {
		return ErrDBNotStarted
	}
	if key == nil {
		return errors.Wrap(ErrNotSupported, "PebbleVersioned.Delete(): delete whole ns not supported")
	}
	err := b.db.Delete(concatNsKey(ns, key), nil)
	if err != nil {
		if errors.Is(err, syscall.ENOSPC) {
			log.L().Fatal("Failed to delete db.", zap.Error(err))
		}
		err = errors.Wrap(ErrIO, err.Error())
	}
	return err
}

// Version returns the key's most recent version
func (b *PebbleVersioned) Version(ns string, key []byte) (uint64, error) {
	if !b.IsReady() {
		return 0, ErrDBNotStarted
	}
	// check key's metadata
	if err := b.checkNamespaceAndKey(ns, key); err != nil {
		return 0, err
	}
	last, _, err := b.get(math.MaxUint64, ns, key)
	if isNotExist(err) {
		// key not yet written
		err = errors.Wrapf(ErrNotExist, "PebbleVersioned.Version(): key = %x doesn't exist", key)
	}
	return last, err
}

// CommitToDB write a batch to DB, where the batch can contain keys for
// both versioned and non-versioned namespace
func (b *PebbleVersioned) CommitToDB(version uint64, vns map[string]bool, kvsb batch.KVStoreBatch) error {
	vnsize, ve, nve, err := dedup(vns, kvsb)
	if err != nil {
		return errors.Wrap(err, "PebbleVersioned.CommitToDB(): failed to dedup batch")
	}
	return b.commitToDB(version, vnsize, ve, nve)
}

func (b *PebbleVersioned) commitToDB(version uint64, vnsize map[string]int, ve, nve []*batch.WriteInfo) error {
	return nil
}

func concatNsKey(ns string, key []byte) []byte {
	h := hash.Hash160b([]byte(ns))
	return append(h[:_nsHashLen], key...)
}

func (b *PebbleVersioned) checkNamespace(ns string) (*versionedNamespace, error) {
	data, err := b.getNV(ns, _minKey)
	switch errors.Cause(err) {
	case nil:
		vn, err := deserializeVersionedNamespace(data)
		if err != nil {
			return nil, err
		}
		return vn, nil
	case ErrNotExist, ErrBucketNotExist:
		return nil, nil
	default:
		return nil, err
	}
}

func (b *PebbleVersioned) checkNamespaceAndKey(ns string, key []byte) error {
	vn, err := b.checkNamespace(ns)
	if err != nil {
		return err
	}
	if vn == nil {
		return errors.Wrapf(ErrNotExist, "namespace = %x doesn't exist", ns)
	}
	if len(key) != int(vn.keyLen) {
		return errors.Wrapf(ErrInvalid, "invalid key length, expecting %d, got %d", vn.keyLen, len(key))
	}
	return nil
}
