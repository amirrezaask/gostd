package gostd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

var (
	CompilerMode string

	//You can set this variable inside your config code if you have another way of doing configuration.
	DebugMode = os.Getenv("DEBUG_MODE") == "1"
)

const (
	std_context_contextual_info_key = "___std_contextual_info___"
	std_context_file_key            = "___std_file____"
	std_context_line_key            = "___std_line____"
	std_debug_mode_string           = "debug"
)

func Debug() bool {
	return CompilerMode == std_debug_mode_string
}

func AddFunctionInfoToContext(ctx context.Context) context.Context {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		newC := context.WithValue(ctx, std_context_file_key, file)
		return context.WithValue(newC, std_context_line_key, line)
	}

	return ctx

}

func AddContextualInfoToContext(ctx context.Context, kvs ...any) context.Context {
	if len(kvs)%2 == 0 {
		return ctx
	}

	info := map[string]any{}

	for i := 0; i < len(kvs); i++ {
		if i%2 == 0 {
			info[fmt.Sprint(kvs[i])] = kvs[i+1]
		}
	}

	if ctx.Value(std_context_contextual_info_key) != nil {
		oldInfo := ctx.Value(std_context_contextual_info_key).(map[string]any)
		for k, v := range oldInfo {
			if _, exists := info[k]; !exists {
				info[k] = v
			}
		}
	}

	return context.WithValue(ctx, std_context_contextual_info_key, info)
}

func LogDebug(ctx context.Context, msg string, kvs ...any) {
	if Debug() {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			kvs = append(kvs, "file", file, "line", line)
		}
	}

	if ctxInfo := ctx.Value(std_context_contextual_info_key); ctxInfo != nil {
		if ctxInfoMap, ok := ctxInfo.(map[string]any); ok {
			for k, v := range ctxInfoMap {
				kvs = append(kvs, k, v)
			}
		}

	}
	slog.Debug(msg, kvs...)
}

func LogError(ctx context.Context, msg string, kvs ...any) {
	var file string
	var line int
	if Debug() {
		_, file, line, _ = runtime.Caller(1)
	}

	if ctx.Value(std_context_file_key) != nil {
		file = ctx.Value(std_context_file_key).(string)
	}
	if ctx.Value(std_context_line_key) != nil {
		line = ctx.Value(std_context_line_key).(int)
	}

	if file != "" {
		kvs = append(kvs, "file", file)
	}
	if line != 0 {
		kvs = append(kvs, "line", line)
	}

	if ctxInfo := ctx.Value(std_context_contextual_info_key); ctxInfo != nil {
		if ctxInfoMap, ok := ctxInfo.(map[string]any); ok {
			for k, v := range ctxInfoMap {
				kvs = append(kvs, k, v)
			}
		}

	}
	slog.Error(msg, kvs...)
}

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(item T)    { s[item] = struct{}{} }
func (s Set[T]) Remove(item T) { delete(s, item) }

func Assert(expr bool, msgs ...string) {
	if Debug() {
		_, file, line, ok := runtime.Caller(1)
		msg := "Assert Failed:"
		if ok {
			msg += fmt.Sprintf("File:'%s'", file)
			msg += fmt.Sprintf("Line:'%d'", line)
		}

		if len(msgs) > 0 {
			msg += " " + msgs[0]
		}
		panic(msg)
	} else {
		ErrorLog(context.Background(), "Assert failed")
	}
}

func RetryDo(f func() error, retries int, backoff time.Duration) error {
	err := f()
	if err != nil {
		for i := 0; i < retries; i++ {
			time.Sleep(backoff)
			err = f()
			if err != nil {
				slog.Error("error in running with retries", "err", err, "retry", i)
				continue
			}
			break

		}
	}

	return err
}
