package apiNonVersioned

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"fmt"
	
	"regexp"
)

func DetectApiNonVersioned(graph string) ([]string, error) {
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %w", err)
	}

	var nonVersionedApis []string
	for _, edge := range graphStruct.Edges {
		if isApiNonVersioned(edge.Properties.CallDefinitionInSource) {
			nonVersionedApis = append(nonVersionedApis, edge.Properties.CallDefinitionInSource)
		}
	}
	return nonVersionedApis, nil
}

func isApiNonVersioned(path string) bool {
	a := "/?api/v[0-9]+.*"
	b := "/?v[0-9]+.*"
	// See if it matches the regex pattern for versioned APIs
	matched, _ := regexp.MatchString(a, path)
	if matched {
		return false
	}
	matched, _ = regexp.MatchString(b, path)
	return !matched
}
	