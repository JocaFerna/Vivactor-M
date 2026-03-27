package esbUsage

import (
	// Graph Parsing
	graphparsing "architecture-retrieval/architecture/graphParsing"
)

// Given a graph, check if some node is of type "ESBNode"
func DetectESBUsage(graph string) (bool, error) {
	
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return false, err
	}
	
	for _, node := range graphStruct.Nodes {
		if node.Type == "ESB" {
			return true, nil
		}
	}
	return false, nil
}