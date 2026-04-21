package wrongCuts

import (
	"fmt"
	
	"context"
	"log"
	"time"
	"strings"

	genai "google.golang.org/genai"
	emulation "architecture-retrieval/architecture/emulation"
)

// Given a graph, suggests another potential graph that separates the services
// by business capabilities, if the original graph has wrong cuts.
func MitigateWrongCuts(graph string) (string, error) {
	// Get the suggested refactor from Gemini
	refactoredGraphJson, err := RefactorToBusinessCapabilities(graph)
	if err != nil {
		return "", fmt.Errorf("error refactoring graph: %v", err)
	}
	
	fmt.Printf("Refactored graph JSON:\n%s\n", refactoredGraphJson)

	// We must restart the system with the new graph.
	err = emulation.RestartEmulationWithNewGraph(refactoredGraphJson)
	if err != nil {
		return "", fmt.Errorf("error emulating refactored architecture: %v", err)
	}

	return refactoredGraphJson, nil
}


func RefactorToBusinessCapabilities(inputGraphJson string) (string, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	config := &genai.GenerateContentConfig{
		Temperature:      ptr(float32(0.0)),
		ResponseMIMEType: "application/json", // Note: Ensure it is 'Mime' not 'MIME' depending on SDK version
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(
					"You are an expert Software Architect. You will receive a JSON representing an abstracted software architecture currently divided by technical layers. " +
						"Suggest a realistic architecture refactored, that separates the services by BUSINESS CAPABILITIES (e.g., accounts, notifications, orders).\n\n" +
						"STRICT CONSTRAINTS:\n" +
						"1. OUTPUT: Return ONLY the JSON. No backticks, no prose.\n" +
						"2. NODES:\n" +
						"   - id: Unique and incremented based on existing count.\n" +
						"   - label: Must represent a business capability relevant to the system's purpose.\n" +
						"   - type: BasicNode, DatabaseNode, ESB, or APIGateway.\n" +
						"   - properties: ONLY present for BasicNode, ESB, or APIGateway. NOT present for DatabaseNode.\n" +
						"   - properties.language: MUST be one of: python, javascript, html, java.\n" +
						//TODO: WHEN we add traefik support, we can add the option also here.
						//"   - properties.language: MUST be one of: python, javascript, html, java, traefik (traefik only for APIGateway).\n" +
						"   - properties.orderOfMagnitudeOfFiles: Format '10^x' where x is 0-4.\n" +
						"   - properties.port: Omit unless specific; assume 8080 default (or 5432 for DB internal logic).\n" +
						"3. EDGES:\n" +
						"  - source and target: Must reference valid node ids.\n" +
						"  - endpoint: The endpoint of the call. \n" +
						"   - properties.callDefinitionInSource: \n" +
						"     - Standard: http://{service_name}:{port}{endpoint}\n" +
						"     - Database: jdbc:postgresql://{database-service_name}:{port(default 5432)}/mydb\n" +
						"   - properties.method: 'POST' or 'GET' for HTTP; 'SQL' for DatabaseNode targets.\n" +
						"4. SYSTEM CONTEXT:\n" +
						"   - servicesSeparatedByBusinessDomain: Must be set to true.\n" +
						"   - selfManagedLibraries: Update 'servicesUsingLibrary' array to match the new business-service labels.\n" +
						"5. REASONING: Ensure the services make sense for the software's purpose defined in any fields you can find useful for gathering that information.\n" +
						"6. STRUCTURE: This must follow the same structure as the input JSON, but you can modify any field as needed.\n" +
						"7. SIZE: The refactored architecture should have the same order of magnitude of nodes and edges as the original, but with improved separation by business capabilities.\n\n"),
			},
		},
	}

	prompt := fmt.Sprintf("Refactor the following architecture:\n\n%s", inputGraphJson)
	modelID := "gemini-3.1-flash-lite-preview"

	// --- RETRY LOGIC START ---
	maxRetries := 30

	for i := 0; i < maxRetries; i++ {
		result, err := client.Models.GenerateContent(ctx, modelID, genai.Text(prompt), config)
		
		if err == nil {
			if len(result.Candidates) == 0 {
				return "", fmt.Errorf("no refactoring generated")
			}
			return result.Text(), nil 
		}

		// Check if it's a 503/High Demand error
		if strings.Contains(err.Error(), "503") || strings.Contains(err.Error(), "UNAVAILABLE") {
			// If error, just forget it.
			log.Printf("Gemini is currently under high demand - 503 error. Will not try again.")
			break
		}

		//If its RESOURCE_EXHAUSTED, we can also retry, but with a fixed delay of 1 minute, as it seems to be more related to quota limits than to momentary high demand.
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
			log.Printf("Gemini quota exhausted. Retrying in 1 minute... (Attempt %d/%d)", i+1, maxRetries)
			time.Sleep(1 * time.Minute)
			continue
		}
	}
	// --- RETRY LOGIC END ---

	return "", fmt.Errorf("failed to refactor due to high demand")
}

func ptr[T any](v T) *T {
	return &v
}