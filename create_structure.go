package main

import (
	"os"
	"path/filepath"
)

func main() {
	// Define the root project directory
	rootDir := "."

	// Define the structure
	structure := map[string][]string{
		rootDir: {
			"cmd/server/main.go",
			"internal/auth/auth.go",
			"internal/models/models.go",
			"internal/service/service.go",
			"pkg/utils/utils.go",
			"api/api_spec.yaml",
			"configs/config.yaml",
			"docs/README.md",
			"test/test.go",
			"go.mod",
			"LICENSE",
		},
	}

	// Create the directories and files
	for dir, files := range structure {
		os.MkdirAll(dir, os.ModePerm)
		for _, file := range files {
			filePath := filepath.Join(dir, file)
			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				panic(err)
			}
			if _, err := os.Create(filePath); err != nil {
				panic(err)
			}
		}
	}

	// Optionally, write initial content to go.mod
	goModContent := "module example.com/my_go_project\n\ngo 1.16\n"
	os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(goModContent), 0644)
}
