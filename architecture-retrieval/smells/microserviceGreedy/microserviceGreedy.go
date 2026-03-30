package microserviceGreedy

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	
)

// Given a graph, detects if there are services that only serve one or few
// static html pages.
func GetMicroserviceGreedy(graph string) ([]string, error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, err
	}
	
	var microserviceGreedy []string
	for _, node := range graphStruct.Nodes {
		exponent := graphparsing.GetOrderOfMagnitudeOfFiles(node)
		if node.Type == "BasicNode" && node.Properties.Language == "html" && exponent <= 1 {
			microserviceGreedy = append(microserviceGreedy, node.Label)
		}
	}
	return microserviceGreedy, nil
}