package state

import (
	"encoding/binary"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	"github.com/onflow/flow-go/model/flow"
)

type Height struct {
	core *Core
}

func (h *Height) ForBlock(blockID flow.Identifier) (uint64, error) {
	key := make([]byte, 1+len(blockID))
	key[0] = prefixBlockIndex
	copy(key[1:], blockID[:])
	var height uint64
	err := h.core.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return fmt.Errorf("could not retrieve block index: %w", err)
		}
		_ = item.Value(func(val []byte) error {
			height = binary.BigEndian.Uint64(val)
			return nil
		})
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("could not look up block: %w", err)
	}
	return height, nil
}

func (h *Height) ForCommit(commit flow.StateCommitment) (uint64, error) {
	key := make([]byte, 1+len(commit))
	key[0] = prefixCommitIndex
	copy(key[1:], commit[:])
	var height uint64
	err := h.core.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return fmt.Errorf("could not retrieve commit index: %w", err)
		}
		_ = item.Value(func(val []byte) error {
			height = binary.BigEndian.Uint64(val)
			return nil
		})
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("could not look up commit: %w", err)
	}
	return height, nil
}
