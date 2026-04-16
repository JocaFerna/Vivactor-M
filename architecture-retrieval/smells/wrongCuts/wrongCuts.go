package wrongCuts

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"fmt"
	"strings"
	"time"

	"context"
	"log"

	"google.golang.org/genai"
)

// Given a graph, detects if there are wrong cuts.
func GetWrongCuts(graph string) (bool, error) {
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return false, err
	}
	

	

	// Assert the smell.
	smell, err := AssertWrongCutsSmell(graphStruct.System.Name, graphStruct.Nodes)
	if err != nil {
		return false, fmt.Errorf("error asserting wrong cuts smell: %v", err)
	}
	if smell == "TECHNICAL LAYERS" {
		log.Printf("Gemini classified the system as having TECHNICAL LAYERS. This indicates a potential wrong cuts smell.")
		return true, nil
	} else if smell == "BUSINESS CAPABILITIES" {
		log.Printf("Gemini classified the system as having BUSINESS CAPABILITIES. This indicates a good partitioning with no wrong cuts smell.")
		return false, nil
	} else {
		log.Printf("Gemini returned an UNKNOWN classification. Falling back to graph property for wrong cuts smell detection.")
	}
	return false, err 
}

// AssertWrongCutsSmell asks Gemini to classify the architecture partitioning.
func AssertWrongCutsSmell(systemName string, nodes []graphparsing.Node) (string, error) {
	ctx := context.Background()

	// 1. Initialize the client (requires GEMINI_API_KEY environment variable)
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// 2. Prepare the list of service names
	var serviceNames []string
	for _, n := range nodes {
		if n.Type == "BasicNode" {
			serviceNames = append(serviceNames, n.Label)
		}
	}
	serviceList := strings.Join(serviceNames, ", ")

	// 3. Define the deterministic Configuration (Updated for 2026)
	config := &genai.GenerateContentConfig{
		// Ensure Temperature is float64
		Temperature: ptr(float32(0.0)), 
		
		// SystemInstruction expects a *genai.Content object.
		// In this SDK, you wrap Parts (which are created from text) inside Content.
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(
					"You are an expert Software Architect. Analyze the service names. " +
					"Rules: \n" +
					"1. If services are technical tiers (UI, Backend, Logic), output: TECHNICAL LAYERS.\n" +
					"2. If services are domains (Orders, Shipping, Inventory), output: BUSINESS CAPABILITIES.\n" +
					"3. Provide ONLY one of those two phrases. No explanation.\n" +
					"4. If you find both technical and business services, classify based on the majority. If it's a tie, classify as TECHNICAL LAYERS.\n" +
					"Examples:\n" +
					"- System Name: E-commerce | Services: [UI, Backend, Database] => TECHNICAL LAYERS\n" +
					"- System Name: E-commerce | Services: [Orders, Shipping, Inventory] => BUSINESS CAPABILITIES\n" +
					"- System Name: E-commerce | Services: [UI, Orders, Shipping] => TECHNICAL LAYERS\n" +
					"- System Name: E-commerce | Services: [Inventory, Database] => BUSINESS CAPABILITIES",
							),
			},
		},
	}

	// Use a 2026-current model
	modelID := "gemini-3.1-flash-lite-preview"
	// 4. Construct the prompt
	prompt := fmt.Sprintf("System Name: %s | Services: [%s]", systemName, serviceList)

	// --- RETRY LOGIC START ---
	maxRetries := 30
	baseDelay := 5 * time.Second

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
			// Exponential backoff: 2s, 4s, 8s, 16s...
			delay := baseDelay * time.Duration(1<<i) 
			log.Printf("Gemini is busy (503). Retrying in %v... (Attempt %d/%d)", delay, i+1, maxRetries)
			
			time.Sleep(delay)
			continue
		}

		//If its RESOURCE_EXHAUSTED, we can also retry, but with a fixed delay of 1 minute, as it seems to be more related to quota limits than to momentary high demand.
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
			log.Printf("Gemini quota exhausted. Retrying in 1 minute... (Attempt %d/%d)", i+1, maxRetries)
			time.Sleep(1 * time.Minute)
			continue
		}

		// If it's a different error (like a 400), don't bother retrying
		return "", fmt.Errorf("gemini call failed with non-retryable error: %w", err)
	}
	// --- RETRY LOGIC END ---

	return "", fmt.Errorf("failed to assert wrong cuts smell after %d attempts due to high demand", maxRetries)
}

// Helper to handle pointer-based config values
func ptr[T any](v T) *T {
	return &v
}