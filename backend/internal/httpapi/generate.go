package httpapi

//go:generate go run ./specgen.go
//go:generate go run ./clientgen.go
//go:generate mkdir -p generated
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.2 --config=types.cfg.yaml ../../openapi/openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.2 --config=server.cfg.yaml ../../openapi/openapi.yaml
