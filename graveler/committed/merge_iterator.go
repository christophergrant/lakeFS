package committed

import (
	"bytes"
	"context"

	"github.com/treeverse/lakefs/graveler"
)

type mergeIterator struct {
	diffIt graveler.DiffIterator
	val    *graveler.ValueRecord
	base   Iterator
	ctx    context.Context
	ns     graveler.StorageNamespace
	err    error
}

// NewMergeIterator accepts an iterator describing a diff from theirs to ours.
// It returns a ValueIterator with the changes to perform on theirs, in order to merge ours into it,
// relative to base as the merge base.
// The iterator will return ErrConflictFound when it reaches a conflict.
func NewMergeIterator(diffTheirsToOurs graveler.DiffIterator, base Iterator) (*mergeIterator, error) {
	return &mergeIterator{diffIt: diffTheirsToOurs, base: base}, nil
}

func (d *mergeIterator) valueFromBase(key graveler.Key) *graveler.ValueRecord {
	d.base.SeekGE(key)
	var val *graveler.ValueRecord
	for d.base.Next() && val == nil {
		val, _ = d.base.Value()
	}
	if val == nil || !bytes.Equal(val.Key, key) {
		return nil
	}
	return val
}

func (d *mergeIterator) Next() bool {
	for d.diffIt.Next() {
		val := d.diffIt.Value()
		key := val.Key
		typ := val.Type
		baseVal := d.valueFromBase(key)
		switch typ {
		case graveler.DiffTypeAdded:
			if baseVal == nil {
				d.setValue()
				return true
			}
			if !bytes.Equal(baseVal.Identity, val.Value.Identity) {
				d.err = graveler.ErrConflictFound
				return false
			}
			continue
		case graveler.DiffTypeChanged:
			if baseVal == nil {
				d.err = graveler.ErrConflictFound
				return false
			}
			if bytes.Equal(baseVal.Identity, val.Value.Identity) {
				continue // no change from base
			}
			if !bytes.Equal(baseVal.Identity, val.OldIdentity) {
				d.err = graveler.ErrConflictFound
				return false
			}
			d.setValue()
			return true
		case graveler.DiffTypeRemoved:
			if baseVal != nil {
				if bytes.Equal(baseVal.Identity, val.OldIdentity) {
					d.setValue()
					return true // removed
				}
				d.err = graveler.ErrConflictFound
			}
			// continue
		}
	}
	return false
}

func (d *mergeIterator) setValue() {
	diff := d.diffIt.Value()
	if diff.Type == graveler.DiffTypeRemoved {
		d.val = &graveler.ValueRecord{Key: diff.Key}
	} else {
		d.val = &graveler.ValueRecord{
			Key:   diff.Key,
			Value: diff.Value,
		}
	}
}

func (d *mergeIterator) SeekGE(id graveler.Key) {
	d.val = nil
	d.err = nil
	d.diffIt.SeekGE(id)
}

func (d *mergeIterator) Value() *graveler.ValueRecord {
	return d.val
}

func (d *mergeIterator) Err() error {
	return d.err
}

func (d *mergeIterator) Close() {
	d.diffIt.Close()
}
