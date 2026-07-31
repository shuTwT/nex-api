//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type manifest struct {
	Routes []route `json:"routes"`
}

type route struct {
	File       string    `json:"file"`
	Path       string    `json:"path"`
	AuthScheme string    `json:"authScheme,omitempty"`
	Handlers   []handler `json:"handlers"`
}

type handler struct {
	Method           string `json:"method"`
	Auth             string `json:"auth"`
	MediaType        string `json:"mediaType"`
	RequestMediaType string `json:"requestMediaType,omitempty"`
	ResponseClass    string `json:"responseClass"`
}

func main() {
	manifestData, err := os.ReadFile(filepath.Join("..", "..", "test", "contract", "manifest.json"))
	if err != nil {
		fatal(err)
	}

	var inventory manifest
	if err := json.Unmarshal(manifestData, &inventory); err != nil {
		fatal(err)
	}

	var document strings.Builder
	document.WriteString("openapi: 3.1.0\n")
	document.WriteString("info:\n  title: Nex API\n  version: 1.0.0\n  license: { name: MIT, identifier: MIT }\n  description: |\n    Contract for the typed JSON API and its explicitly recorded raw transports.\n")
	document.WriteString("servers:\n  - url: https://api.nex-api.invalid\n    description: Contract placeholder\n")
	document.WriteString("tags:\n")
	document.WriteString("  - { name: auth, description: Authentication and session operations }\n")
	document.WriteString("  - { name: admin, description: Administrator operations }\n")
	document.WriteString("  - { name: user, description: Authenticated user operations }\n")
	document.WriteString("  - { name: public, description: Public operations }\n")
	document.WriteString("  - { name: payment, description: Payment and membership operations }\n")
	document.WriteString("  - { name: gateway, description: Raw API gateway operations }\n")
	document.WriteString("  - { name: mcp, description: MCP stream operations }\n")
	document.WriteString("  - { name: files, description: File transfer operations }\n")
	document.WriteString("paths:\n")
	for _, currentRoute := range inventory.Routes {
		writePath(&document, currentRoute)
	}
	writeComponents(&document)

	if err := os.MkdirAll(filepath.Join("..", "..", "openapi"), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join("..", "..", "openapi", "openapi.yaml"), []byte(document.String()), 0o644); err != nil {
		fatal(err)
	}
}

func writePath(document *strings.Builder, currentRoute route) {
	path := openAPIPath(currentRoute.Path)
	fmt.Fprintf(document, "  %s:\n", path)
	parameters := pathParameters(currentRoute.Path)
	if len(parameters) > 0 {
		document.WriteString("    parameters:\n")
		for _, parameter := range parameters {
			fmt.Fprintf(document, "      - { name: %s, in: path, required: true, schema: { type: string } }\n", parameter)
		}
	}
	for _, currentHandler := range currentRoute.Handlers {
		writeOperation(document, currentRoute, currentHandler)
	}
}

func writeOperation(document *strings.Builder, currentRoute route, currentHandler handler) {
	method := strings.ToLower(currentHandler.Method)
	operationID := operationID(currentRoute.File, currentHandler.Method)
	tag := routeTag(currentRoute.Path, currentRoute.AuthScheme, currentHandler.Auth, currentHandler.ResponseClass)
	fmt.Fprintf(document, "    %s:\n", method)
	fmt.Fprintf(document, "      operationId: %s\n", operationID)
	fmt.Fprintf(document, "      summary: %s\n", strconv.Quote(operationID))
	fmt.Fprintf(document, "      tags: [%s]\n", tag)
	fmt.Fprintf(document, "      x-source-file: %s\n", strconv.Quote(currentRoute.File))
	fmt.Fprintf(document, "      x-auth-role: %s\n", strconv.Quote(currentHandler.Auth))
	fmt.Fprintf(document, "      x-response-class: %s\n", strconv.Quote(currentHandler.ResponseClass))
	if currentRoute.AuthScheme != "" {
		fmt.Fprintf(document, "      x-auth-scheme: %s\n", strconv.Quote(currentRoute.AuthScheme))
	}
	if currentHandler.ResponseClass != "json" || strings.Contains(currentRoute.Path, "/payment/callback/") {
		document.WriteString("      x-client-exclude: true\n")
	}
	writeSecurity(document, currentRoute, currentHandler)
	writeRequestBody(document, currentRoute, currentHandler)
	writeQueryParameters(document, currentRoute, currentHandler)
	writeResponses(document, currentRoute, currentHandler)
	if strings.Contains(currentRoute.Path, "/payment/business/") {
		callback := "RechargeWebhook"
		if strings.Contains(currentRoute.Path, "/subscription") {
			callback = "SubscriptionWebhook"
		}
		fmt.Fprintf(document, "      callbacks:\n        paymentStatus: { $ref: '#/components/callbacks/%s' }\n", callback)
	}
}

