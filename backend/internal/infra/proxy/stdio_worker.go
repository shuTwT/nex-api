package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func WorkerEnvironment(path string, envVars map[string]string) ([]string, error) {
	env := map[string]string{"PATH": path}
	for key, value := range envVars {
		if key == "PATH" {
			return nil, errors.New("mcp gateway: service envVars cannot override PATH")
		}
		if !validEnvironmentKey(key) {
			return nil, fmt.Errorf("mcp gateway: invalid service environment key %q", key)
		}
		env[key] = value
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+env[key])
	}
	return result, nil
}

func validEnvironmentKey(key string) bool {
	if key == "" || (key[0] != '_' && (key[0] < 'A' || key[0] > 'Z') && (key[0] < 'a' || key[0] > 'z')) {
		return false
	}
	for _, character := range key[1:] {
		if character != '_' && (character < 'A' || character > 'Z') && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func workerScript(command string, maxMemoryBytes int64) string {
	limitKB := (maxMemoryBytes + 1023) / 1024
	return "ulimit -v " + strconv.FormatInt(limitKB, 10) + " >/dev/null 2>&1 || true\n" + command
}

func writeWorkerInput(ctx context.Context, stream *stdioStream, input []byte) error {
	payload := append([]byte(nil), input...)
	if len(payload) > 0 && payload[len(payload)-1] != '\n' {
		payload = append(payload, '\n')
	}
	done := make(chan error, 1)
	go func() {
		if len(payload) > 0 {
			_, err := stream.stdin.Write(payload)
			if err != nil {
				done <- fmt.Errorf("mcp gateway: write stdio input: %w", err)
				return
			}
		}
		done <- stream.stdin.Close()
	}()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("mcp gateway: close stdio input: %w", err)
		}
		return nil
	case <-ctx.Done():
		stream.terminate()
		return ctx.Err()
	}
}

func startMemoryWatchdog(ctx context.Context, processID int, limitBytes int64, onLimit func()) func() {
	watchContext, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		limitKB := limitBytes / 1024
		for {
			select {
			case <-watchContext.Done():
				return
			case <-ticker.C:
				rss, err := processGroupRSS(watchContext, processID)
				if err == nil && rss > limitKB {
					onLimit()
					return
				}
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

func processGroupRSS(ctx context.Context, processID int) (int64, error) {
	command := exec.CommandContext(ctx, "/bin/ps", "-axo", "pid=,pgid=,rss=")
	command.Env = []string{"PATH=" + defaultWorkerPath}
	output, err := command.Output()
	if err != nil {
		return 0, fmt.Errorf("mcp gateway: inspect worker memory: %w", err)
	}
	var total int64
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		pgid, err := strconv.Atoi(fields[1])
		if err != nil || pgid != processID {
			continue
		}
		rss, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			continue
		}
		total += rss
	}
	return total, nil
}

func ParseServiceEnvVars(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}
	var values map[string]string
	decoder := json.NewDecoder(strings.NewReader(raw))
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("mcp gateway: parse service envVars: %w", err)
	}
	if values == nil {
		return map[string]string{}, nil
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("mcp gateway: service envVars contains multiple JSON values")
	}
	return values, nil
}
