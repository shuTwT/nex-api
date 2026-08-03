package mcpgateway

import (
	"strings"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/internal/infra/proxy"
)

// ValidateService checks an MCP service configuration before invocation.
// stdioPath is the script-worker executable path used to validate the stdio
// environment.
func ValidateService(service *ent.McpService, stdioPath string) (map[string]string, error) {
	switch service.Type {
	case "stdio":
		if strings.TrimSpace(service.Command) == "" {
			return nil, ErrInvalidService
		}
		envVars, err := proxy.ParseServiceEnvVars(service.EnvVars)
		if err != nil {
			return nil, err
		}
		if _, err := proxy.WorkerEnvironment(stdioPath, envVars); err != nil {
			return nil, err
		}
		return envVars, nil
	case "sse", "streamableHttp":
		if err := proxy.ValidateHTTPURL(service.Endpoint); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, ErrUnsupportedType
	}
}