func writeSecurity(document *strings.Builder, currentRoute route, currentHandler handler) {
	switch {
	case currentRoute.AuthScheme == "api-token":
		document.WriteString("      security: [{ apiToken: [] }]\n")
	case currentRoute.AuthScheme == "cron-secret":
		document.WriteString("      security: [{ cronSecret: [] }]\n")
	case currentHandler.Auth == "none":
		document.WriteString("      security: []\n")
	default:
		document.WriteString("      security: [{ sessionCookie: [] }]\n")
	}
}

func writeRequestBody(document *strings.Builder, currentRoute route, currentHandler handler) {
	method := currentHandler.Method
	if method == "GET" || method == "DELETE" || method == "OPTIONS" || method == "HEAD" {
		return
	}
	mediaType := currentHandler.RequestMediaType
	if mediaType == "" {
		mediaType = currentHandler.MediaType
	}
	schema := "JsonRequest"
	if currentRoute.Path == "/api/upload" {
		schema = "UploadRequest"
	} else if currentRoute.Path == "/api/payment/callback/alipay" {
		schema = "AlipayWebhook"
	} else if currentRoute.Path == "/api/payment/callback/wechat" {
		schema = "WechatWebhook"
	} else if currentRoute.Path == "/api/payment/callback/mock" {
		schema = "MockPaymentWebhook"
	} else if currentHandler.ResponseClass == "raw" {
		schema = "RawPayload"
	}
	if mediaType == "application/json; charset=utf-8" {
		mediaType = "application/json"
	}
	if strings.Contains(mediaType, "|") {
		mediaType = "application/json"
	}
	fmt.Fprintf(document, "      requestBody:\n        required: true\n        content:\n          %s:\n            schema: { $ref: '#/components/schemas/%s' }\n", strconv.Quote(mediaType), schema)
}

func writeQueryParameters(document *strings.Builder, currentRoute route, currentHandler handler) {
	if currentHandler.Method != "GET" {
		return
	}
	parameters := make([]string, 0, 3)
	if strings.Contains(currentRoute.Path, "/export") || !strings.Contains(currentRoute.Path, "{") {
		parameters = append(parameters, "Page", "Limit", "Search")
	}
	if len(parameters) == 0 {
		return
	}
	document.WriteString("      parameters:\n")
	for _, parameter := range parameters {
		if parameter == "Page" || parameter == "Limit" || parameter == "Search" {
			fmt.Fprintf(document, "        - { $ref: '#/components/parameters/%s' }\n", parameter)
			continue
		}
		fmt.Fprintf(document, "        - { $ref: '#/components/parameters/%s' }\n", title(parameter))
	}
}

func writeResponses(document *strings.Builder, currentRoute route, currentHandler handler) {
	responseRef := "JsonSuccess"
	if currentHandler.ResponseClass == "raw" {
		responseRef = "RawSuccess"
	} else if currentHandler.ResponseClass == "stream" {
		responseRef = "StreamSuccess"
	} else if currentHandler.ResponseClass == "file" {
		responseRef = "FileSuccess"
	} else if currentRoute.Path == "/api/upload" {
		responseRef = "UploadSuccess"
	} else if currentRoute.Path == "/api/auth/me" {
		responseRef = "AuthSuccess"
	} else if strings.Contains(currentRoute.Path, "/export") {
		responseRef = "CsvOrJsonSuccess"
	} else if currentHandler.Method == "GET" && !strings.Contains(currentRoute.Path, "{") {
		responseRef = "PaginatedSuccess"
	}
	fmt.Fprintf(document, "      responses:\n        '200': { $ref: '#/components/responses/%s' }\n", responseRef)
	if currentHandler.Method == "POST" && currentHandler.ResponseClass == "json" {
		document.WriteString("        '201': { $ref: '#/components/responses/JsonSuccess' }\n")
	}
	if currentHandler.Method == "OPTIONS" {
		document.WriteString("        '204': { $ref: '#/components/responses/NoContent' }\n")
	}
	document.WriteString("        '400': { $ref: '#/components/responses/BadRequest' }\n")
	document.WriteString("        '401': { $ref: '#/components/responses/Unauthorized' }\n")
	document.WriteString("        '403': { $ref: '#/components/responses/Forbidden' }\n")
	document.WriteString("        '404': { $ref: '#/components/responses/NotFound' }\n")
	document.WriteString("        '500': { $ref: '#/components/responses/InternalError' }\n")
}

