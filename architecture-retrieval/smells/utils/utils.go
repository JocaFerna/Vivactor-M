package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetFilesInDirectory(dirPath string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Do not consider files in "target" and "test" directories
		// Only consider java and yaml files (config files)
		if !info.IsDir() && (filepath.Ext(path) == ".java" || filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".js") && !strings.Contains(path, "/target/") && !strings.Contains(path, "/test/") {
			files = append(files, path)
		}
		return nil
	},
	)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", dirPath, err)
	}
	return files, nil
}

func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filePath, err)
	}
	return string(content), nil
}