package kill

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"os"
	"os/exec"
	emulation "architecture-retrieval/architecture/emulation"
)

// Given a graph, it will kill the architecture totally.
func KillArchitecture(graph string) (string, error) {
	// Parse the graph
	parsedGraph, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}
	
	// Get base path based on the graph
	basePath := graphparsing.GetBasePathOfGraph(parsedGraph)

	// 1. Clean project lock
	err = emulation.CleanProjectLock(basePath)
	if err != nil {
		return "", err
	}

	// 2. "docker compose down" is the clean way. 
	// Adding -v removes volumes, --remove-orphans catches stragglers.
	downCmd := exec.Command("docker", "compose", "down", "-v", "--remove-orphans")
	downCmd.Dir = basePath
	downCmd.Stderr = os.Stderr
	if err := downCmd.Run(); err != nil {
		return "", err
	}

	// Serialize the graph back to string
	updatedGraph, err := graphparsing.SerializeGraph(parsedGraph)
	if err != nil {
		return "", err
	}
	return updatedGraph, nil
}