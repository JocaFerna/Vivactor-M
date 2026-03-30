package tooManyStandards

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
)

// Given a graph, detects how many different languages are used.
func GetTooManyStandards(graph string) (int, error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return 0, err
	}
	
	languageSet := make(map[string]struct{})
	for _, node := range graphStruct.Nodes {
		if node.Type == "BasicNode" && node.Properties.Language != "" {
			if _, exists := languageSet[node.Properties.Language]; !exists {
				languageSet[node.Properties.Language] = struct{}{}
			}
			
		}
	}
	return len(languageSet), nil
}
