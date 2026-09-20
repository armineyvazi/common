package graphqlgo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/graphql/graphqlgo"
)

const testSchema = `
	type Query {
		hello: String!
	}
`

type helloResolver struct{}

func (r *helloResolver) Hello() string { return "world" }

func TestNew_HandlerResponds(t *testing.T) {
	srv := graphqlgo.New(testSchema, &helloResolver{}, graphqlgo.Config{})

	body, _ := json.Marshal(map[string]string{"query": `{ hello }`})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Data struct {
			Hello string `json:"hello"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Data.Hello != "world" {
		t.Fatalf("expected hello=world, got %q", result.Data.Hello)
	}
}

func TestNew_InvalidSchema_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid schema")
		}
	}()
	graphqlgo.New("type Query { bad syntax !!!}", &helloResolver{}, graphqlgo.Config{})
}
