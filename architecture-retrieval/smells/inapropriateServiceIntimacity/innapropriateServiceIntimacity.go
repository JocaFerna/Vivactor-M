package inapropriateServiceIntimacity

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"fmt"
	"strings"
)

// Given a graph, it retrieves all service intimacies.
// An innapropriate service intimacy is when
// The microservice keeps on connecting to
// private data from other services instead of dealing
// with its own data. Basically, calls to another service's db.
// This is particularly not intuitive to detect
// as we cannot be sure who exactly "owns" the data,
// and also, if the database is shared, then it is not a problem.
func GetInnapropriateServiceIntimacity(graph string) ([]string, error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, err
	}

	var innapropriateServiceIntimacity []string
	for _, edge := range graphStruct.Edges {
		// Get source and target nodes
		sourceNode, err := graphparsing.GetNodeById(graphStruct, edge.Source)
		targetNode, err := graphparsing.GetNodeById(graphStruct, edge.Target)
		if err != nil {
			return nil, fmt.Errorf("Error getting nodes: %s", err)
		}

		// Check if targetNode is a database and sourceNode is a service
		// Now, we need to check if the database is "owned" by the service or not. 
		// We will assume that, if somehow the name of the database contains the name of the service, then it is owned by that service. 
		// This is a very naive approach, but it can work for some cases.
		if targetNode.Type == "DatabaseNode" && sourceNode.Type == "BasicNode" && strings.Contains(targetNode.Label, sourceNode.Label) {
			
			// Now that we found a db that is owned by a service, any other services that call
			// that db are innapropriate service intimacies.
			for _, otherEdge := range graphStruct.Edges {
				otherSourceNode, err := graphparsing.GetNodeById(graphStruct, otherEdge.Source)
				otherTargetNode, err := graphparsing.GetNodeById(graphStruct, otherEdge.Target)
				if err != nil {
					return nil, fmt.Errorf("Error getting nodes: %s", err)
				}

				if otherTargetNode.Id == targetNode.Id && otherSourceNode.Id != sourceNode.Id && otherSourceNode.Type == "BasicNode" {
					innapropriateServiceIntimacity = append(innapropriateServiceIntimacity, fmt.Sprintf("Service %s is calling database %s owned by service %s", otherSourceNode.Label, targetNode.Label, sourceNode.Label))
				}
			}
		}
			
	}
	return innapropriateServiceIntimacity, nil
}