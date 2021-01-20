//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"math"

	"github.com/couchbase/query/value"
)

///////////////////////////////////////////////////
//
// IfInf
//
///////////////////////////////////////////////////

/*
This represents the Conditional function IFINF(expr1, expr2, ...)
for numbers. It returns the first non-MISSING, non-Inf number or
NULL.
*/
type IfInf struct {
	FunctionBase
}

func NewIfInf(operands ...Expression) Function {
	rv := &IfInf{
		*NewFunctionBase("ifinf", operands...),
	}

	rv.setConditional()
	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *IfInf) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *IfInf) Type() value.Type { return value.NUMBER }

/*
First non missing, non infinity number in the input argument values,
or null.
*/
func (this *IfInf) Evaluate(item value.Value, context Context) (value.Value, error) {
	null := false
	missing := false
	var rv value.Value
	for _, op := range this.operands {
		a, err := op.Evaluate(item, context)
		if err != nil {
			return nil, err
		} else if a.Type() == value.MISSING {
			// only set missing if we've not already encountered a null
			if !null {
				missing = true
			}
		} else if a.Type() != value.NUMBER {
			null = true
		} else if !null && rv == nil {
			f := a.Actual().(float64)
			if !math.IsInf(f, 0) {
				rv = value.NewValue(f)
			}
		}
	}

	if rv != nil {
		return rv, nil
	}
	if missing {
		return value.MISSING_VALUE, nil
	}
	return value.NULL_VALUE, nil
}

func (this *IfInf) DependsOn(other Expression) bool {
	return this.dependsOn(other)
}

/*
Minimum input arguments required is 2
*/
func (this *IfInf) MinArgs() int { return 2 }

/*
Maximum number of input arguments defined for the IfInf function is
MaxInt16 = 1<<15 - 1.
*/
func (this *IfInf) MaxArgs() int { return math.MaxInt16 }

/*
Factory method pattern.
*/
func (this *IfInf) Constructor() FunctionConstructor {
	return NewIfInf
}

///////////////////////////////////////////////////
//
// IfNaN
//
///////////////////////////////////////////////////

/*
This represents the Conditional function IFNAN(expr1, expr2, ...).  It
returns the first non-MISSING, non-NaN number or NULL.
*/
type IfNaN struct {
	FunctionBase
}

func NewIfNaN(operands ...Expression) Function {
	rv := &IfNaN{
		*NewFunctionBase("ifnan", operands...),
	}

	rv.setConditional()
	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *IfNaN) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *IfNaN) Type() value.Type { return value.NUMBER }

func (this *IfNaN) Evaluate(item value.Value, context Context) (value.Value, error) {
	null := false
	var rv value.Value
	for _, op := range this.operands {
		a, err := op.Evaluate(item, context)
		if err != nil {
			return nil, err
		} else if a.Type() == value.MISSING {
			continue
		} else if a.Type() != value.NUMBER {
			null = true
		} else if !null && rv == nil {
			f := a.Actual().(float64)
			if !math.IsNaN(f) {
				rv = value.NewValue(f)
			}
		}
	}

	if rv != nil {
		return rv, nil
	}
	return value.NULL_VALUE, nil
}

func (this *IfNaN) DependsOn(other Expression) bool {
	return this.dependsOn(other)
}

/*
Minimum input arguments required is 2.
*/
func (this *IfNaN) MinArgs() int { return 2 }

/*
Maximum number of input arguments defined for the IfNaN
function is MaxInt16  = 1<<15 - 1.
*/
func (this *IfNaN) MaxArgs() int { return math.MaxInt16 }

/*
Factory method pattern.
*/
func (this *IfNaN) Constructor() FunctionConstructor {
	return NewIfNaN
}

///////////////////////////////////////////////////
//
// IfNaNOrInf
//
///////////////////////////////////////////////////

/*
This represents the Conditional function IFNANORINF(expr1, expr2, ...).
It returns the first non-MISSING, non-Inf, non-NaN number or NULL.
*/
type IfNaNOrInf struct {
	FunctionBase
}

func NewIfNaNOrInf(operands ...Expression) Function {
	rv := &IfNaNOrInf{
		*NewFunctionBase("ifnanorinf", operands...),
	}

	rv.setConditional()
	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *IfNaNOrInf) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *IfNaNOrInf) Type() value.Type { return value.NUMBER }

func (this *IfNaNOrInf) Evaluate(item value.Value, context Context) (value.Value, error) {
	null := false
	var rv value.Value
	for _, op := range this.operands {
		a, err := op.Evaluate(item, context)
		if err != nil {
			return nil, err
		} else if a.Type() == value.MISSING {
			continue
		} else if a.Type() != value.NUMBER {
			null = true
		} else if !null && rv == nil {
			f := a.Actual().(float64)
			if !math.IsInf(f, 0) && !math.IsNaN(f) {
				rv = value.NewValue(f)
			}
		}
	}

	if rv != nil {
		return rv, nil
	}
	return value.NULL_VALUE, nil
}

