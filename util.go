package hw2

import (
	"github.com/gonum/graph"
)

// WhichBucket returns the index that the node should be added to
func WhichBucket(weight, stepSize float64) int {
	count := 0
	for i := stepSize; i < weight; i += stepSize {
		count++
	}
	return count
}

// AddToBucket puts nodes in the right bucket
func AddToBucket(bucket [][]graph.Node, stepSize int, fromNode, toNode graph.Node, g graph.Graph) [][]graph.Node {

	// first need to figure out what it should go into
	var index = WhichBucket(g.Edge(fromNode, toNode).Weight(), float64(stepSize))

	if bucket == nil {
		bucket = append(bucket, make([]graph.Node, 0))
	}
	for i := 0; i <= index; i++ {
		if len(bucket) <= i {
			bucket = append(bucket, make([]graph.Node, 0))
		}
	}
	bucket[index] = append(bucket[index], toNode)
	return bucket
}

// SliceContains determines if a node is currently in the slice
func SliceContains(slice []graph.Node, node graph.Node) bool {
	for _, item := range slice {
		if item == node {
			return true
		}
	}
	return false
}
