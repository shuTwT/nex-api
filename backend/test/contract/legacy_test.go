package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

const (
	wantRouteFiles = 76
	wantHandlers   = 106
)

var (
	functionHandlerPattern = regexp.MustCompile(`(?m)export\s+(?:async\s+)?function\s+(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)`)
	aliasExportPattern     = regexp.MustCompile(`(?s)export\s*\{([^}]*)\}`)
	aliasMethodPattern     = regexp.MustCompile(`\bas\s+(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)`)
)

type manifestDocument struct {
	Version        int             `json:"version"`
	SourceRoot     string          `json:"sourceRoot"`
	RouteFileCount int             `json:"routeFileCount"`
	Routes         []manifestRoute `json:"routes"`
}

type manifestRoute struct {
	File       string            `json:"file"`
	Path       string            `json:"path"`
	AuthScheme string            `json:"authScheme,omitempty"`
	Handlers   []manifestHandler `json:"handlers"`
}

type manifestHandler struct {
	Method           string `json:"method"`
	Auth             string `json:"auth"`
	MediaType        string `json:"mediaType"`
	RequestMediaType string `json:"requestMediaType,omitempty"`
	ResponseClass    string `json:"responseClass"`
}

func TestLegacyManifest(t *testing.T) {
	// Given
	contractDir := packageDirectory(t)
	manifest := readManifest(t, filepath.Join(contractDir, "manifest.json"))
	repoRoot := filepath.Clean(filepath.Join(contractDir, "..", "..", ".."))
	sourceRoot := filepath.Join(repoRoot, filepath.FromSlash(manifest.SourceRoot))
	sourceFiles := discoverRouteFiles(t, repoRoot, sourceRoot)

	// When
	manifestFiles := make(map[string]manifestRoute, len(manifest.Routes))
	manifestHandlerCount := 0
	seenHandlers := make(map[string]string)
	for _, route := range manifest.Routes {
		if _, exists := manifestFiles[route.File]; exists {
			t.Errorf("duplicate manifest route file %q", route.File)
		}
		manifestFiles[route.File] = route
		manifestHandlerCount += len(route.Handlers)

		sourcePath := filepath.Join(repoRoot, filepath.FromSlash(route.File))
		if _, err := os.Stat(sourcePath); err != nil {
			t.Errorf("manifest route %q is not backed by a source file: %v", route.File, err)
			continue
		}
		if got := routePath(route.File); got != route.Path {
			t.Errorf("route %q has path %q, want %q", route.File, route.Path, got)
		}

		sourceMethods := readExportedMethods(t, sourcePath)
		manifestMethods := make(map[string]int, len(route.Handlers))
		for _, handler := range route.Handlers {
			validateHandler(t, route.File, handler)
			manifestMethods[handler.Method]++
			key := handler.Method + " " + route.Path
			if previous, exists := seenHandlers[key]; exists {
				t.Errorf("duplicate handler %q in %s and %s", key, previous, route.File)
			} else {
				seenHandlers[key] = route.File
			}
		}

		compareMethodCounts(t, route.File, sourceMethods, manifestMethods)
	}

	// Then
	if manifest.Version != 1 {
		t.Errorf("manifest version = %d, want 1", manifest.Version)
	}
	if manifest.SourceRoot != "src/app/api" {
		t.Errorf("manifest sourceRoot = %q, want %q", manifest.SourceRoot, "src/app/api")
	}
	if manifest.RouteFileCount != wantRouteFiles {
		t.Errorf("manifest routeFileCount = %d, want %d", manifest.RouteFileCount, wantRouteFiles)
	}
	if len(sourceFiles) != wantRouteFiles {
		t.Errorf("source route files = %d, want %d", len(sourceFiles), wantRouteFiles)
	}
	if len(manifest.Routes) != wantRouteFiles {
		t.Errorf("manifest route files = %d, want %d", len(manifest.Routes), wantRouteFiles)
	}
	if manifestHandlerCount != wantHandlers {
		t.Errorf("manifest exported handlers = %d, want %d", manifestHandlerCount, wantHandlers)
	}

	for _, sourceFile := range sourceFiles {
		if _, exists := manifestFiles[sourceFile]; !exists {
			t.Errorf("source route file is unmapped: %s", sourceFile)
		}
	}
	for manifestFile := range manifestFiles {
		if !slices.Contains(sourceFiles, manifestFile) {
			t.Errorf("manifest route file is not in source tree: %s", manifestFile)
		}
	}
}

