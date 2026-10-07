// Command genoperations emits frontend/src/api/generated/operations.ts from
// openapi/openapi.yaml so the frontend facade's operationId -> [method, path]
// table is generated, never hand-maintained.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

type openAPIDocument struct {
	Paths map[string]map[string]struct {
		OperationID string `yaml:"operationId"`
	} `yaml:"paths"`
}

func main() {
	in := flag.String("in", "openapi/openapi.yaml", "input OpenAPI YAML document")
	out := flag.String("out", "frontend/src/api/generated/operations.ts", "output TypeScript module")
	flag.Parse()

	source, err := os.ReadFile(*in)
	if err != nil {
		fail("read %s: %v", *in, err)
	}
	var document openAPIDocument
	if err := yaml.Unmarshal(source, &document); err != nil {
		fail("parse %s: %v", *in, err)
	}

	type entry struct {
		operationID string
		method      string
		path        string
	}
	entries := make([]entry, 0, 128)
	seen := make(map[string]string)
	for path, methods := range document.Paths {
		for method, operation := range methods {
			switch method {
			case "get", "post", "put", "patch", "delete", "options":
			default:
				continue
			}
			if operation.OperationID == "" {
				fail("path %s method %s has no operationId", path, method)
			}
			if previous, duplicate := seen[operation.OperationID]; duplicate {
				fail("duplicate operationId %s (%s and %s)", operation.OperationID, previous, path)
			}
			seen[operation.OperationID] = path
			entries = append(entries, entry{operation.OperationID, method, path})
		}
	}
	if len(entries) == 0 {
		fail("no operations found in %s", *in)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].operationID < entries[j].operationID })

	var builder strings.Builder
	builder.WriteString("// Code generated from openapi/openapi.yaml by tools/genoperations; DO NOT EDIT.\n")
	builder.WriteString("// Run `make generate` to regenerate. The keys must stay aligned with the\n")
	builder.WriteString("// `operations` interface in ./schema.ts; the `satisfies` clause below fails\n")
	builder.WriteString("// to compile as soon as the two drift apart.\n")
	builder.WriteString("import type { operations as schemaOperations } from \"./schema\";\n\n")
	builder.WriteString("export type OperationHttpMethod = \"get\" | \"post\" | \"put\" | \"patch\" | \"delete\" | \"options\";\n\n")
	builder.WriteString("export const operations = {\n")
	for _, item := range entries {
		fmt.Fprintf(&builder, "  %s: [%q, %q],\n", item.operationID, item.method, item.path)
	}
	builder.WriteString("} as const satisfies Record<keyof schemaOperations, readonly [OperationHttpMethod, string]>;\n\n")
	builder.WriteString("export type OperationName = keyof typeof operations;\n")

	if err := os.WriteFile(*out, []byte(builder.String()), 0o644); err != nil {
		fail("write %s: %v", *out, err)
	}
	fmt.Printf("wrote %s (%d operations)\n", *out, len(entries))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "genoperations: "+format+"\n", args...)
	os.Exit(1)
}
