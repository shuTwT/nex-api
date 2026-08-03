package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	wantRoutes   = 75
	wantHandlers = 104
)

type manifestDocument struct {
	Version int             `json:"version"`
	Routes  []manifestRoute `json:"routes"`
}

type manifestRoute struct {
	ID         string            `json:"id"`
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

func TestContractManifest(t *testing.T) {
	contractDir := packageDirectory(t)
	manifest := readManifest(t, filepath.Join(contractDir, "manifest.json"))

	if manifest.Version != 2 {
		t.Errorf("manifest version = %d, want 2", manifest.Version)
	}
	if len(manifest.Routes) != wantRoutes {
		t.Errorf("contract routes = %d, want %d", len(manifest.Routes), wantRoutes)
	}

	seenIDs := make(map[string]struct{}, len(manifest.Routes))
	seenHandlers := make(map[string]string)
	handlerCount := 0
	for _, route := range manifest.Routes {
		if route.ID == "" {
			t.Error("contract route has an empty id")
		} else if strings.Contains(route.ID, "src/") || strings.HasSuffix(route.ID, ".ts") {
			t.Errorf("contract route id %q must not reference a source file", route.ID)
		}
		if _, exists := seenIDs[route.ID]; exists {
			t.Errorf("duplicate contract route id %q", route.ID)
		}
		seenIDs[route.ID] = struct{}{}
		if !strings.HasPrefix(route.Path, "/api/") {
			t.Errorf("contract route %q has non-API path %q", route.ID, route.Path)
		}

		for _, handler := range route.Handlers {
			validateHandler(t, route.ID, handler)
			handlerCount++
			key := handler.Method + " " + route.Path
			if previous, exists := seenHandlers[key]; exists {
				t.Errorf("duplicate contract handler %q in %s and %s", key, previous, route.ID)
			} else {
				seenHandlers[key] = route.ID
			}
		}
	}
	if handlerCount != wantHandlers {
		t.Errorf("contract handlers = %d, want %d", handlerCount, wantHandlers)
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

func validateHandler(t *testing.T, routeID string, handler manifestHandler) {
	t.Helper()
	if handler.Method == "" {
		t.Errorf("route %s has an empty method", routeID)
	}
	if !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}, handler.Method) {
		t.Errorf("route %s has unsupported method %q", routeID, handler.Method)
	}
	if !slices.Contains([]string{"none", "user", "admin"}, handler.Auth) {
		t.Errorf("route %s %s has unsupported auth %q", routeID, handler.Method, handler.Auth)
	}
	if handler.MediaType == "" {
		t.Errorf("route %s %s has empty media type", routeID, handler.Method)
	}
	if !slices.Contains([]string{"json", "raw", "stream", "file"}, handler.ResponseClass) {
		t.Errorf("route %s %s has unsupported response class %q", routeID, handler.Method, handler.ResponseClass)
	}
}