func writeComponents(document *strings.Builder) {
	document.WriteString(`components:
  securitySchemes:
    sessionCookie:
      type: apiKey
      in: cookie
      name: next-auth.session-token
    apiToken:
      type: http
      scheme: bearer
      bearerFormat: API token
    cronSecret:
      type: apiKey
      in: header
      name: x-cron-secret
  parameters:
    Page: { name: page, in: query, required: false, schema: { type: integer, minimum: 1, default: 1 } }
    Limit: { name: limit, in: query, required: false, schema: { type: integer, minimum: 1, maximum: 100, default: 10 } }
    Search: { name: search, in: query, required: false, schema: { type: string } }
  schemas:
    JsonValue:
      oneOf:
        - { type: string }
        - { type: number }
        - { type: boolean }
        - type: array
          items: { $ref: '#/components/schemas/JsonValue' }
        - { $ref: '#/components/schemas/JsonObject' }
    JsonObject:
      type: object
      additionalProperties: { $ref: '#/components/schemas/JsonValue' }
    JsonRequest:
      $ref: '#/components/schemas/JsonObject'
    Pagination:
      type: object
      required: [page, limit, total, totalPages]
      properties:
        page: { type: integer, minimum: 1 }
        limit: { type: integer, minimum: 1 }
        total: { type: integer, minimum: 0 }
        totalPages: { type: integer, minimum: 0 }
    JsonResponse:
      type: object
      required: [success]
      properties:
        success: { type: boolean, const: true }
        data: { $ref: '#/components/schemas/JsonValue' }
        error: { type: string }
        pagination: { $ref: '#/components/schemas/Pagination' }
    PaginatedResponse:
      type: object
      required: [success, data, pagination]
      properties:
        success: { type: boolean, const: true }
        data: { type: array, items: { $ref: '#/components/schemas/JsonValue' } }
        error: { type: string }
        pagination: { $ref: '#/components/schemas/Pagination' }
    ErrorResponse:
      type: object
      required: [success, error]
      properties:
        success: { type: boolean, const: false }
        error: { type: string }
    AuthUser:
      type: object
      required: [id, role]
      properties:
        id: { type: string }
        email: { type: string, format: email }
        username: { type: string }
        role: { type: string, enum: [user, admin] }
        credits: { type: number }
    AuthResponse:
      allOf:
        - $ref: '#/components/schemas/JsonResponse'
        - type: object
          properties:
            data: { $ref: '#/components/schemas/AuthUser' }
    UploadRequest:
      type: object
      required: [file]
      properties:
        file: { type: string, format: binary }
      additionalProperties: false
    UploadMetadata:
      type: object
      required: [url, filename, size, type]
      properties:
        url: { type: string, format: uri-reference }
        filename: { type: string }
        size: { type: integer, minimum: 0 }
        type: { type: string }
    UploadResponse:
      allOf:
        - $ref: '#/components/schemas/JsonResponse'
        - type: object
          properties:
            data: { $ref: '#/components/schemas/UploadMetadata' }
    RawPayload:
      type: string
      format: binary
    CsvFile:
      type: string
      format: binary
      contentMediaType: text/csv
    JsonStream:
      type: object
      description: JSON records are emitted as a UTF-8 stream; clients must consume incrementally.
      additionalProperties: true
    AlipayWebhook:
      type: object
      additionalProperties: { type: string }
    WechatWebhook:
      type: object
      required: [outTradeNo, status]
      properties:
        outTradeNo: { type: string }
        transactionId: { type: string }
        status: { type: string }
        paidAt: { type: string, format: date-time }
      additionalProperties: true
    MockPaymentWebhook:
      type: object
      required: [outTradeNo]
      properties:
        outTradeNo: { type: string }
        success: { type: boolean, default: true }
      additionalProperties: false
    PaymentWebhook:
      type: object
      required: [event, outTradeNo, status]
      properties:
        event: { type: string }
        outTradeNo: { type: string }
        status: { type: string }
        transactionId: { type: string }
        paidAt: { type: string, format: date-time }
        callbackUrl: { type: string, format: uri-reference }
      additionalProperties: true
  callbacks:
    RechargeWebhook:
      '{$request.body#/callbackUrl}':
        post:
          operationId: payment_recharge_webhook
          summary: Receive recharge payment status
          security: []
          requestBody:
            required: true
            content:
              application/json: { schema: { $ref: '#/components/schemas/PaymentWebhook' } }
          responses:
            '200': { $ref: '#/components/responses/JsonSuccess' }
            '400': { $ref: '#/components/responses/BadRequest' }
    SubscriptionWebhook:
      '{$request.body#/callbackUrl}':
        post:
          operationId: payment_subscription_webhook
          summary: Receive subscription payment status
          security: []
          requestBody:
            required: true
            content:
              application/json: { schema: { $ref: '#/components/schemas/PaymentWebhook' } }
          responses:
            '200': { $ref: '#/components/responses/JsonSuccess' }
            '400': { $ref: '#/components/responses/BadRequest' }
  responses:
    JsonSuccess:
      description: Successful JSON envelope.
      content:
        application/json: { schema: { $ref: '#/components/schemas/JsonResponse' } }
    PaginatedSuccess:
      description: Successful paginated JSON envelope.
      content:
        application/json: { schema: { $ref: '#/components/schemas/PaginatedResponse' } }
    AuthSuccess:
      description: Authenticated user response.
      content:
        application/json: { schema: { $ref: '#/components/schemas/AuthResponse' } }
    UploadSuccess:
      description: Uploaded file metadata.
      content:
        application/json: { schema: { $ref: '#/components/schemas/UploadResponse' } }
    CsvOrJsonSuccess:
      description: Export response. Legacy handlers wrap CSV text in the JSON data field.
      content:
        application/json: { schema: { $ref: '#/components/schemas/JsonResponse' } }
        text/csv: { schema: { $ref: '#/components/schemas/CsvFile' } }
    RawSuccess:
      description: Raw upstream or webhook response; not part of the typed client.
      content:
        '*/*': { schema: { $ref: '#/components/schemas/RawPayload' } }
        application/json: { schema: { $ref: '#/components/schemas/RawPayload' } }
        text/plain: { schema: { type: string } }
    StreamSuccess:
      description: Incremental JSON stream; not part of the typed client.
      content:
        application/json: { schema: { $ref: '#/components/schemas/JsonStream' } }
        text/event-stream: { schema: { type: string } }
    FileSuccess:
      description: Binary file download; not part of the typed client.
      content:
        '*/*': { schema: { $ref: '#/components/schemas/RawPayload' } }
    NoContent:
      description: No content.
    BadRequest:
      description: Invalid request.
      content:
        application/json: { schema: { $ref: '#/components/schemas/ErrorResponse' } }
    Unauthorized:
      description: Authentication required.
      content:
        application/json: { schema: { $ref: '#/components/schemas/ErrorResponse' } }
    Forbidden:
      description: Access denied.
      content:
        application/json: { schema: { $ref: '#/components/schemas/ErrorResponse' } }
    NotFound:
      description: Resource not found.
      content:
        application/json: { schema: { $ref: '#/components/schemas/ErrorResponse' } }
    InternalError:
      description: Internal server error.
      content:
        application/json: { schema: { $ref: '#/components/schemas/ErrorResponse' } }
`)
}

