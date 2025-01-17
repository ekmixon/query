//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package planner

import (
	"fmt"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
	base "github.com/couchbase/query/plannerbase"
)

func (this *builder) buildPrimaryScan(keyspace datastore.Keyspace, node *algebra.KeyspaceTerm,
	indexes []datastore.Index, id expression.Expression, force, exact, hasDeltaKeyspace bool) (
	plan.Operator, error) {
	primary, err := buildPrimaryIndex(keyspace, indexes, node, force)
	if primary == nil || err != nil {
		return nil, err
	}

	this.resetProjection()
	if this.group != nil {
		this.resetPushDowns()
	} else if !exact {
		this.resetOffsetLimit()
	}

	var indexOrder plan.IndexKeyOrders

	if this.order != nil {
		keys := expression.Expressions{id}
		entry := newIndexEntry(primary, keys, keys, nil, 1, 1, 1, nil, nil, _EXACT_VALUED_SPANS, exact, []bool{true})
		ok := true
		if ok, indexOrder = this.useIndexOrder(entry, entry.keys); ok {
			this.maxParallelism = 1
		} else {
			this.resetPushDowns()
		}
	}

	baseKeyspace, _ := this.baseKeyspaces[node.Alias()]
	if primary.Type() != datastore.SYSTEM {
		this.collectIndexKeyspaceNames(baseKeyspace.Keyspace())
	}
	if primary3, ok := primary.(datastore.PrimaryIndex3); ok && useIndex3API(primary, this.context.IndexApiVersion()) {
		cost := OPT_COST_NOT_AVAIL
		cardinality := OPT_CARD_NOT_AVAIL
		size := OPT_SIZE_NOT_AVAIL
		frCost := OPT_COST_NOT_AVAIL
		if this.useCBO && this.keyspaceUseCBO(node.Alias()) {
			cost, cardinality, size, frCost = primaryIndexScanCost(primary, this.context.RequestId(), this.context)
		}
		return plan.NewPrimaryScan3(primary3, keyspace, node, this.offset, this.limit,
			plan.NewIndexProjection(0, true), indexOrder, nil, cost, cardinality,
			size, frCost, hasDeltaKeyspace), nil
	}

	var limit expression.Expression
	if exact {
		limit = offsetPlusLimit(this.offset, this.limit)
		this.resetOffset()
	}

	return plan.NewPrimaryScan(primary, keyspace, node, limit, hasDeltaKeyspace), nil
}

func (this *builder) buildCoveringPrimaryScan(keyspace datastore.Keyspace, node *algebra.KeyspaceTerm,
	id expression.Expression, indexes []datastore.Index) (plan.Operator, error) {

	primary, err := buildPrimaryIndex(keyspace, indexes, node, false)
	if err != nil {
		return nil, err
	}

	keys := expression.Expressions{id}

	formalizer := expression.NewSelfFormalizer(node.Alias(), nil)
	partitionKeys, err := indexPartitionKeys(primary, formalizer)
	if err != nil {
		return nil, err
	}

	entry := newIndexEntry(primary, keys, keys, partitionKeys, 1, 1, 1, nil, nil, _EXACT_VALUED_SPANS, true, []bool{true})
	secondaries := map[datastore.Index]*indexEntry{primary: entry}

	pred := expression.NewIsNotNull(id)
	baseKeyspace := base.NewBaseKeyspace(node.Alias(), node.Path())
	keyspaces := make(map[string]string, 1)
	keyspaces[node.Alias()] = node.Keyspace()
	origKeyspaces := make(map[string]string, 1)
	origKeyspaces[node.Alias()] = node.Keyspace()
	newfilter := base.NewFilter(pred, pred, keyspaces, origKeyspaces, false, false)
	if this.useCBO && this.keyspaceUseCBO(node.Alias()) {
		newfilter.SetSelec(1.0)
		newfilter.SetSelecDone()
	}
	baseKeyspace.AddFilter(newfilter)
	baseKeyspace.SetPreds(pred, nil, nil)
	op, _, err := this.buildCoveringScan(secondaries, node, baseKeyspace, id)
	return op, err
}

func buildPrimaryIndex(keyspace datastore.Keyspace, indexes []datastore.Index, node *algebra.KeyspaceTerm, force bool) (
	primary datastore.PrimaryIndex, err error) {
	ok := false

	// Prefer hints
	for _, index := range indexes {
		if !index.IsPrimary() {
			continue
		}

		primary, ok = index.(datastore.PrimaryIndex)
		if ok {
			return
		} else {
			return nil, fmt.Errorf("Unable to cast index %s to primary index", index.Name())
		}
	}

	if force {
		return
	}

	// Return first online primary index
	indexers, er := keyspace.Indexers()
	if er != nil {
		return nil, er
	}

	for _, indexer := range indexers {
		primaries, er := indexer.PrimaryIndexes()
		if er != nil {
			return nil, er
		}

		for _, primary = range primaries {
			state, _, er := primary.State()
			if er != nil {
				return nil, er
			}

			if state == datastore.ONLINE {
				return
			}
		}
	}

	if primary == nil {
		return nil, fmt.Errorf(
			"No index available on keyspace %s that matches your query. Use CREATE PRIMARY INDEX ON %s to create a primary index, or check that your expected index is online.",
			node.PathString(), node.PathString())
	}

	return nil, fmt.Errorf("Primary index %s not online.", primary.Name())
}