func packageDirectory(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get package directory: %v", err)
	}
	return workingDirectory
}

func readManifest(t *testing.T, path string) manifestDocument {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest %s: %v", path, err)
	}

	var manifest manifestDocument
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode manifest %s: %v", path, err)
	}
	return manifest
}

func discoverRouteFiles(t *testing.T, repoRoot, sourceRoot string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(sourceRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "route.ts" {
			return nil
		}

		relativePath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return fmt.Errorf("make route path relative: %w", err)
		}
		files = append(files, filepath.ToSlash(relativePath))
		return nil
	})
	if err != nil {
		t.Fatalf("walk route tree: %v", err)
	}
	sort.Strings(files)
	return files
}

func readExportedMethods(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read route %s: %v", path, err)
	}
	text := string(data)

	methods := make([]string, 0)
	for _, match := range functionHandlerPattern.FindAllStringSubmatch(text, -1) {
		methods = append(methods, match[1])
	}
	for _, exportMatch := range aliasExportPattern.FindAllStringSubmatch(text, -1) {
		for _, methodMatch := range aliasMethodPattern.FindAllStringSubmatch(exportMatch[1], -1) {
			methods = append(methods, methodMatch[1])
		}
	}
	return methods
}

func compareMethodCounts(t *testing.T, routeFile string, sourceMethods []string, manifestMethods map[string]int) {
	t.Helper()
	sourceCounts := make(map[string]int, len(sourceMethods))
	for _, method := range sourceMethods {
		sourceCounts[method]++
	}
	for method, count := range sourceCounts {
		if manifestMethods[method] != count {
			t.Errorf("route %s method %s mapped %d times, want %d", routeFile, method, manifestMethods[method], count)
		}
	}
	for method, count := range manifestMethods {
		if sourceCounts[method] != count {
			t.Errorf("route %s manifest method %s appears %d times, source has %d", routeFile, method, count, sourceCounts[method])
		}
	}
}

func validateHandler(t *testing.T, routeFile string, handler manifestHandler) {
	t.Helper()
	if handler.Method == "" {
		t.Errorf("route %s has an empty method", routeFile)
	}
	if !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}, handler.Method) {
		t.Errorf("route %s has unsupported method %q", routeFile, handler.Method)
	}
	if !slices.Contains([]string{"none", "user", "admin"}, handler.Auth) {
		t.Errorf("route %s %s has unsupported auth %q", routeFile, handler.Method, handler.Auth)
	}
	if handler.MediaType == "" {
		t.Errorf("route %s %s has empty media type", routeFile, handler.Method)
	}
	if !slices.Contains([]string{"json", "raw", "stream", "file"}, handler.ResponseClass) {
		t.Errorf("route %s %s has unsupported response class %q", routeFile, handler.Method, handler.ResponseClass)
	}
}

func routePath(sourceFile string) string {
	const prefix = "src/app/api/"
	relativePath := strings.TrimPrefix(strings.TrimSuffix(sourceFile, "/route.ts"), prefix)
	segments := strings.Split(relativePath, "/")
	for index, segment := range segments {
		switch {
		case strings.HasPrefix(segment, "[...") && strings.HasSuffix(segment, "]"):
			segments[index] = ":" + strings.TrimSuffix(strings.TrimPrefix(segment, "[..."), "]") + "*"
		case strings.HasPrefix(segment, "[[...") && strings.HasSuffix(segment, "]]"):
			segments[index] = ":" + strings.TrimSuffix(strings.TrimPrefix(segment, "[[..."), "]]") + "*"
		case strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]"):
			segments[index] = ":" + strings.TrimSuffix(strings.TrimPrefix(segment, "["), "]")
		}
	}
	return "/api/" + strings.Join(segments, "/")
}
