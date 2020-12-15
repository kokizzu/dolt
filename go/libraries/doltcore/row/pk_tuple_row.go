// Copyright 2020 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package row

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/types"
)

type tupleRow struct {
	key types.Tuple
	val types.Tuple
}

var _ Row = tupleRow{}

func RowFromTuples(key, val types.Tuple) Row {
	return tupleRow{
		key: key,
		val: val,
	}
}

func (r tupleRow) NomsMapKey(sch schema.Schema) types.LesserValuable {
	return r.key
}

func (r tupleRow) NomsMapValue(sch schema.Schema) types.Valuable {
	return r.val
}

func (r tupleRow) IterCols(cb func(tag uint64, val types.Value) (stop bool, err error)) (bool, error) {
	stop, err := iterTuple(r.key, cb)
	if err != nil || !stop {
		return stop, err
	}

	stop, err = iterTuple(r.val, cb)
	if err != nil || !stop {
		return stop, err
	}

	return false, nil
}

func (r tupleRow) IterSchema(sch schema.Schema, cb func(tag uint64, val types.Value) (stop bool, err error)) (bool, error) {
	tags := sch.GetAllCols().Tags
	vals := make([]types.Value, len(tags))

	_, err := iterTuple(r.key, func(tag uint64, val types.Value) (stop bool, err error) {
		idx := sch.GetAllCols().TagToIdx[tag]
		vals[idx] = val
		return
	})
	if err != nil {
		return false, err
	}

	_, err = iterTuple(r.val, func(tag uint64, val types.Value) (stop bool, err error) {
		idx := sch.GetAllCols().TagToIdx[tag]
		vals[idx] = val
		return
	})
	if err != nil {
		return false, err
	}

	for idx, tag := range tags {
		stop, err := cb(tag, vals[idx])
		if err != nil {
			return false, err
		}
		if stop {
			return stop, nil
		}
	}

	return true, nil
}

func (r tupleRow) GetColVal(tag uint64) (val types.Value, ok bool) {
	_, _ = r.IterCols(func(t uint64, v types.Value) (stop bool, err error) {
		if tag == t {
			val = v
			ok, stop = true, true
		}
		return
	})
	return val, ok
}

func (r tupleRow) SetColVal(updateTag uint64, updateVal types.Value, sch schema.Schema) (Row, error) {
	if _, ok := sch.GetPKCols().GetByTag(updateTag); ok {
		return r.setPkVal(updateTag, updateVal, sch)
	} else if _, ok := sch.GetNonPKCols().GetByTag(updateTag); ok {
		return r.setNonPkVal(updateTag, updateVal, sch)
	} else {
		return nil, fmt.Errorf("tag %d not found in sch", updateTag)
	}
}

// assumes all values present
func (r tupleRow) setPkVal(updateTag uint64, updateVal types.Value, sch schema.Schema) (Row, error) {
	i := 0
	vals := make([]types.Value, r.key.Len())
	_, err := iterTuple(r.key, func(tag uint64, val types.Value) (stop bool, err error) {
		if tag == updateTag {
			val = updateVal
		}
		vals[i] = types.Uint(tag)
		vals[i+1] = val
		i += 2
		return
	})
	if err != nil {
		return nil, err
	}

	tup, err := types.NewTuple(r.Format(), vals...)
	if err != nil {
		return nil, err
	}

	return tupleRow{
		key: tup,
		val: r.val,
	}, nil
}

// assumes (ColTag,ColVal) pairs are in order of sch.GetNonPkCols().Tags.
func (r tupleRow) setNonPkVal(updateTag uint64, updateVal types.Value, sch schema.Schema) (Row, error) {
	i := 0
	vals := make([]types.Value, r.val.Len()+2) // allow for insert

	j := 0
	tags := sch.GetNonPKCols().Tags

	iter, err := r.val.Iterator()
	if err != nil {
		return nil, err
	}

	for {
		_, v1, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if v1 == nil {
			break
		}

		readTag, ok := v1.(types.Uint)
		if !ok {
			return nil, fmt.Errorf("expected tag types.Uint, got %v", v1)
		}

		for {
			if tags[j] == uint64(readTag) {
				break
			}
			if tags[j] == updateTag {
				// value was previously nil
				vals[i] = types.Uint(updateTag)
				vals[i+1] = updateVal
				i += 2
			}
			j++
		}

		_, v2, err := iter.Next()
		if err != nil {
			return nil, err
		}

		if tags[j] == updateTag {
			// value was previously v2
			v2 = updateVal
		}

		vals[i] = v1
		vals[i+1] = v2
		i += 2
	}

	tup, err := types.NewTuple(r.Format(), vals[:i]...)
	if err != nil {
		return nil, err
	}

	return tupleRow{
		key: r.key,
		val: tup,
	}, nil
}

func (r tupleRow) ReduceToIndex(idx schema.Index) (Row, error) {
	i := 0
	vals := make([]types.Value, len(idx.AllTags()))
	for _, tag := range idx.AllTags() {
		if val, ok := r.GetColVal(tag); ok {
			vals[i] = val
			i++
		}
	}

	k, err := types.NewTuple(r.Format(), vals[:i]...)
	if err != nil {
		return nil, err
	}

	v, err := types.NewTuple(r.Format())
	if err != nil {
		return nil, err
	}

	return tupleRow{key: k, val: v}, nil
}

func (r tupleRow) ReduceToIndexPartialKey(idx schema.Index) (types.Tuple, error) {
	var vals []types.Value
	for _, tag := range idx.IndexedColumnTags() {
		val, ok := r.GetColVal(tag)
		if !ok {
			val = types.NullValue
		}
		vals = append(vals, types.Uint(tag), val)
	}
	return types.NewTuple(r.Format(), vals...)
}

func (r tupleRow) Format() *types.NomsBinFormat {
	return r.val.Format()
}

func iterTuple(tup types.Tuple, cb func(tag uint64, val types.Value) (bool, error)) (bool, error) {
	iter, err := tup.Iterator()
	if err != nil {
		return false, err
	}

	for {
		_, v, err := iter.Next()
		if err != nil {
			return false, err
		}
		if v == nil {
			break
		}

		tag, ok := v.(types.Uint)
		if !ok {
			return false, fmt.Errorf("expected tag types.Uint, got %v", v)
		}

		_, v, err = iter.Next()
		if err != nil {
			return false, err
		}

		stop, err := cb(uint64(tag), v)
		if err != nil {
			return false, nil
		}
		if stop {
			return stop, nil
		}
	}
	return false, nil
}
