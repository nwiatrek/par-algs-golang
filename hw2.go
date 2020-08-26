package hw2

import (
	"sync"

	"github.com/gonum/graph"
)

var waitGroup sync.WaitGroup
var mutex sync.Mutex

// BellmanFord Apply the bellman-ford algorihtm to Graph and return
// a shortest path tree.
//
// Note that this uses Shortest to make it easier for you,
// but you can use another struct if that makes more sense
// for the concurrency model you chose.
func BellmanFord(s graph.Node, g graph.Graph) Shortest {
	// Small note, I wrote deltaStepping first and then came back to this algorithm
	// this is why both the sets of code look similar
	if !g.Has(s) {
		return Shortest{from: s}
	}
	var weight Weighting
	if wg, ok := g.(graph.Weighter); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	nodes := g.Nodes()
	path := newShortestFrom(s, nodes)
	var exploringNodes []graph.Node
	exploringNodes = make([]graph.Node, 0)
	exploringNodes = append(exploringNodes, s)
	// to handle negative cycles we need to count how many iterations
	// we've done compared to total possible iterations
	maxIterations := len(nodes) * len(nodes)
	count := 0

	for {
		if len(exploringNodes) == 0 {
			break
		}
		var item = exploringNodes[0]
		// remove from list
		copy(exploringNodes[0:], exploringNodes[1:])
		exploringNodes[len(exploringNodes)-1] = nil
		exploringNodes = exploringNodes[:len(exploringNodes)-1]
		// explore the current node
		num := len(g.From(item))
		c := make(chan ParPair, num)
		// with the go routine we get all the weights in 1 step
		for _, node := range g.From(item) {
			waitGroup.Add(1)
			go getWeight(c, path, weight, item, node)
		}
		waitGroup.Wait()
		close(c)
		for parPair := range c {
			j := path.indexOf[parPair.to.ID()]
			k := path.indexOf[parPair.from.ID()]
			if parPair.weight < path.dist[j] {
				path.set(j, parPair.weight, k)
				if !SliceContains(exploringNodes, parPair.to) {
					exploringNodes = append(exploringNodes, parPair.to)
				}
			}
		}
		if count > maxIterations {
			path.hasNegativeCycle = true
			return path
		}
		count++
	}

	return path
}

func getWeight(c chan ParPair, path Shortest, weight Weighting, item, toNode graph.Node) {
	defer waitGroup.Done()
	k := path.indexOf[item.ID()]

	wg, ok := weight(item, toNode)
	if !ok {
		panic("bellmanFord: unexpected invalid weight")
	}
	joint := wg + path.dist[k]
	var par ParPair
	par.from = item
	par.to = toNode
	par.weight = joint
	c <- par
}

type ParPair struct {
	from, to graph.Node
	weight   float64
}

const delta = 2

// DeltaStep Apply the delta-stepping algorihtm to Graph and return
// a shortest path tree.
//
// Note that this uses Shortest to make it easier for you,
// but you can use another struct if that makes more sense
// for the concurrency model you chose.
func DeltaStep(s graph.Node, g graph.Graph) Shortest {
	// i did use this site to get a visual of what to do: https://cs.iupui.edu/~fgsong/LearnHPC/sssp/deltaStep.html
	// some names were taken from the dijkstra implementation that was given
	// if the graph doesn't have the node, don't do anything
	if !g.Has(s) {
		return Shortest{from: s}
	}
	// set up
	var weight Weighting
	if wg, ok := g.(graph.Weighter); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	nodes := g.Nodes()
	path := newShortestFrom(s, nodes)
	var buckets [][]graph.Node
	// setup to explore the first node
	buckets = append(buckets, make([]graph.Node, 0))
	buckets[0] = append(buckets[0], s)

	for i := 0; i < len(buckets); i++ {
		var lightEdges []graph.Node
		var heavyEdges []Pair
		for {
			if len(buckets[i]) == 0 {
				break
			}
			var item = buckets[i][0]
			// remove from list
			copy(buckets[i][0:], buckets[i][1:])
			buckets[i][len(buckets[i])-1] = nil
			buckets[i] = buckets[i][:len(buckets[i])-1]
			for _, reachableNode := range g.From(item) {
				if g.Edge(item, reachableNode).Weight() <= delta {
					lightEdges = append(lightEdges, reachableNode)
				} else {
					// need a set for heavy edges (have to know the from and to)
					heavyEdges = append(heavyEdges, Pair{from: item, to: reachableNode})
				}
			}
			// explore lightEdges
			for _, lightNode := range lightEdges {
				// These variable names were used based off of the dijkstra.go implementation (except wg, which is w in dijkstra)
				j := path.indexOf[lightNode.ID()]
				k := path.indexOf[item.ID()]
				wg, ok := weight(item, lightNode)
				if !ok {
					panic("deltaStepping: unexpected invalid weight")
				}
				if wg < 0 {
					panic("deltaStepping: unexpected negative weight")
				}
				joint := wg + path.dist[k]
				if joint < path.dist[j] {
					path.set(j, joint, k)
					buckets = AddToBucket(buckets, delta, item, lightNode, g)
				}
			}
			// clear lightEdges
			lightEdges = nil
		}
		// "once the inner while loop terminates we explore all the heavy edges..."
		for _, pair := range heavyEdges {
			heavyNode := pair.to
			item := pair.from
			j := path.indexOf[heavyNode.ID()]
			k := path.indexOf[item.ID()]
			wg, ok := weight(item, heavyNode)
			if !ok {
				panic("deltaStepping: unexpected invalid weight")
			}
			if wg < 0 {
				panic("deltaStepping: unexpected negative weight")
			}
			joint := wg + path.dist[k]
			if joint < path.dist[j] {
				path.set(j, joint, k)
				buckets = AddToBucket(buckets, delta, item, heavyNode, g)
			}
		}
		// clear heavyEdges
		heavyEdges = nil
	}

	return path
}

// Pair is used for the heavy edge since we need a mapping for the nodes
type Pair struct {
	from, to graph.Node
}

// Dijkstra is run to make sure that the tests are correct.
func Dijkstra(s graph.Node, g graph.Graph) Shortest {
	return DijkstraFrom(s, g)
}
