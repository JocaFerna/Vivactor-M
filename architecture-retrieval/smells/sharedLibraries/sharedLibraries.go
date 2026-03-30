package sharedLibraries

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
)

// Given a graph, detects if there are shared libraries.
func GetSharedLibraries(graph string) ([]string, error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, err
	}

	var sharedLibraries []string

	libraries := graphStruct.System.SelfManagedLibraries
	for _, library := range libraries {
		if len(library.ServicesUsingLibrary) > 1 {
			sharedLibraries = append(sharedLibraries, library.Name)
		}
	}
	return sharedLibraries, nil
}