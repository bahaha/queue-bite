package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
)

type ServerComponent func(*TestServer) error

type TestServer struct {
	Port     int
	envVars  map[string]string
	cleanups []func(context.Context) error
}

func randomPort(t *testing.T, vars map[string]string) string {
	t.Helper()

	if port, ok := vars["SERVER_PORT"]; ok {
		return port
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get random port on localhost: %v", err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port
	listener.Close()
	return strconv.Itoa(port)
}

func NewTestServer(t *testing.T, vars map[string]string, comps ...ServerComponent) *TestServer {
	t.Helper()

	port := randomPort(t, vars)
	envVars := map[string]string{
		"APP_ENV":     "test",
		"SERVER_PORT": port,
	}
	for k, v := range vars {
		envVars[k] = v
	}

	p, _ := strconv.Atoi(port)

	ts := &TestServer{
		Port:    p,
		envVars: envVars,
	}

	for _, comp := range comps {
		if err := comp(ts); err != nil {
			t.Fatalf("failed to attach server component: %v", err)
		}
	}

	t.Cleanup(func() {
		ctx := context.Background()
		for _, cleanup := range ts.cleanups {
			if err := cleanup(ctx); err != nil {
				t.Errorf("failed to cleanup server component: %v", err)
			}
		}
	})

	return ts
}

func (s *TestServer) Env(key string) string {
	v, ok := s.envVars[key]
	if ok {
		return v
	}
	return os.Getenv(key)
}

func (s *TestServer) URL(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", s.Port, path)
}

func (s *TestServer) Run(t *testing.T, runFunc func(context.Context, []string, func(string) string, io.Reader, io.Writer, io.Writer) error) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		if err := runFunc(ctx, nil, s.Env, nil, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Errorf("could not run the server: %v", err)
			return
		}
	}()
}
