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

	// NOTE: This approach is based on the idea that, if a service is calling a database with a WRITE method
	// then it is an innapropriate service intimacy.

	/*for _, node := range graphStruct.Nodes {
		// If a node is a db, we must check how many writes it has.
		if node.Type == "DatabaseNode" {
			var writeEdges []graphparsing.Edge
			for _, edge := range graphStruct.Edges {
				if edge.Target == node.Id && edge.Properties.Method == "POST" || edge.Properties.Method == "PUT" || edge.Properties.Method == "PATCH" {
					writeEdges = append(writeEdges, edge)
				}
			}
			// Given all write edges, check the node that perform the most writes.
			writeCount := make(map[string]int)
			for _, edge := range writeEdges {
				writeCount[edge.Source]++
			}
			// Sort the writeCount map by value and get the node with the most writes.
			type kv struct {
				Key string
				Value int
			}
			var ss []kv
			for k, v := range writeCount {
				ss = append(ss, kv{k, v})
			}

			sort.Slice(ss, func(i, j int) bool {
				return ss[i].Value > ss[j].Value
			})

			owner := ss[0].Key

			if(len(ss) > 1 && ss[0].Value == ss[1].Value) {
				// If there are more than one node with the same number of writes, then we cannot be sure who is the owner, so we skip this db.
				continue
			}

			// If there is a node that:
			// - Performs a WRITE operation on the db.
			// (We are unconsidering read call) 
			
			// Then, any other node that performs a WRITE operation on the same db is an innapropriate service intimacy.
			for _, edge := range writeEdges {
				if edge.Source != owner {
					sourceNode, err := graphparsing.GetNodeById(graphStruct, edge.Source)
					targetNode, err := graphparsing.GetNodeById(graphStruct, edge.Target)
					if err != nil {
						return nil, fmt.Errorf("Error getting nodes: %s", err)
					}
					innapropriateServiceIntimacity = append(innapropriateServiceIntimacity, fmt.Sprintf("Service %s is calling database %s owned by service %s", sourceNode.Label, targetNode.Label, owner))
				}
			}
			
		}
	}*/



	//NOTE: This a approach that focus on see if there are databases contains other
	//services name, and then see if other services are calling that database. This is a very naive approach, but it can work for some cases.
	// We are still uncertain which one to use.
	
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