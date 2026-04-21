package microserviceGreedy

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
	"log"
	"strings"
)

// Given a graph, refactor services that only serve one or few static html pages into a single service.
// The refactor approach select bases itself on merging each greedy service with the service that has the most interactions with it.
// It returns the graph
func RefactorMicroserviceGreedy(graph string, greedyMicroservices []string) (string, error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}
	// Get base path of the project from the graph.
	basePath := utils.GetBasePathOfGraph(graphStruct)

	var refactoredMicroservices []string
	for _, greedyMicroservice := range greedyMicroservices {
		// Remove \" from the microservice name if it is present.
		greedyMicroservice = strings.TrimPrefix(greedyMicroservice, "\"")
		greedyMicroservice = strings.TrimSuffix(greedyMicroservice, "\"")
		log.Printf("Refactoring greedy microservice: %s\n", greedyMicroservice)
		greedyNode, err := graphparsing.GetNodeByLabel(graphStruct, greedyMicroservice)
		if err != nil {
			log.Printf("No node is found!")
			continue
		}
		// Get the service with the most interactions with the greedy microservice.
		var serviceInteractions = make(map[string]int)
		for _, edge := range graphStruct.Edges {
			if edge.Source == greedyNode.Id || edge.Target == greedyNode.Id {
				log.Printf("Edge found for greedy microservice: %s -> %s\n", edge.Source, edge.Target)
				if edge.Source == greedyNode.Id {
					serviceInteractions[edge.Target]++
				} else {
					serviceInteractions[edge.Source]++	
				}
			}
		}
		var maxInteractions int
		var serviceToMerge string
		for service, interactions := range serviceInteractions {
			if interactions > maxInteractions {
				maxInteractions = interactions
				serviceToMerge = service
			}
		}
		log.Printf("Service with most interactions: %s with %d interactions\n", serviceToMerge, maxInteractions)
		// Get the node of the service to merge.
		serviceNode, err := graphparsing.GetNodeById(graphStruct, serviceToMerge)
		if err != nil {
			log.Printf("No node is found!")
			continue
		}
		if err = utils.TransferFilesIntoAnotherService(greedyNode,serviceNode,basePath); err != nil {
			log.Printf("Error transferring files: %v\n", err)
			continue
		}
		if err = utils.RemoveService(greedyNode, graphStruct, basePath); err != nil {
			log.Printf("Error removing service: %v\n", err)
			continue
		}
		// Remove call to the greedy microservice from the source code of the service to merge.
		if err = utils.RemoveCallToService(serviceNode, greedyNode, basePath); err != nil {
			log.Printf("Error removing call to service: %v\n", err)
			continue
		}
		// Merge the greedy microservice with the service with the most interactions.
		graphStruct = graphparsing.MergeNodeIntoAnother(graphStruct, greedyNode, serviceNode)
		refactoredMicroservices = append(refactoredMicroservices, serviceNode.Label)
	}

	// Serialize the graph back to JSON
	graph, err = graphparsing.SerializeGraph(graphStruct)
	if err != nil {
		return "", err
	}
	return graph, nil
}
