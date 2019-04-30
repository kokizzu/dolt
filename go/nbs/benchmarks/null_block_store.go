// Copyright 2016 Attic Labs, Inc. All rights reserved.
// Licensed under the Apache License, version 2.0:
// http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"context"
	"github.com/attic-labs/noms/go/chunks"
	"github.com/attic-labs/noms/go/hash"
)

type nullBlockStore struct {
	bogus int32
}

func newNullBlockStore() chunks.ChunkStore {
	return nullBlockStore{}
}

func (nb nullBlockStore) Get(ctx context.Context, h hash.Hash) chunks.Chunk {
	panic("not impl")
}

func (nb nullBlockStore) GetMany(ctx context.Context, hashes hash.HashSet, foundChunks chan *chunks.Chunk) {
	panic("not impl")
}

func (nb nullBlockStore) Has(ctx context.Context, h hash.Hash) bool {
	panic("not impl")
}

func (nb nullBlockStore) HasMany(ctx context.Context, hashes hash.HashSet) (present hash.HashSet) {
	panic("not impl")
}

func (nb nullBlockStore) Put(ctx context.Context, c chunks.Chunk) {}

func (nb nullBlockStore) Version() string {
	panic("not impl")
}

func (nb nullBlockStore) Close() error {
	return nil
}

func (nb nullBlockStore) Rebase(ctx context.Context) {}

func (nb nullBlockStore) Stats() interface{} {
	return nil
}

func (nb nullBlockStore) StatsSummary() string {
	return "Unsupported"
}

func (nb nullBlockStore) Root(ctx context.Context) hash.Hash {
	return hash.Hash{}
}

func (nb nullBlockStore) Commit(ctx context.Context, current, last hash.Hash) bool {
	return true
}