func (this *IfNaNOrInf) DependsOn(other Expression) bool {
	return this.dependsOn(other)
}

/*
Minimum input arguments required is 2
*/
func (this *IfNaNOrInf) MinArgs() int { return 2 }

/*
Maximum number of input arguments defined for the
function is MaxInt16  = 1<<15 - 1.
*/
func (this *IfNaNOrInf) MaxArgs() int { return math.MaxInt16 }

/*
Factory method pattern.
*/
func (this *IfNaNOrInf) Constructor() FunctionConstructor {
	return NewIfNaNOrInf
}

///////////////////////////////////////////////////
//
// NaNIf
//
///////////////////////////////////////////////////

/*
This represents the Conditional function NANIF(expr1, expr2).
It returns a NaN if expr1 = expr2; else expr1.
*/
type NaNIf struct {
	BinaryFunctionBase
}

func NewNaNIf(first, second Expression) Function {
	rv := &NaNIf{
		*NewBinaryFunctionBase("nanif", first, second),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *NaNIf) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *NaNIf) Type() value.Type { return value.JSON }

func (this *NaNIf) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.BinaryEval(this, item, context)
}

/*
This method checks to see if the values of the two input expressions
are equal, and if true then returns a NaN. If not it returns the first
input value. Use the Equals method for the two values to determine
equality.
*/
func (this *NaNIf) Apply(context Context, first, second value.Value) (value.Value, error) {
	eq := first.Equals(second)
	switch eq.Type() {
	case value.MISSING, value.NULL:
		return eq, nil
	default:
		if eq.Truth() {
			return _NAN_VALUE, nil
		} else {
			return first, nil
		}
	}
}

/*
Factory method pattern.
*/
func (this *NaNIf) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewNaNIf(operands[0], operands[1])
	}
}

///////////////////////////////////////////////////
//
// NegInfIf
//
///////////////////////////////////////////////////

/*
This represents the Conditional function NEGINFIF(expr1, expr2).
It returns NegInf if expr1 = expr2; else expr1.
*/
type NegInfIf struct {
	BinaryFunctionBase
}

func NewNegInfIf(first, second Expression) Function {
	rv := &NegInfIf{
		*NewBinaryFunctionBase("neginfif", first, second),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *NegInfIf) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *NegInfIf) Type() value.Type { return value.JSON }

func (this *NegInfIf) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.BinaryEval(this, item, context)
}

/*
This method checks to see if the values of the two input expressions
are equal, and if true then returns a negative infinity.. If not it
returns the first input value. Use the Equals method for the two
values to determine equality.
*/
func (this *NegInfIf) Apply(context Context, first, second value.Value) (value.Value, error) {
	eq := first.Equals(second)
	switch eq.Type() {
	case value.MISSING, value.NULL:
		return eq, nil
	default:
		if eq.Truth() {
			return _NEG_INF_VALUE, nil
		} else {
			return first, nil
		}
	}
}

var _NEG_INF_VALUE = value.NewValue(math.Inf(-1))

/*
Factory method pattern.
*/
func (this *NegInfIf) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewNegInfIf(operands[0], operands[1])
	}
}

///////////////////////////////////////////////////
//
// PosInfIf
//
///////////////////////////////////////////////////

/*
This represents the Conditional function POSINFIF(expr1, expr2).
It returns PosInf if expr1 = expr2; else expr1.
*/
type PosInfIf struct {
	BinaryFunctionBase
}

func NewPosInfIf(first, second Expression) Function {
	rv := &PosInfIf{
		*NewBinaryFunctionBase("posinfif", first, second),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *PosInfIf) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *PosInfIf) Type() value.Type { return value.JSON }

func (this *PosInfIf) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.BinaryEval(this, item, context)
}

/*
This method checks to see if the values of the two input expressions
are equal, and if true then returns a positive infinity. If not it
returns the first input value. Use the Equals method for the two
values to determine equality.
*/
func (this *PosInfIf) Apply(context Context, first, second value.Value) (value.Value, error) {
	eq := first.Equals(second)
	switch eq.Type() {
	case value.MISSING, value.NULL:
		return eq, nil
	default:
		if eq.Truth() {
			return _POS_INF_VALUE, nil
		} else {
			return first, nil
		}
	}
}

var _POS_INF_VALUE = value.NewValue(math.Inf(1))

/*
Factory method pattern.
*/
func (this *PosInfIf) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewPosInfIf(operands[0], operands[1])
	}
}
