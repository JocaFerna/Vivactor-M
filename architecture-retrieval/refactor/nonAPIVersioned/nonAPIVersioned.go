package nonAPIVersioned
import (
	"fmt"
	"path/filepath"
	"strings"

	"architecture-retrieval/refactor/utils"
)

func MitigateNonAPIVersionedSmells(repoName string, nonAPIVersionedSmells []string) (error) {
	
	cleanName := strings.TrimPrefix(repoName, "/")
	repoName = filepath.Join("/api/downloads", cleanName)
	files, err := utils.GetFilesInDirectory(repoName)
	if err != nil {
		return fmt.Errorf("error retrieving files: %v", err)
	}
	

	error_detected := false
	for _, file := range files {
			
			// Search for RequestMapping annotations without versioning in the path
			content, err := utils.ReadFileContent(file)
			if err != nil {
				fmt.Printf("   - Warning: Could not read file %s: %v\n", file, err)
				error_detected = true
				continue
			}
			// Replace all occurrences of the non-versioned paths with versioned ones
			for _, smell := range nonAPIVersionedSmells {
				content = strings.ReplaceAll(content, smell, "/v1"+smell)
			}
			// Write the updated content back to the file
			err = utils.WriteFileContent(file, content)
			if err != nil {
				fmt.Printf("   - Warning: Could not write to file %s: %v\n", file, err)
				error_detected = true
				continue
			}
		
	}
	if error_detected {
		return fmt.Errorf("errors detected while processing files")
	}
	return nil
}