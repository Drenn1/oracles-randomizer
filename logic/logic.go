package logic

import (
	"fmt"
)

// This package contains definitions of nodes and node relationships before
// they are inserted into the graph. This is necessary because nodes
// relationships can't be made until the nodes are added first (and it's nice
// not to clutter the other packages with all these definitions).

// A Type identifies whether a node is an And, Or, or Root node, whether it is
// an item slot, and whether it is a non-item slot milestone.
type Type int

// And, Or, and Root are pretty self-explanatory. One with a Slot suffix is an
// item slot, and one with a Step suffix is treated as a milestone for routing
// purposes. Slot types are also treated as steps; see the Point.IsStep()
// function.
//
// "Hard" nodes are ones that players aren't expected to do because they're too
// difficult or esoteric, but they're used to prevent softlocks by knowing that
// players *can* do them.
//
// The following functions are half syntactic sugar for declaring large lists
// of node relationships.
const (
	RootType Type = iota
	AndType
	OrType
	AndSlotType
	OrSlotType
	AndStepType
	OrStepType
	HardAndType
	HardOrType
)

// A Node is a mapping of strings that will become And or Or nodes in the
// graph. A node can have nested nodes as parents instead of strings.
type Node struct {
	Parents []interface{}
	Type    Type
}

// CreateFunc returns a function that creates graph nodes from a list of key
// strings or sub-nodes, based on the given node type.
func CreateFunc(nodeType Type) func(parents ...interface{}) *Node {
	return func(parents ...interface{}) *Node {
		return &Node{Parents: parents, Type: nodeType}
	}
}

// Convenience functions for creating nodes succinctly. See the Type const
// comment for information on the various types.
var (
	Root    = CreateFunc(RootType)
	And     = CreateFunc(AndType)
	AndSlot = CreateFunc(AndSlotType)
	AndStep = CreateFunc(AndStepType)
	Or      = CreateFunc(OrType)
	OrSlot  = CreateFunc(OrSlotType)
	OrStep  = CreateFunc(OrStepType)
	Hard    = CreateFunc(HardAndType) // for wrapping single nodes
	HardAnd = CreateFunc(HardAndType)
	HardOr  = CreateFunc(HardOrType)
)

var seasonsNodes, agesNodes map[string]*Node

func init() {
	seasonsNodes = make(map[string]*Node)
	appendNodes(seasonsNodes,
		seasonsItemNodes, seasonsBaseItemNodes, seasonsKillNodes,
		holodrumNodes, subrosiaNodes, portalNodes, seasonNodes,
		seasonsD0Nodes, seasonsD1Nodes, seasonsD2Nodes, seasonsD3Nodes,
		seasonsD4Nodes, seasonsD5Nodes, seasonsD6Nodes, seasonsD7Nodes,
		seasonsD8Nodes, seasonsD9Nodes)
	flattenNestedNodes(seasonsNodes)

	agesNodes = make(map[string]*Node)
	appendNodes(agesNodes,
		agesItemNodes, agesKillNodes, labrynnaNodes,
		agesD1Nodes, agesD2Nodes, agesD3Nodes, agesD4Nodes,
		agesD5Nodes, agesD6Nodes, agesD7Nodes, agesD8Nodes, agesD9Nodes)
	flattenNestedNodes(agesNodes)
}

// add nested nodes to the map and turn their references into strings
func flattenNestedNodes(nodes map[string]*Node) {
	done := true

	for name, pn := range nodes {
		subID := 0
		for i, parent := range pn.Parents {
			switch parent := parent.(type) {
			case *Node:
				subID++
				subName := fmt.Sprintf("%s %d", name, subID)
				pn.Parents[i] = subName
				nodes[subName] = parent
				done = false
			}
		}
	}

	// recurse if nodes were added
	if !done {
		flattenNestedNodes(nodes)
	}
}

// SeasonsExtraItems returns a map of item nodes that may be assigned to slots,
// in addition to the ones that are generated from default slot contents.
func SeasonsExtraItems() map[string]*Node {
	return copyMap(seasonsBaseItemNodes)
}

// GetSeasons returns a copy of all seasons nodes.
func GetSeasons() map[string]*Node {
	return copyMap(seasonsNodes)
}

// GetAges returns a copy of all ages nodes.
func GetAges() map[string]*Node {
	return copyMap(agesNodes)
}

// merge the given maps into the first argument
func appendNodes(total map[string]*Node, maps ...map[string]*Node) {
	for _, nodeMap := range maps {
		for k, v := range nodeMap {
			if _, ok := total[k]; ok {
				panic("fatal: duplicate logic key: " + k)
			}
			total[k] = v
		}
	}
}

// returns a shallow copy of a string/node map
func copyMap(src map[string]*Node) map[string]*Node {
	dst := make(map[string]*Node, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
