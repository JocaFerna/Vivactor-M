package tooManyStandards

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
	"fmt"
)

func MitigateTooManyStandardsSmell(graph string) (string, error) {
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}

	// Check the language least used by the services in the graph.
	languageCount := make(map[string]int)
	for _, node := range graphStruct.Nodes {
		if node.Type == "BasicNode" {
			lang := node.Properties.Language
			if lang != "" {
				languageCount[lang]++
			}
		}			
	}
	// Find the least used language.
	leastUsedLanguage := ""
	minCount := int(^uint(0) >> 1)
	for lang, count := range languageCount {
		if count < minCount {
			minCount = count
			leastUsedLanguage = lang
		}
	}

	// Get the most used language, to convert the services using the least used language to the most used one.
	mostUsedLanguage := ""
	maxCount := 0
	for lang, count := range languageCount {
		if count > maxCount {
			maxCount = count
			mostUsedLanguage = lang
		}
	}

	newNodes := make([]graphparsing.Node, 0)
	// Convert all services using the least used language to the most used one.
	for i, node := range graphStruct.Nodes {
		if node.Type == "BasicNode" && node.Properties.Language == leastUsedLanguage {
			graphStruct.Nodes[i].Properties.Language = mostUsedLanguage
			newNodes = append(newNodes, graphStruct.Nodes[i])
		}
	}
	serializedGraph, err := graphparsing.SerializeGraph(graphStruct)
	if err != nil {
		return "", err
	}
	// Get base path of the project from the graph.
	basePath := utils.GetBasePathOfGraph(graphStruct)
	// Restart the architecture emulation with the new graph structure.
	err = utils.RedoServicesWithNewLanguage(newNodes, graphStruct, basePath)
	if err != nil {
		return "", fmt.Errorf("error emulating refactored architecture: %v", err)
	}
	return serializedGraph, nil
}