/*
 * Copyright 2019 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package migrate

import "fmt"

type NodeColor int

const (
	WHITE NodeColor = iota
	GREY
	BLACK
)

// visit traverses the dependency graph in the order of depth-first search
// and when we are done visiting a node, it will be added to the collector.
// As a result, it returns the new collector
func visit(tables map[string]*TableInfo, nodeColor map[string]NodeColor, curTable string,
	collector []string) ([]string, error) {
	switch nodeColor[curTable] {
	case WHITE:
		nodeColor[curTable] = GREY
		// visit all the children of this table
		for childTable := range tables[curTable].referencedTables {
			var err error
			collector, err = visit(tables, nodeColor, childTable, collector)
			if err != nil {
				return nil, err
			}
		}

		// done visiting the current node, add it to the collector
		nodeColor[curTable] = BLACK
		collector = append(collector, curTable)
	case GREY:
		// this forms a loop, error out
		return nil, fmt.Errorf("found reference loops while visiting table %s", curTable)
	case BLACK:
		// there are multiple paths pointing to curTable, that's allowed
		// we simply ignore the node since it has been visited
	}

	return collector, nil
}

// topoSortTables runs a topological sort among the tables following the dependency created
// by foreign key references, the goal is to process the most deeply referenced tables first,
// and the unreferenced tables later
func topoSortTables(tables map[string]*TableInfo) ([]string, error) {
	nodeColor := make(map[string]NodeColor)
	// initialize each node to have the WHITE coler
	for table := range tables {
		nodeColor[table] = WHITE
	}
	collector := make([]string, 0)
	var err error
	for table := range tables {
		collector, err = visit(tables, nodeColor, table, collector)
		if err != nil {
			return nil, err
		}
	}

	return collector, nil
}
