//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"encoding/json"
	"fmt"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

type IntersectScan struct {
	base
	plan   *plan.IntersectScan
	scans  []Operator
	values map[string]value.AnnotatedValue
	bits   map[string]int64
	sent   int64
}

func NewIntersectScan(plan *plan.IntersectScan, context *Context, scans []Operator) *IntersectScan {
	rv := &IntersectScan{
		plan:  plan,
		scans: scans,
	}

	newBase(&rv.base, context)
	rv.trackChildren(len(scans))
	rv.output = rv
	return rv
}

func (this *IntersectScan) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIntersectScan(this)
}

func (this *IntersectScan) Copy() Operator {
	scans := _INDEX_SCAN_POOL.Get()

	for _, s := range this.scans {
		scans = append(scans, s.Copy())
	}

	rv := &IntersectScan{
		plan:  this.plan,
		scans: scans,
	}
	this.base.copy(&rv.base)
	return rv
}

func (this *IntersectScan) RunOnce(context *Context, parent value.Value) {
	this.once.Do(func() {
		defer context.Recover() // Recover from any panic
		active := this.active()
		defer this.close(context)
		this.switchPhase(_EXECTIME)
		defer this.switchPhase(_NOTIME)
		defer this.notify() // Notify that I have stopped

		defer func() {
			this.values = nil
			this.bits = nil
		}()

		if !active || !context.assert(len(this.scans) != 0, "Intersect scan has no scans") {
			return
		}
		pipelineCap := int(context.GetPipelineCap())
		if pipelineCap <= _INDEX_VALUE_POOL.Size() {
			this.values = _INDEX_VALUE_POOL.Get()
			this.bits = _INDEX_BIT_POOL.Get()

			defer func() {
				_INDEX_VALUE_POOL.Put(this.values)
				_INDEX_BIT_POOL.Put(this.bits)
			}()
		} else {
			this.values = make(map[string]value.AnnotatedValue, pipelineCap)
			this.bits = make(map[string]int64, pipelineCap)
		}

		fullBits := int64(0)
		for i, scan := range this.scans {
			scan.SetBit(uint8(i))
			fullBits |= int64(0x01) << uint8(i)
		}

		channel := NewChannel(context)
		this.SetInput(channel)

		for _, scan := range this.scans {
			scan.SetParent(this)
			scan.SetOutput(channel)
			go scan.RunOnce(context, parent)
		}

		limit := getLimit(this.plan.Limit(), this.plan.Covering(), context)
		n := len(this.scans)
		nscans := len(this.scans)
		childBits := int64(0)
		stopped := false
		ok := true

	loop:
		for ok {
			item, childBit, cont := this.getItemChildren()
			if cont {
				if childBit >= 0 {

					// MB-22321 terminate when first child terminates
					if n == nscans {
						notifyChildren(this.scans...)
						childBits |= int64(0x01) << uint(childBit)
					}
					n--

					// now that all children are gone, flag that there's
					// no more values coming in
					if n == 0 {
						channel.close(context)
					}
				} else if item != nil {
					this.addInDocs(1)
					ok = this.processKey(item, context, fullBits, limit)
				} else {
					break loop
				}
			} else {
				stopped = true

				// if not done already, stop children, wait and clean up
				if n == nscans {
					notifyChildren(this.scans...)
				}
				if n > 0 {
					this.childrenWaitNoStop(n)
					channel.close(context)
				}
				break loop
			}
		}

		if !stopped && ok && childBits != 0 && (limit <= 0 || this.sent < limit) {
			this.sendItems(childBits)
		}
	})
}

func (this *IntersectScan) processKey(item value.AnnotatedValue,
	context *Context, fullBits, limit int64) bool {

	m := item.GetAttachment("meta")
	meta, ok := m.(map[string]interface{})
	if !ok {
		context.Error(errors.NewInvalidValueError(
			fmt.Sprintf("Missing or invalid meta %v of type %T.", m, m)))
		return false
	}

	k := meta["id"]
	key, ok := k.(string)
	if !ok {
		context.Error(errors.NewInvalidValueError(
			fmt.Sprintf("Missing or invalid primary key %v of type %T.", k, k)))
		return false
	}

	bits := this.bits[key]
	if bits == 0 {
		this.values[key] = item
	}

	bits |= int64(0x01) << item.Bit()

	if (bits&fullBits)^fullBits == 0 {
		delete(this.values, key)
		delete(this.bits, key)

		if limit > 0 {
			this.sent++
		}

		item.SetBit(this.bit)
		return this.sendItem(item) && (limit <= 0 || this.sent < limit)
	}

	this.bits[key] = bits
	return true
}

func (this *IntersectScan) sendItems(childBits int64) {
	if childBits == 0 {
		return
	}

	for key, bits := range this.bits {
		if (bits&childBits)^childBits == 0 {
			item := this.values[key]
			item.SetBit(this.bit)
			if !this.sendItem(item) {
				return
			}
		}
	}
}

func (this *IntersectScan) MarshalJSON() ([]byte, error) {
	r := this.plan.MarshalBase(func(r map[string]interface{}) {
		this.marshalTimes(r)
		r["scans"] = this.scans
	})
	return json.Marshal(r)
}

func (this *IntersectScan) accrueTimes(o Operator) {
	if baseAccrueTimes(this, o) {
		return
	}
	copy, _ := o.(*IntersectScan)
	childrenAccrueTimes(this.scans, copy.scans)
}

func (this *IntersectScan) SendStop() {
	this.baseSendStop()
	for _, scan := range this.scans {
		scan.SendStop()
	}
}

func (this *IntersectScan) reopen(context *Context) {
	this.baseReopen(context)
	for _, scan := range this.scans {
		scan.reopen(context)
	}
}

func (this *IntersectScan) Done() {
	this.baseDone()
	for s, scan := range this.scans {
		scan.Done()
		this.scans[s] = nil
	}
	_INDEX_SCAN_POOL.Put(this.scans)
	this.scans = nil
}
