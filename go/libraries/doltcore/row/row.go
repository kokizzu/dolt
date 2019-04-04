package row

import (
	"errors"
	"github.com/attic-labs/noms/go/types"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/ld/dolt/go/libraries/utils/valutil"
)

var ErrRowNotValid = errors.New("invalid row for current schema.")

type Row interface {
	NomsMapKey(sch schema.Schema) types.Value
	NomsMapValue(sch schema.Schema) types.Value

	IterCols(cb func(tag uint64, val types.Value) (stop bool)) bool
	GetColVal(tag uint64) (types.Value, bool)
	SetColVal(tag uint64, val types.Value, sch schema.Schema) (Row, error)
}

func GetFieldByName(colName string, r Row, sch schema.Schema) (types.Value, bool) {
	col, ok := sch.GetAllCols().GetByName(colName)

	if !ok {
		panic("Requesting column that isn't in the schema. This is a bug. columns should be verified in the schema beforet attempted retrieval.")
	} else {
		return r.GetColVal(col.Tag)
	}
}

func GetFieldByNameWithDefault(colName string, defVal types.Value, r Row, sch schema.Schema) types.Value {
	col, ok := sch.GetAllCols().GetByName(colName)

	if !ok {
		panic("Requesting column that isn't in the schema. This is a bug. columns should be verified in the schema beforet attempted retrieval.")
	} else {
		val, ok := r.GetColVal(col.Tag)

		if !ok {
			return defVal
		}

		return val
	}
}

func IsValid(r Row, sch schema.Schema) bool {
	allCols := sch.GetAllCols()

	valid := true
	allCols.Iter(func(tag uint64, col schema.Column) (stop bool) {
		if len(col.Constraints) > 0 {
			val, _ := r.GetColVal(tag)

			for _, cnst := range col.Constraints {
				if !cnst.SatisfiesConstraint(val) {
					valid = false
					return true
				}
			}
		}

		return false
	})

	return valid
}

func GetInvalidCol(r Row, sch schema.Schema) *schema.Column {
	allCols := sch.GetAllCols()

	var badCol *schema.Column
	allCols.Iter(func(tag uint64, col schema.Column) (stop bool) {
		if len(col.Constraints) > 0 {
			val, _ := r.GetColVal(tag)

			for _, cnst := range col.Constraints {
				if !cnst.SatisfiesConstraint(val) {
					badCol = &col
					return true
				}
			}
		}

		return false
	})

	return badCol
}

func AreEqual(row1, row2 Row, sch schema.Schema) bool {
	if row1 == nil && row2 == nil {
		return true
	} else if row1 == nil || row2 == nil {
		return false
	}

	for _, tag := range sch.GetAllCols().Tags {
		val1, _ := row1.GetColVal(tag)
		val2, _ := row2.GetColVal(tag)

		if !valutil.NilSafeEqCheck(val1, val2) {
			return false
		}
	}

	return true
}

func GetTaggedVals(row Row) TaggedValues {
	taggedVals := make(TaggedValues)
	row.IterCols(func(tag uint64, val types.Value) (stop bool) {
		taggedVals[tag] = val
		return false
	})

	return taggedVals
}
