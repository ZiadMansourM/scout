package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	REPO_DIR                  = "/Users/ziadh/Desktop/playgroud/go/api/budgetly/"
	OUTPUT_FILE_SUFFIX        = "_output.txt"
	GITIGNORE_FILENAME        = ".gitignore"
	MAX_LINES_FOR_LARGE_FILES = 300
)

var (
	REPO_NAME   string
	OUTPUT_FILE string
	SCRIPT_NAME string

	BINARY_EXTENSIONS = map[string]bool{
		".webp": true, ".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".pdf": true, ".zip": true, ".exe": true,
		".ico": true, ".svg": true, ".pyc": true,
	}

	IGNORE_FILES = map[string]bool{
		"package-lock.json": true, "yarn.lock": true, ".DS_Store": true,
		"Thumbs.db": true, ".gitattributes": true, ".eslintcache": true,
		".npmrc": true, ".yarnrc": true,
	}

	IGNORE_DIRS = map[string]bool{
		"__pycache__": true, ".git": true,
	}

	TRUNCATE_EXTENSIONS = map[string]bool{
		".json": true, ".geojson": true,
	}
)

func init() {
	REPO_NAME = filepath.Base(REPO_DIR)
	OUTPUT_FILE = REPO_NAME + OUTPUT_FILE_SUFFIX
	SCRIPT_NAME = filepath.Base(os.Args[0])
}

func handleError(errorMessage string, exitProgram bool) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", errorMessage)
	if exitProgram {
		fmt.Fprintln(os.Stderr, "Exiting program due to error.")
		os.Exit(1)
	}
}

func parseGitignore(filePath string) func(string) bool {
	patterns := []string{}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		handleError(fmt.Sprintf("Failed to parse .gitignore file: %v", err), false)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	return func(file string) bool {
		for _, pattern := range patterns {
			regexPattern := regexp.QuoteMeta(pattern)
			regexPattern = strings.ReplaceAll(regexPattern, `\*`, `.*`)
			regexPattern = strings.ReplaceAll(regexPattern, `\?`, `.`)
			matched, _ := regexp.MatchString("^"+regexPattern+"$", file)
			if matched {
				return true
			}
		}
		return false
	}
}

func isFileEmpty(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}
	return info.Size() == 0
}

func shouldIncludeFile(filePath string, gitignore func(string) bool, includeEmpty bool) bool {
	fileName := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))

	if fileName == OUTPUT_FILE || fileName == SCRIPT_NAME || fileName == GITIGNORE_FILENAME {
		return false
	}
	if IGNORE_FILES[fileName] {
		return false
	}
	if gitignore != nil && gitignore(filePath) {
		return false
	}
	if !includeEmpty && isFileEmpty(filePath) {
		return false
	}
	if BINARY_EXTENSIONS[ext] {
		return false
	}
	return true
}

func getRelevantFiles(repoDir string, gitignore func(string) bool) []string {
	relevantFiles := []string{}

	filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			handleError(fmt.Sprintf("Error accessing %s: %v", path, err), false)
			return nil
		}

		if info.IsDir() {
			if IGNORE_DIRS[info.Name()] || (gitignore != nil && gitignore(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIncludeFile(path, gitignore, false) {
			relevantFiles = append(relevantFiles, path)
		}
		return nil
	})

	return relevantFiles
}

func getFolderStructure(startPath string, gitignore func(string) bool) []string {
	structure := []string{REPO_NAME + string(os.PathSeparator)}

	var addToStructure func(string, string)
	addToStructure = func(currentPath, prefix string) {
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			handleError(fmt.Sprintf("Error reading directory %s: %v", currentPath, err), false)
			return
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name() < entries[j].Name()
		})

		filteredEntries := []os.DirEntry{}
		for _, entry := range entries {
			if IGNORE_DIRS[entry.Name()] ||
				entry.Name() == OUTPUT_FILE ||
				entry.Name() == SCRIPT_NAME ||
				entry.Name() == GITIGNORE_FILENAME ||
				IGNORE_FILES[entry.Name()] ||
				(gitignore != nil && gitignore(filepath.Join(currentPath, entry.Name()))) {
				continue
			}
			filteredEntries = append(filteredEntries, entry)
		}

		for i, entry := range filteredEntries {
			isLast := i == len(filteredEntries)-1
			var connector string
			if isLast {
				connector = "└── "
			} else {
				connector = "├── "
			}

			if entry.IsDir() {
				structure = append(structure, prefix+connector+entry.Name()+string(os.PathSeparator))
				var extension string
				if isLast {
					extension = "    "
				} else {
					extension = "│   "
				}
				addToStructure(filepath.Join(currentPath, entry.Name()), prefix+extension)
			} else {
				structure = append(structure, prefix+connector+entry.Name())
			}
		}
	}

	addToStructure(startPath, "")
	return structure
}

