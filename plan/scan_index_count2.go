//  Copyright (c) 2016 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plan

import (
	"encoding/json"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/expression/parser"
	"github.com/couchbase/query/value"
)

type IndexCountScan2 struct {
	readonly
	index        datastore.CountIndex2
	term         *algebra.KeyspaceTerm
	spans        Spans2
	covers       expression.Covers
	filterCovers map[*expression.Cover]value.Value
}

func NewIndexCountScan2(index datastore.CountIndex2, term *algebra.KeyspaceTerm,
	spans Spans2, covers expression.Covers, filterCovers map[*expression.Cover]value.Value) *IndexCountScan2 {
	return &IndexCountScan2{
		index:        index,
		term:         term,
		spans:        spans,
		covers:       covers,
		filterCovers: filterCovers,
	}
}

func (this *IndexCountScan2) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIndexCountScan2(this)
}

func (this *IndexCountScan2) New() Operator {
	return &IndexCountScan2{}
}

func (this *IndexCountScan2) Index() datastore.CountIndex2 {
	return this.index
}

func (this *IndexCountScan2) Term() *algebra.KeyspaceTerm {
	return this.term
}

func (this *IndexCountScan2) Spans() Spans2 {
	return this.spans
}

func (this *IndexCountScan2) Covers() expression.Covers {
	return this.covers
}

func (this *IndexCountScan2) FilterCovers() map[*expression.Cover]value.Value {
	return this.filterCovers
}

func (this *IndexCountScan2) Covering() bool {
	return len(this.covers) > 0
}

func (this *IndexCountScan2) Limit() expression.Expression {
	return nil
}

func (this *IndexCountScan2) SetLimit(limit expression.Expression) {
}

func (this *IndexCountScan2) Offset() expression.Expression {
	return nil
}

func (this *IndexCountScan2) SetOffset(offset expression.Expression) {
}

func (this *IndexCountScan2) CoverJoinSpanExpressions(coverer *expression.Coverer) error {
	return nil
}

func (this *IndexCountScan2) String() string {
	bytes, _ := this.MarshalJSON()
	return string(bytes)
}

func (this *IndexCountScan2) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.MarshalBase(nil))
}

func (this *IndexCountScan2) MarshalBase(f func(map[string]interface{})) map[string]interface{} {
	r := map[string]interface{}{"#operator": "IndexCountScan2"}
	r["index"] = this.index.Name()
	r["index_id"] = this.index.Id()
	r["namespace"] = this.term.Namespace()
	r["keyspace"] = this.term.Keyspace()
	r["using"] = this.index.Type()
	r["spans"] = this.spans

	if this.term.As() != "" {
		r["as"] = this.term.As()
	}

	if len(this.covers) > 0 {
		r["covers"] = this.covers
	}

	if len(this.filterCovers) > 0 {
		fc := make(map[string]value.Value, len(this.filterCovers))
		for c, v := range this.filterCovers {
			fc[c.String()] = v
		}

		r["filter_covers"] = fc
	}

	if f != nil {
		f(r)
	}
	return r
}

func (this *IndexCountScan2) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_            string                 `json:"#operator"`
		Index        string                 `json:"index"`
		IndexId      string                 `json:"index_id"`
		Namespace    string                 `json:"namespace"`
		Keyspace     string                 `json:"keyspace"`
		As           string                 `json:"as"`
		Using        datastore.IndexType    `json:"using"`
		Spans        Spans2                 `json:"spans"`
		Covers       []string               `json:"covers"`
		FilterCovers map[string]interface{} `json:"filter_covers"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	k, err := datastore.GetKeyspace(_unmarshalled.Namespace, _unmarshalled.Keyspace)
	if err != nil {
		return err
	}

	this.term = algebra.NewKeyspaceTerm(_unmarshalled.Namespace, _unmarshalled.Keyspace, _unmarshalled.As, nil, nil)
	this.spans = _unmarshalled.Spans

	if len(_unmarshalled.Covers) > 0 {
		this.covers = make(expression.Covers, len(_unmarshalled.Covers))
		for i, c := range _unmarshalled.Covers {
			expr, err := parser.Parse(c)
			if err != nil {
				return err
			}

			this.covers[i] = expression.NewCover(expr)
		}
	}

	if len(_unmarshalled.FilterCovers) > 0 {
		this.filterCovers = make(map[*expression.Cover]value.Value, len(_unmarshalled.FilterCovers))
		for k, v := range _unmarshalled.FilterCovers {
			expr, err := parser.Parse(k)
			if err != nil {
				return err
			}

			c := expression.NewCover(expr)
			this.filterCovers[c] = value.NewValue(v)
		}
	}

	indexer, err := k.Indexer(_unmarshalled.Using)
	if err != nil {
		return err
	}

	index, err := indexer.IndexById(_unmarshalled.IndexId)
	if err != nil {
		return err
	}

	countIndex2, ok := index.(datastore.CountIndex2)
	if !ok {
		return errors.NewError(nil, "Unable to find Count2() for index")
	}
	this.index = countIndex2

	return nil
}
