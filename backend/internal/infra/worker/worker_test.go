package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFrame_roundTripsAndRejectsOversizedPayload(t *testing.T) {
	// Given
	want := []byte(`{"type":"execute","job_id":"job-1"}`)
	var wire bytes.Buffer

	// When
	if err := WriteFrame(&wire, want, 1024); err != nil {
		t.Fatalf("write frame: %v", err)
	}
	got, err := ReadFrame(&wire, 1024)

	// Then
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("frame payload = %q, want %q", got, want)
	}
	if err := WriteFrame(&bytes.Buffer{}, bytes.Repeat([]byte("x"), 5), 4); !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("oversized frame error = %v, want ErrFrameTooLarge", err)
	}
}

func TestEngine_preservesPreAndPostTransformSemantics(t *testing.T) {
	// Given
	engine := NewEngine(EngineOptions{MaxOutputBytes: 64 * 1024})
	pre := Job{
		ID:      "pre-1",
		Kind:    ScriptKindPre,
		Script:  `headers["x-added"] = "yes"; query.page = "2"; body.n = body.n + 1;`,
		Headers: map[string]string{"x-original": "keep"},
		Query:   map[string]string{"page": "1"},
		Body:    json.RawMessage(`{"n":1}`),
	}
	post := Job{
		ID:              "post-1",
		Kind:            ScriptKindPost,
		Script:          `responseHeaders["x-script"] = "yes"; responseBody.ok = true;`,
		ResponseHeaders: map[string]string{"x-upstream": "keep"},
		ResponseBody:    json.RawMessage(`{"ok":false}`),
	}

	// When
	preResult, preErr := engine.Execute(context.Background(), pre)
	postResult, postErr := engine.Execute(context.Background(), post)

	// Then
	if preErr != nil || postErr != nil {
		t.Fatalf("execute transforms: pre=%v post=%v", preErr, postErr)
	}
	if preResult.Headers["x-added"] != "yes" || preResult.Query["page"] != "2" {
		t.Fatalf("pre result = %#v, want merged fields", preResult)
	}
	if string(preResult.Body) != `{"n":2}` || !preResult.BodySet {
		t.Fatalf("pre body = %s (set=%t), want {\"n\":2}", preResult.Body, preResult.BodySet)
	}
	if postResult.ResponseHeaders["x-script"] != "yes" || string(postResult.ResponseBody) != `{"ok":true}` || !postResult.ResponseBodySet {
		t.Fatalf("post result = %#v, want response mutation", postResult)
	}
}

func TestEngine_doesNotExposeHostGlobals(t *testing.T) {
	// Given
	engine := NewEngine(EngineOptions{MaxOutputBytes: 64 * 1024})
	job := Job{
		ID:     "globals-1",
		Kind:   ScriptKindPre,
		Script: `body = { process: typeof process, require: typeof require, fetch: typeof fetch, fs: typeof fs };`,
		Body:   json.RawMessage(`{}`),
	}

	// When
	result, err := engine.Execute(context.Background(), job)

	// Then
	if err != nil {
		t.Fatalf("execute globals probe: %v", err)
	}
	var globals map[string]string
	if err := json.Unmarshal(result.Body, &globals); err != nil {
		t.Fatalf("decode globals %s: %v", result.Body, err)
	}
	for name, value := range globals {
		if value != "undefined" {
			t.Fatalf("global %s = %q, want undefined", name, value)
		}
	}
}

func TestEngine_returnsStructuredCancellationAndOutputErrors(t *testing.T) {
	// Given
	engine := NewEngine(EngineOptions{MaxOutputBytes: 32})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When
	_, cancelErr := engine.Execute(ctx, Job{ID: "cancel-1", Kind: ScriptKindPre, Script: `while (true) {}`})
	_, outputErr := engine.Execute(context.Background(), Job{
		ID:     "output-1",
		Kind:   ScriptKindPre,
		Script: `body = { value: "this is larger than the output budget" };`,
		Body:   json.RawMessage(`null`),
	})

	// Then
	var cancelWorkerErr *WorkerError
	if !errors.As(cancelErr, &cancelWorkerErr) || cancelWorkerErr.Code != ErrorCodeCanceled {
		t.Fatalf("cancel error = %v, want structured canceled error", cancelErr)
	}
	var outputWorkerErr *WorkerError
	if !errors.As(outputErr, &outputWorkerErr) || outputWorkerErr.Code != ErrorCodeOutputLimit {
		t.Fatalf("output error = %v, want structured output limit error", outputErr)
	}
}

func TestPool_recyclesWorkerAndCancelsLongRunningJob(t *testing.T) {
	// Given
	if runtime.GOOS == "windows" {
		t.Skip("the worker process uses Unix resource limits")
	}
	workerPath := buildWorker(t)
	pool, err := NewPool(context.Background(), PoolOptions{
		Executable:  workerPath,
		WorkerCount: 1,
		// Race instrumentation makes the first IPC round trip substantially
		// slower on loaded builders; cancellation is exercised by its own
		// 50 ms context below, so keep the pool timeout comfortably above startup.
		JobTimeout:     10 * time.Second,
		CancelGrace:    500 * time.Millisecond,
		MaxJobs:        1,
		MaxMemoryBytes: 256 << 20,
		MaxOutputBytes: 64 * 1024,
		MaxFrameBytes:  256 * 1024,
	})
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	defer pool.Close()

	// When
	first, firstErr := pool.Execute(context.Background(), Job{
		Kind:   ScriptKindPre,
		Script: `headers["x-worker"] = "one";`,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, cancelErr := pool.Execute(ctx, Job{Kind: ScriptKindPre, Script: `while (true) {}`})
	second, secondErr := pool.Execute(context.Background(), Job{
		Kind:         ScriptKindPost,
		Script:       `responseBody = { recycled: true };`,
		ResponseBody: json.RawMessage(`{}`),
	})

	// Then
	if firstErr != nil || first.Headers["x-worker"] != "one" {
		t.Fatalf("first result=%#v err=%v", first, firstErr)
	}
	var cancelWorkerErr *WorkerError
	if !errors.As(cancelErr, &cancelWorkerErr) || cancelWorkerErr.Code != ErrorCodeCanceled {
		t.Fatalf("pool cancel error = %v, want structured canceled error", cancelErr)
	}
	if secondErr != nil || !second.ResponseBodySet || string(second.ResponseBody) != `{"recycled":true}` {
		t.Fatalf("recycled result=%#v err=%v", second, secondErr)
	}
}

func buildWorker(t *testing.T) string {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	backend := filepath.Clean(filepath.Join(root, "..", "..", ".."))
	path := filepath.Join(t.TempDir(), "script-worker")
	cmd := exec.Command("go", "build", "-o", path, "./cmd/script-worker")
	cmd.Dir = backend
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build script-worker: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	return path
}
