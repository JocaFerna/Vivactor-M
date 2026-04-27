package noAPIGateway

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
)

// Given a graph, detects if there is a service qualified as
// an API Gateway.
// True -> There is not an API Gateway 
// False -> There is an API Gateway OR there are more than 50 services, so an API Gateway is not necessary.
// Also, note that, if there are less than 50 services,
// then it is not necessary to have an API Gateway.
func GetNoAPIGateway(graph string) (bool, error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return true, err
	}
	
	// Check if there are less than 50 services
	if len(graphStruct.Nodes) < 50 {
		return false, nil
	}

	// Check if there is an API Gateway
	for _, node := range graphStruct.Nodes {
		if node.Type == "APIGateway" {
			return false, nil
		}
	}

	// In case it not finds, there is no API Gateway, so we return true. 
	return true, nil
}