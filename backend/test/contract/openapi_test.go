package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type openAPIDocument struct {
	Paths map[string]openAPIPathItem `yaml:"paths"`
}

type openAPIPathItem struct {
	Get     *openAPIOperation `yaml:"get"`
	Post    *openAPIOperation `yaml:"post"`
	Put     *openAPIOperation `yaml:"put"`
	Patch   *openAPIOperation `yaml:"patch"`
	Delete  *openAPIOperation `yaml:"delete"`
	Options *openAPIOperation `yaml:"options"`
	Head    *openAPIOperation `yaml:"head"`
}

type openAPIOperation struct {
	OperationID string `yaml:"operationId"`
}

func TestManifestRoutesMapToOpenAPI(t *testing.T) {
	// Given
	contractDir := packageDirectory(t)
	manifest := readManifest(t, filepath.Join(contractDir, "manifest.json"))
	specPath := filepath.Join(contractDir, "..", "..", "openapi", "openapi.yaml")
	specData, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read OpenAPI spec: %v", err)
	}
	var document openAPIDocument
	if err := yaml.Unmarshal(specData, &document); err != nil {
		t.Fatalf("decode OpenAPI spec: %v", err)
	}

	// When
	mappedRoutes := 0
	for _, route := range manifest.Routes {
		path := normalizeManifestPath(route.Path)
		pathItem, exists := document.Paths[path]
		if !exists {
			t.Errorf("manifest route %s is missing from OpenAPI paths", route.File)
			continue
		}
		routeMapped := true
		for _, handler := range route.Handlers {
			operation := pathItem.operation(handler.Method)
			if operation == nil || operation.OperationID == "" {
				routeMapped = false
				t.Errorf("manifest handler %s %s has no operationId", handler.Method, route.File)
			}
		}
		if routeMapped {
			mappedRoutes++
		}
	}

	// Then
	if mappedRoutes != len(manifest.Routes) {
		t.Fatalf("OpenAPI manifest mapping: %d/%d routes", mappedRoutes, len(manifest.Routes))
	}
	t.Logf("OpenAPI manifest mapping: %d/%d routes", mappedRoutes, len(manifest.Routes))
}

func (item openAPIPathItem) operation(method string) *openAPIOperation {
	switch strings.ToUpper(method) {
	case "GET":
		return item.Get
	case "POST":
		return item.Post
	case "PUT":
		return item.Put
	case "PATCH":
		return item.Patch
	case "DELETE":
		return item.Delete
	case "OPTIONS":
		return item.Options
	case "HEAD":
		return item.Head
	default:
		return nil
	}
}

func normalizeManifestPath(path string) string {
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[index] = "{" + strings.TrimSuffix(strings.TrimPrefix(segment, ":"), "*") + "}"
		}
	}
	return "/" + strings.Join(segments, "/")
}
