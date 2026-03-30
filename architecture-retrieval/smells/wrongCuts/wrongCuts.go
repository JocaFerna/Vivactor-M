package wrongCuts

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
)

// Given a graph, detects if there are wrong cuts.
func GetWrongCuts(graph string) (bool, error) {
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return false, err
	}

	// Actually, we just check if the bool of ServicesSeparatedByBusinessDomain
	// is true or false. If it is false, then there are wrong cuts.

	// TODO: An idea of implementation would be to, present the labels.
	// Of the services to an AI, that would judge if they are
	// separated by business domain or not. This would be a more accurate way,
	// but it would require a lot of work to implement and train it.
	// To discuss with the supervisor.
	return !graphStruct.System.ServicesSeparatedByBusinessDomain, nil
}