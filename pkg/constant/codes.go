package constant

// Response codes carried in the `code` field of every JSON response body.
//
// Layout: 0 means success. Every other value is HTTP status * 100 + sequence,
// so the first three digits always match the real HTTP status (for example
// 40101 is a 401) and a client can tell the category without a lookup.
// 00 in each block is the generic code for that status.
const CodeOK = 0

const (
	// 400
	CodeBadRequest         = 40000
	CodeValidation         = 40001
	CodeInvalidFile        = 40002
	CodeFileTooLarge       = 40003
	CodeInvalidBody        = 40010
	CodeUnsupportedMCPType = 40011
	CodeMinimumRecharge    = 40020
	CodeInvalidCredits     = 40021
	CodePlanInactive       = 40022
	CodeInvalidCodeType    = 40023
	CodeCodeExpired        = 40024
	CodeCodeUnavailable    = 40025
	CodeCodeUsed           = 40026
	CodeCodesUnavailable   = 40027
	CodeBatchContainsUsed  = 40028
	CodeAlreadyInitialized = 40029
	// 401
	CodeUnauthorized       = 40100
	CodeInvalidCredentials = 40101
	CodeInvalidAPIToken    = 40102
	// 402
	CodeInsufficientCredits = 40200
	// 403
	CodeForbidden          = 40300
	CodeCSRFFailed         = 40301
	CodeMCPServiceDisabled = 40302
	// 404
	CodeNotFound           = 40400
	CodeMCPServiceNotFound = 40401
	// 405
	CodeMethodNotAllowed = 40500
	// 409
	CodeConflict            = 40900
	CodeScheduledJobExists  = 40901
	CodeScheduledJobOff     = 40902
	CodeSchedulerNotRunning = 40903
	// 413
	CodeRequestTooLarge = 41300
	// 429
	CodeTooManyRequests = 42900
	// 499
	CodeRequestCanceled = 49900
	// 500
	CodeInternal               = 50000
	CodeUploadFailed           = 50001
	CodeLoginFailed            = 50002
	CodeLogoutFailed           = 50003
	CodeCSRFUnavailable        = 50004
	CodeAuthUnavailable        = 50005
	CodeTransformFailed        = 50006
	CodeAccountingFailed       = 50007
	CodeCreditReservationFail  = 50008
	CodeMCPServiceLookupFailed = 50009
	CodeInvalidMCPService      = 50010
	CodeMCPGatewayUnavailable  = 50011
	CodeSystemUnavailable      = 50012
	CodeInitializationFailed   = 50013
	CodeInitStatusFailed       = 50014
	// 502
	CodeBadGateway          = 50200
	CodeMCPUpstreamFailed   = 50201
	CodeMCPToolsUnavailable = 50202
	// 503
	CodeServiceUnavailable = 50300
	CodeLoginUnavailable   = 50301
	// 504
	CodeGatewayTimeout = 50400
)

// slugCodes maps the string slugs used inside the code base (service errors,
// handler errors) to their numeric response code.
var slugCodes = map[string]int{
	"invalid_request":              CodeBadRequest,
	"validation_error":             CodeValidation,
	"invalid_file":                 CodeInvalidFile,
	"file_too_large":               CodeFileTooLarge,
	"invalid_body":                 CodeInvalidBody,
	"unsupported_mcp_service_type": CodeUnsupportedMCPType,
	"minimum_recharge":             CodeMinimumRecharge,
	"invalid_credits":              CodeInvalidCredits,
	"plan_inactive":                CodePlanInactive,
	"invalid_code_type":            CodeInvalidCodeType,
	"code_expired":                 CodeCodeExpired,
	"code_unavailable":             CodeCodeUnavailable,
	"code_used":                    CodeCodeUsed,
	"codes_unavailable":            CodeCodesUnavailable,
	"batch_contains_used_codes":    CodeBatchContainsUsed,
	"already_initialized":          CodeAlreadyInitialized,
	"unauthorized":                 CodeUnauthorized,
	"invalid_credentials":          CodeInvalidCredentials,
	"invalid_api_token":            CodeInvalidAPIToken,
	"insufficient_credits":         CodeInsufficientCredits,
	"forbidden":                    CodeForbidden,
	"csrf_failed":                  CodeCSRFFailed,
	"mcp_service_disabled":         CodeMCPServiceDisabled,
	"not_found":                    CodeNotFound,
	"mcp_service_not_found":        CodeMCPServiceNotFound,
	"method_not_allowed":           CodeMethodNotAllowed,
	"conflict":                     CodeConflict,
	"scheduled_job_exists":         CodeScheduledJobExists,
	"scheduled_job_disabled":       CodeScheduledJobOff,
	"scheduler_not_running":        CodeSchedulerNotRunning,
	"request_too_large":            CodeRequestTooLarge,
	"request_body_too_large":       CodeRequestTooLarge,
	"login_rate_limited":           CodeTooManyRequests,
	"request_canceled":             CodeRequestCanceled,
	"internal_error":               CodeInternal,
	"upload_failed":                CodeUploadFailed,
	"login_failed":                 CodeLoginFailed,
	"logout_failed":                CodeLogoutFailed,
	"csrf_unavailable":             CodeCSRFUnavailable,
	"auth_unavailable":             CodeAuthUnavailable,
	"transform_error":              CodeTransformFailed,
	"accounting_error":             CodeAccountingFailed,
	"credit_reservation_failed":    CodeCreditReservationFail,
	"mcp_service_lookup_failed":    CodeMCPServiceLookupFailed,
	"invalid_mcp_service":          CodeInvalidMCPService,
	"mcp_gateway_unavailable":      CodeMCPGatewayUnavailable,
	"system_unavailable":           CodeSystemUnavailable,
	"initialization_failed":        CodeInitializationFailed,
	"initialization_status_failed": CodeInitStatusFailed,
	"mcp_upstream_failed":          CodeMCPUpstreamFailed,
	"mcp_tools_unavailable":        CodeMCPToolsUnavailable,
	"login_unavailable":            CodeLoginUnavailable,
	"request_timeout":              CodeGatewayTimeout,
	"upstream_error":               CodeBadGateway,
	"upstream_timeout":             CodeGatewayTimeout,
}

// CodeFor returns the numeric response code for a slug. Unknown or empty slugs
// fall back to the generic code of the HTTP status (status * 100), so a new
// error never produces a code that disagrees with its status.
func CodeFor(slug string, status int) int {
	if code, ok := slugCodes[slug]; ok {
		return code
	}
	if status < 100 || status > 599 {
		return CodeInternal
	}
	return status * 100
}