func countTotalLines(relevantFiles []string) int {
	totalLines := 0
	for _, file := range relevantFiles {
		fileLines := 0
		f, err := os.Open(file)
		if err != nil {
			handleError(fmt.Sprintf("Error counting lines in %s: %v", file, err), false)
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fileLines++
		}
		totalLines += fileLines
		f.Close()
	}
	return totalLines
}

func countLinesPerFileType(relevantFiles []string) map[string]int {
	linesPerType := make(map[string]int)
	for _, file := range relevantFiles {
		ext := filepath.Ext(file)
		fileLines := 0
		f, err := os.Open(file)
		if err != nil {
			handleError(fmt.Sprintf("Error counting lines in %s: %v", file, err), false)
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fileLines++
		}
		linesPerType[ext] += fileLines
		f.Close()
	}
	return linesPerType
}

func countFilesPerType(relevantFiles []string) map[string]int {
	filesPerType := make(map[string]int)
	for _, file := range relevantFiles {
		ext := filepath.Ext(file)
		filesPerType[ext]++
	}
	return filesPerType
}

func calculateAverageFileSize(relevantFiles []string) float64 {
	totalSize := int64(0)
	for _, file := range relevantFiles {
		info, err := os.Stat(file)
		if err != nil {
			handleError(fmt.Sprintf("Error getting size of %s: %v", file, err), false)
			continue
		}
		totalSize += info.Size()
	}
	if len(relevantFiles) == 0 {
		return 0
	}
	return float64(totalSize) / float64(len(relevantFiles))
}

type LargestFile struct {
	Name  string
	Size  int64
	Lines int
}

func findLargestFile(relevantFiles []string) LargestFile {
	var largestFile LargestFile
	for _, file := range relevantFiles {
		info, err := os.Stat(file)
		if err != nil {
			handleError(fmt.Sprintf("Error processing largest file %s: %v", file, err), false)
			continue
		}
		if info.Size() > largestFile.Size {
			fileLines := 0
			f, err := os.Open(file)
			if err != nil {
				handleError(fmt.Sprintf("Error processing largest file %s: %v", file, err), false)
				continue
			}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				fileLines++
			}
			f.Close()
			largestFile = LargestFile{Name: info.Name(), Size: info.Size(), Lines: fileLines}
		}
	}
	return largestFile
}

type Statistics struct {
	TotalFiles       int
	TotalLines       int
	LinesPerFileType map[string]int
	FilesPerType     map[string]int
	AverageFileSize  float64
	LargestFile      LargestFile
}

func calculateStatistics(repoDir string, gitignore func(string) bool) Statistics {
	relevantFiles := getRelevantFiles(repoDir, gitignore)
	return Statistics{
		TotalFiles:       len(relevantFiles),
		TotalLines:       countTotalLines(relevantFiles),
		LinesPerFileType: countLinesPerFileType(relevantFiles),
		FilesPerType:     countFilesPerType(relevantFiles),
		AverageFileSize:  calculateAverageFileSize(relevantFiles),
		LargestFile:      findLargestFile(relevantFiles),
	}
}

