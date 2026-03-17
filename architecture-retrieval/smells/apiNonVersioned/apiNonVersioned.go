package apiNonVersioned

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"log"

	"architecture-retrieval/smells/utils"
	"github.com/hashicorp/go-set/v3"
)

func DetectApiNonVersioned(repoName string) ([]string, error) {
	var nonAPIVersionedSmells []string
	cleanName := strings.TrimPrefix(repoName, "/")
	repoName = filepath.Join("/api/downloads", cleanName)
	files, err := utils.GetFilesInDirectory(repoName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving files: %v", err)
	}
	

	for _, file := range files {
		if filepath.Ext(file) == ".java" {
			
			// Search for RequestMapping annotations without versioning in the path
			content, err := utils.ReadFileContent(file)
			if err != nil {
				fmt.Printf("   - Warning: Could not read file %s: %v\n", file, err)
				continue
			}
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line := strings.ReplaceAll(line, " ", "")
				
				if strings.Contains(line, "@RequestMapping") {
					log.Printf("Line w/@RequestMapping: ",line)
					if strings.Contains(line, "value=") {
						log.Printf("Line w/value=: ",line)
						// Extract the path value
						start := strings.Index(line, "value=") + len("value=") + 1 // Skip the \"
						end := strings.Index(line[start:], "\"") // Find the closing \"
						if end > 0 {
							pathValue := line[start : start+end]
							if isApiNonVersioned(pathValue) && pathValue != "/" {
								nonAPIVersionedSmells = append(nonAPIVersionedSmells, pathValue)
								fmt.Printf("   - Detected API Non-Versioned in %s: %s\n", file, pathValue)
							}
						}
					}
				}
			}
		} else if filepath.Ext(file) == ".yml" || filepath.Ext(file) == ".yaml"{
			// Search for RequestMapping annotations without versioning in the path
			content, err := utils.ReadFileContent(file)
			if err != nil {
				fmt.Printf("   - Warning: Could not read file %s: %v\n", file, err)
				continue
			}
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line := strings.ReplaceAll(line, " ", "")

				if line == ""{ continue }

				// It can be in any line tbf, so we don't have many rules:
				// 1 - The found string must come after a ":"
				// 2 - Must have a / (may have another /, if not, take till the end.)
				if strings.Index(line,"/") > strings.Index(line,":") && strings.Contains(line,"/") {

					// We must address the case the https:// case.
					if (strings.Contains(line,"http://") || strings.Contains(line,"https://")){
						if (strings.Index(line,"/")+1) == strings.LastIndex(line,"/") {
							continue
						}
						// Ok, we have a endpoint here, so, let's take it from the third slash, as the first two are for the protocol.
						count_of_slashes := 0
						for i, char := range line {
							if char == '/' {
								count_of_slashes++
								if count_of_slashes == 3 {
									line = line[i:] // Take the substring from the third slash onward
									break
								}
							}
						}
					}
					// As it exists, we take the endpoint till the last index.
					substr := line[strings.Index(line,"/"):]
					if isApiNonVersioned(substr) && substr != "/" {
						// Remove any characters not intended at the end of the substr
						substr = strings.ReplaceAll(substr, "(", "")
						substr = strings.ReplaceAll(substr, ")", "")
						substr = strings.ReplaceAll(substr, "'", "")
						substr = strings.ReplaceAll(substr, "\"", "")
						substr = strings.ReplaceAll(substr, ",", "")
						substr = strings.ReplaceAll(substr, "`", "")
						nonAPIVersionedSmells = append(nonAPIVersionedSmells, substr)
						fmt.Printf("   - Detected API Non-Versioned in %s: %s\n", file, substr)
					}

					//}
						
				}

			}
		}
	}
	nonAPIVersionedSmells = set.From[string](nonAPIVersionedSmells).Slice()
	log.Println("nonAPIVersionedSmells: ",nonAPIVersionedSmells)
	return nonAPIVersionedSmells, nil
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
	