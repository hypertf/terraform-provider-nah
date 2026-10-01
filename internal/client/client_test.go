package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/prefix/v1/projects" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing authentication")
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["slug"] != "slug" || body["name"] != "Display Name" {
			t.Errorf("bad project request: %v", body)
		}
		w.WriteHeader(201)
		_, _ = fmt.Fprint(w, `{"id":"opaque","slug":"slug","name":"Display Name"}`)
	}))
	defer server.Close()
	p, err := NewClient(server.URL+"/prefix/", "secret").CreateProject(t.Context(), "slug", "Display Name")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "opaque" || p.Slug != "slug" {
		t.Fatalf("bad decoded project: %+v", p)
	}
}

func TestResponseErrorsAndRedaction(t *testing.T) {
	for _, status := range []int{301, 400, 401, 404, 409, 429, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Location", "/redirect")
				w.WriteHeader(status)
				_, _ = fmt.Fprint(w, `{"error":"secret-content"}`)
			}))
			defer server.Close()
			_, err := NewClient(server.URL, "secret-token").GetProject(t.Context(), "project")
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
				t.Fatalf("bad error: %v", err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("response leaked secret")
			}
			if IsNotFound(err) != (status == 404) {
				t.Fatal("incorrect not-found classification")
			}
			if calls != 1 {
				t.Fatalf("unexpected retries/redirects: %d", calls)
			}
		})
	}
}

func TestMalformedResponses(t *testing.T) {
	for _, body := range []string{"", "null", "not-json", `{"id":42}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = fmt.Fprint(w, body) }))
			defer server.Close()
			if _, err := NewClient(server.URL, "test").GetProject(t.Context(), "p"); err == nil {
				t.Fatal("accepted malformed resource")
			}
		})
	}
}

func TestCancellationAndEscaping(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := NewClient("http://127.0.0.1:1", "test").GetProject(ctx, "p"); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
	if got := Route("projects", "a/b?c", "instances", "x#y"); got != "/v1/projects/a%2Fb%3Fc/instances/x%23y" {
		t.Fatalf("unsafe route: %s", got)
	}
}

func TestMissingAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/api-keys/deleted" {
			t.Error("expected direct key lookup")
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	if _, err := NewClient(server.URL, "test").GetAPIKey(t.Context(), "deleted"); !IsNotFound(err) {
		t.Fatalf("missing key: %v", err)
	}
}

func BenchmarkReadProject(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"id":"opaque","slug":"project","name":"Project"}`)
	}))
	defer server.Close()
	c := NewClient(server.URL, "test")
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := c.GetProject(b.Context(), "project"); err != nil {
				b.Error(err)
			}
		}
	})
}