func openAPIPath(path string) string {
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[index] = "{" + strings.TrimSuffix(strings.TrimPrefix(segment, ":"), "*") + "}"
		}
	}
	return "/" + strings.Join(segments, "/")
}

func pathParameters(path string) []string {
	converted := openAPIPath(path)
	parts := strings.Split(converted, "/")
	parameters := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			parameters = append(parameters, strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}"))
		}
	}
	return parameters
}

func operationID(file, method string) string {
	name := strings.TrimPrefix(file, "src/app/api/")
	name = strings.TrimSuffix(name, "/route.ts")
	parts := strings.Split(name, "/")
	for index, part := range parts {
		part = strings.TrimPrefix(strings.TrimSuffix(part, "]"), "[")
		part = strings.TrimPrefix(part, "...")
		parts[index] = strings.NewReplacer("-", "_", ".", "_").Replace(part)
	}
	return strings.ToLower(strings.Join(append(parts, "route", strings.ToLower(method)), "_"))
}

func routeTag(path, authScheme, authRole, responseClass string) string {
	if authScheme == "api-token" {
		if strings.Contains(path, "/mcp/") {
			return "mcp"
		}
		return "gateway"
	}
	if responseClass == "file" {
		return "files"
	}
	if strings.Contains(path, "/payment/") {
		return "payment"
	}
	if authRole == "admin" {
		return "admin"
	}
	if authRole == "user" {
		return "user"
	}
	first := strings.TrimPrefix(path, "/api/")
	first = strings.Split(first, "/")[0]
	if first == "auth" {
		return "auth"
	}
	if authScheme == "cron-secret" || strings.HasPrefix(first, "admin") {
		return "admin"
	}
	if first == "marketplace" || authScheme == "" {
		return "public"
	}
	return authScheme
}

func title(value string) string {
	if value == "outTradeNo" {
		return "OutTradeNo"
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
