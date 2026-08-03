package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dop251/goja"
)

type EngineOptions struct {
	MaxOutputBytes int
}

type Engine struct {
	maxOutputBytes int
}

func NewEngine(options EngineOptions) *Engine {
	return &Engine{maxOutputBytes: options.MaxOutputBytes}
}

func (e *Engine) Execute(ctx context.Context, job Job) (TransformResult, error) {
	if err := job.validate(); err != nil {
		return TransformResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return TransformResult{}, err
	}
	if stringsAreBlank(job.Script) {
		return TransformResult{}, nil
	}

	vm := goja.New()
	if err := installContext(vm, job); err != nil {
		return TransformResult{}, fmt.Errorf("install script context: %w", err)
	}
	if _, err := vm.RunString(`var console = { log: function() {}, error: function() {} };`); err != nil {
		return TransformResult{}, fmt.Errorf("install console: %w", err)
	}

	interruptDone := make(chan struct{})
	interruptStopped := make(chan struct{})
	go func() {
		defer close(interruptStopped)
		select {
		case <-ctx.Done():
			vm.Interrupt(interruptSignal{code: interruptionCode(ctx.Err()), message: ctx.Err().Error()})
		case <-interruptDone:
		}
	}()

	value, runErr := vm.RunString(wrappedScript(job))
	close(interruptDone)
	<-interruptStopped
	if runErr != nil {
		return TransformResult{}, scriptError(runErr)
	}
	result, err := exportResult(vm, value, job.Kind, e.maxOutputBytes)
	if err != nil {
		return TransformResult{}, err
	}
	if err := contextError(ctx); err != nil {
		return TransformResult{}, err
	}
	if e.maxOutputBytes > 0 {
		encoded, err := json.Marshal(result)
		if err != nil {
			return TransformResult{}, fmt.Errorf("encode transform result: %w", err)
		}
		if len(encoded) > e.maxOutputBytes {
			return TransformResult{}, &WorkerError{Code: ErrorCodeOutputLimit, Message: "transform output exceeds configured limit"}
		}
	}
	return result, nil
}

type interruptSignal struct {
	code    ErrorCode
	message string
}

func interruptionCode(err error) ErrorCode {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorCodeTimeout
	}
	return ErrorCodeCanceled
}

func contextError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return &WorkerError{Code: interruptionCode(err), Message: err.Error(), Retryable: true}
	}
	return nil
}

func installContext(vm *goja.Runtime, job Job) error {
	if err := setJSONValue(vm, "body", job.Body); err != nil {
		return fmt.Errorf("body: %w", err)
	}
	if err := setJSONValue(vm, "responseBody", job.ResponseBody); err != nil {
		return fmt.Errorf("response body: %w", err)
	}
	vm.Set("headers", cloneStrings(job.Headers))
	vm.Set("query", cloneStrings(job.Query))
	vm.Set("responseHeaders", cloneStrings(job.ResponseHeaders))
	return nil
}

func setJSONValue(vm *goja.Runtime, name string, raw json.RawMessage) error {
	if len(raw) == 0 {
		vm.Set(name, nil)
		return nil
	}
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	vm.Set(name, value)
	return nil
}

func wrappedScript(job Job) string {
	var body string
	switch job.Kind {
	case ScriptKindPre:
		body = "headers: typeof headers !== 'undefined' ? headers : undefined,\n" +
			"query: typeof query !== 'undefined' ? query : undefined,\n" +
			"body: typeof body !== 'undefined' ? body : undefined,"
	case ScriptKindPost:
		body = "responseBody: typeof responseBody !== 'undefined' ? responseBody : undefined,\n" +
			"responseHeaders: typeof responseHeaders !== 'undefined' ? responseHeaders : undefined,"
	}
	return "(function() {\n" + job.Script + "\nreturn {\n" + body + "\n};\n})()"
}

func exportResult(vm *goja.Runtime, value goja.Value, kind ScriptKind, maxBytes int) (TransformResult, error) {
	object := value.ToObject(vm)
	result := TransformResult{}
	switch kind {
	case ScriptKindPre:
		var err error
		result.Headers, err = exportStringMap(vm, object.Get("headers"))
		if err != nil {
			return TransformResult{}, fmt.Errorf("headers: %w", err)
		}
		result.Query, err = exportStringMap(vm, object.Get("query"))
		if err != nil {
			return TransformResult{}, fmt.Errorf("query: %w", err)
		}
		result.Body, result.BodySet, err = exportJSON(vm, object.Get("body"), maxBytes)
		if err != nil {
			return TransformResult{}, fmt.Errorf("body: %w", err)
		}
	case ScriptKindPost:
		var err error
		result.ResponseHeaders, err = exportStringMap(vm, object.Get("responseHeaders"))
		if err != nil {
			return TransformResult{}, fmt.Errorf("response headers: %w", err)
		}
		result.ResponseBody, result.ResponseBodySet, err = exportJSON(vm, object.Get("responseBody"), maxBytes)
		if err != nil {
			return TransformResult{}, fmt.Errorf("response body: %w", err)
		}
	}
	return result, nil
}

func exportStringMap(vm *goja.Runtime, value goja.Value) (map[string]string, error) {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, nil
	}
	object := value.ToObject(vm)
	result := make(map[string]string, len(object.Keys()))
	for _, key := range object.Keys() {
		result[key] = object.Get(key).ToString().String()
	}
	return result, nil
}

func exportJSON(vm *goja.Runtime, value goja.Value, maxBytes int) (json.RawMessage, bool, error) {
	if goja.IsUndefined(value) {
		return nil, false, nil
	}
	stringify, ok := goja.AssertFunction(vm.Get("JSON").ToObject(vm).Get("stringify"))
	if !ok {
		return nil, false, fmt.Errorf("JSON.stringify is unavailable")
	}
	serialized, err := stringify(goja.Undefined(), value)
	if err != nil {
		return nil, false, fmt.Errorf("JSON.stringify: %w", err)
	}
	if goja.IsUndefined(serialized) {
		return nil, false, nil
	}
	raw := json.RawMessage(serialized.String())
	if !json.Valid(raw) {
		return nil, false, fmt.Errorf("JSON.stringify returned invalid JSON")
	}
	if maxBytes > 0 && len(raw) > maxBytes {
		return nil, false, &WorkerError{Code: ErrorCodeOutputLimit, Message: "transform output exceeds configured limit"}
	}
	return raw, true, nil
}

func scriptError(err error) error {
	var interrupted *goja.InterruptedError
	if errors.As(err, &interrupted) {
		if signal, ok := interrupted.Value().(interruptSignal); ok {
			return &WorkerError{Code: signal.code, Message: signal.message, Retryable: true}
		}
		return &WorkerError{Code: ErrorCodeCanceled, Message: interrupted.Error(), Retryable: true}
	}
	return &WorkerError{Code: ErrorCodeScript, Message: err.Error()}
}

func cloneStrings(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func stringsAreBlank(value string) bool {
	for _, character := range value {
		if character != ' ' && character != '\n' && character != '\r' && character != '\t' {
			return false
		}
	}
	return true
}
