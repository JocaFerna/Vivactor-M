package sharedPersistency

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
)

// Given a graph, detects if there are services that access the same db.
func GetSharedPersistency(graph string) ([][]string, error){

	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, err
	}
	
	var sharedPersistency [][]string

	// Map to keep track of which services access which databases	
	dbAccessMap := make(map[graphparsing.Node]map[string]struct{})


	for i, edge := range graphStruct.Edges {
		targetNode, err := graphparsing.GetNodeById(graphStruct, edge.Target)
		if err != nil || targetNode.Type != "DatabaseNode" {
			continue
		}
		for j := i + 1; j < len(graphStruct.Edges); j++ {
			otherEdge := graphStruct.Edges[j]
			otherTargetNode, err := graphparsing.GetNodeById(graphStruct, otherEdge.Target)
			if err != nil || otherTargetNode.Type != "DatabaseNode" {
				continue
			}
			if targetNode.Id == otherTargetNode.Id {
				sourceNode, err := graphparsing.GetNodeById(graphStruct, edge.Source)
				if err != nil {
					continue
				}
				otherSourceNode, err := graphparsing.GetNodeById(graphStruct, otherEdge.Source)
				if err != nil {
					continue
				}
				if sourceNode.Type == "BasicNode" && otherSourceNode.Type == "BasicNode" {
					if _, exists := dbAccessMap[targetNode]; !exists {
						dbAccessMap[targetNode] = make(map[string]struct{})
					}
					dbAccessMap[targetNode][sourceNode.Label] = struct{}{}
					dbAccessMap[targetNode][otherSourceNode.Label] = struct{}{}
				}
			}
		}

		
	}
	// Now, we can retrieve the shared persistency information from the map
	for _, services := range dbAccessMap {
		if len(services) > 1 {
			var serviceList []string
			for service := range services {
				serviceList = append(serviceList, service)
			}
			sharedPersistency = append(sharedPersistency, serviceList)
		}
	}
	return sharedPersistency, nil
}