func formatStatistics(stats Statistics) string {
	var builder strings.Builder
	builder.WriteString("Code Statistics:\n")
	builder.WriteString(fmt.Sprintf("1. Total number of files: %d\n", stats.TotalFiles))
	builder.WriteString(fmt.Sprintf("2. Total lines of code: %d\n", stats.TotalLines))
	builder.WriteString("3. Lines of code per file type:\n")
	for ext, lines := range stats.LinesPerFileType {
		if ext == "" {
			ext = "No extension"
		}
		builder.WriteString(fmt.Sprintf("   - %s: %d\n", ext, lines))
	}
	builder.WriteString("4. Number of files per file type:\n")
	for ext, count := range stats.FilesPerType {
		if ext == "" {
			ext = "No extension"
		}
		builder.WriteString(fmt.Sprintf("   - %s: %d\n", ext, count))
	}
	builder.WriteString(fmt.Sprintf("5. Average file size: %.2f bytes\n", stats.AverageFileSize))
	builder.WriteString("6. Largest file:\n")
	builder.WriteString(fmt.Sprintf("   - Name: %s\n", stats.LargestFile.Name))
	builder.WriteString(fmt.Sprintf("   - Size: %d bytes\n", stats.LargestFile.Size))
	builder.WriteString(fmt.Sprintf("   - Lines: %d\n", stats.LargestFile.Lines))
	return builder.String()
}

func processFileContent(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	content := bytes.Buffer{}

	file, err := os.Open(filePath)
	if err != nil {
		handleError(fmt.Sprintf("Error processing file %s: %v", filePath, err), false)
		return fmt.Sprintf("Error: Unable to process file content for %s", filePath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if TRUNCATE_EXTENSIONS[ext] && len(lines) > MAX_LINES_FOR_LARGE_FILES {
		for _, line := range lines[:MAX_LINES_FOR_LARGE_FILES] {
			content.WriteString(line + "\n")
		}
		content.WriteString(fmt.Sprintf("\n\n(truncated to %d lines for brevity, total file length: %d lines)", MAX_LINES_FOR_LARGE_FILES, len(lines)))
	} else {
		for _, line := range lines {
			content.WriteString(line + "\n")
		}
	}

	return content.String()
}

func main() {
	gitignorePath := filepath.Join(REPO_DIR, GITIGNORE_FILENAME)
	gitignore := parseGitignore(gitignorePath)

	if gitignore == nil {
		fmt.Printf("No %s file found. Proceeding without ignoring any files.\n", GITIGNORE_FILENAME)
	}

	folderStructure := getFolderStructure(REPO_DIR, gitignore)
	stats := calculateStatistics(REPO_DIR, gitignore)

	aiInstructions := `
AI INSTRUCTIONS:
When assisting with this project, please adhere to the following guidelines:

1. Always return the full, working file when editing code, ready for copy-paste.
2. Follow language-specific conventions and maintain consistent style.
3. Write clear, self-documenting code with concise comments for complex logic.
4. Implement proper error handling and logging.
5. Design modular, reusable code following SOLID principles.
6. Prioritize readability and maintainability over cleverness.
7. Use descriptive names for variables, functions, and classes.
8. Keep functions small and focused on a single responsibility.
9. Practice proper scoping and avoid global variables.
10. Apply appropriate design patterns to improve code structure.

Please keep these instructions in mind when providing assistance or generating code for this project.

`

	var output strings.Builder
	output.WriteString(aiInstructions + "\n")
	output.WriteString(formatStatistics(stats))
	output.WriteString("\n\nFolder Structure:\n")
	output.WriteString(strings.Join(folderStructure, "\n"))
	output.WriteString("\n\nFile Index:\n")

	relevantFiles := getRelevantFiles(REPO_DIR, gitignore)
	for i, file := range relevantFiles {
		relPath, _ := filepath.Rel(REPO_DIR, file)
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, relPath))
	}
	output.WriteString("\n\n")

	for i, filePath := range relevantFiles {
		relPath, _ := filepath.Rel(REPO_DIR, filePath)
		content := processFileContent(filePath)
		separator := strings.Repeat("=", 80) + "\n"
		fileHeader := fmt.Sprintf("FILE_%04d: %s\n", i+1, relPath)
		output.WriteString("\n" + separator + fileHeader + separator + "\n")
		output.WriteString(content)
		output.WriteString(fmt.Sprintf("\n%sEND OF FILE_%04d: %s\n%s\n", separator, i+1, relPath, separator))
	}

	err := os.WriteFile(OUTPUT_FILE, []byte(output.String()), 0644)
	if err != nil {
		handleError(fmt.Sprintf("An unexpected error occurred: %v", err), true)
	}
	fmt.Printf("All files have been concatenated into %s\n", OUTPUT_FILE)
}
