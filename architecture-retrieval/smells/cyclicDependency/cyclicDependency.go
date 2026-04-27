package cyclicDependency

import (
	// Graph Parsing
	graphparsing "architecture-retrieval/architecture/graphParsing"
	
	"slices"
	"sort"
)

// Given a graph, detects if there are cyclic dependencies.
// If there are, it returns a list of all the cycles in the graph. Each cycle is represented as a list of node labels.
func DetectCyclicDependency(graph string) ([][]string, error) {

	// Parse the graph and create an adjacency list
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, err
	}


	// Build the Adjacency List using the stable pointers
    // We use the index 'i' to get the address of the node inside the slice
    adjList := make(map[*graphparsing.Node][]*graphparsing.Node)
    
    for i := range graphStruct.Nodes {
        nodePtr := &graphStruct.Nodes[i]
        
        // Call the function using the node's ID
        neighbors, err := graphparsing.GetAdjacentNodes(graphStruct, nodePtr.Id)
        if err != nil {
            return nil, err
        }
        adjList[nodePtr] = neighbors
    }

	// Detect Cycles using DFS
	var cycles [][]string
	for node := range adjList {
		// Visited map with everything set to false
		visited := make(map[*graphparsing.Node]bool)
		currentCycles, found := dfs(adjList, node, visited, node, []*graphparsing.Node{}, [][]*graphparsing.Node{})
		if found {
			for _, cycle := range currentCycles {
				var cycleNames []string
				for _, nodeInCycle := range cycle {
					cycleNames = append(cycleNames, nodeInCycle.Label)
				}
				sort.Strings(cycleNames) // Sort the cycle names to ensure consistent ordering
				if !containsSlice(cycles, cycleNames) {
					cycles = append(cycles, cycleNames)
				}
			}
		}
	}

	return cycles, nil
}

// Check if a slice of string slices contains a specific slice of strings 
// (used to avoid duplicate cycles)
func containsSlice(slice [][]string, element []string) bool {
    for _, v := range slice {
        if slices.Equal(v, element) {
            return true
        }
    }
    return false
}

// DFS Function to extract cycles
// adjList: The adjacency list of the graph
// node: The current node being visited
// visited: A map to keep track of visited nodes
// startNode: The node where the DFS started (used to detect cycles)
// nodeStack: A stack to keep track of the current path in the DFS
// currentCycles: A list of cycles found so far
func dfs(adjList map[*graphparsing.Node][]*graphparsing.Node, node *graphparsing.Node, visited map[*graphparsing.Node]bool ,startNode *graphparsing.Node, nodeStack []*graphparsing.Node, currentCycles [][]*graphparsing.Node) ([][]*graphparsing.Node, bool) {
	
	// We do not consider API Gateway nodes as part of the cycles, since they are not really a cyclic dependency, but rather a design choice.
	if node.Type == "APIGateway" {
		return currentCycles, false
	}
	
	if visited[node]{
		if node == startNode {
			return append(currentCycles, nodeStack), true
		}
		return currentCycles, false
	}
	
	areThereCycles := false
	visited[node] = true
	
	for _, neighbor := range adjList[node] {
		newCurrentCycles, found := dfs(adjList, neighbor, visited, startNode, append(nodeStack, node), currentCycles)
		if found {
			currentCycles = newCurrentCycles
			areThereCycles = true
		}
	}
	visited[node] = false
	return currentCycles, areThereCycles
}