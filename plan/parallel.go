//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package plan

import (
	"encoding/json"
	"runtime"
)

type Parallel struct {
	child          Operator
	maxParallelism int
}

func NewParallel(child Operator, maxParallelism int) *Parallel {
	return &Parallel{child, maxParallelism}
}

func (this *Parallel) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitParallel(this)
}

func (this *Parallel) New() Operator {
	return &Parallel{}
}

func (this *Parallel) Readonly() bool {
	return this.child.Readonly()
}

func (this *Parallel) Child() Operator {
	return this.child
}

func (this *Parallel) MaxParallelism() int {
	if this.maxParallelism <= 0 {
		return GetMaxParallelism()
	}

	return this.maxParallelism
}

func (this *Parallel) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.MarshalBase(nil))
}

func (this *Parallel) MarshalBase(f func(map[string]interface{})) map[string]interface{} {
	r := map[string]interface{}{"#operator": "Parallel"}

	if this.maxParallelism > 0 {
		r["maxParallelism"] = this.maxParallelism
	}

	if f != nil {
		f(r)
	} else {
		r["~child"] = this.child
	}
	return r
}

func (this *Parallel) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_              string          `json:"#operator"`
		MaxParallelism int             `json:"maxParallelism"`
		Child          json.RawMessage `json:"~child"`
	}
	var child_type struct {
		Operator string `json:"#operator"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	err = json.Unmarshal(_unmarshalled.Child, &child_type)
	if err != nil {
		return err
	}

	this.maxParallelism = _unmarshalled.MaxParallelism
	this.child, err = MakeOperator(child_type.Operator, _unmarshalled.Child)
	return err
}

func (this *Parallel) verify(prepared *Prepared) bool {
	return this.child.verify(prepared)
}

func (this *Parallel) Cost() float64 {
	return this.child.Cost()
}

func (this *Parallel) Cardinality() float64 {
	return this.child.Cardinality()
}

func (this *Parallel) Size() int64 {
	return this.child.Size()
}

func (this *Parallel) FrCost() float64 {
	return this.child.FrCost()
}

func GetMaxParallelism() int {
	return runtime.NumCPU()
}
