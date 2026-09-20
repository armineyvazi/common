// Package graphqlgo provides a ports.GraphQLHandler adapter backed by
// github.com/graph-gophers/graphql-go. It requires no code generation:
// pass in a SDL schema string and a resolver struct, and get back an
// http.Handler ready to mount at any path.
//
// Usage:
//
//	srv := graphqlgo.New(schemaString, &MyResolver{}, graphqlgo.Config{
//	    Pretty: true,
//	})
//	http.Handle("/graphql", srv.Handler())
package graphqlgo

import (
	"net/http"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Config controls optional behaviour of the GraphQL handler.
type Config struct {
	// Pretty enables pretty-printed JSON responses (useful for development).
	Pretty bool
	// MaxDepth limits query depth to prevent deeply nested abuse.
	// Zero means unlimited (not recommended for public APIs).
	MaxDepth int
}

type server struct {
	handler http.Handler
}

// New parses schemaSDL, binds resolver, and returns a ports.GraphQLHandler.
// It panics if the schema is invalid — fail fast at startup, not at request time.
func New(schemaSDL string, resolver any, cfg Config) ports.GraphQLHandler {
	opts := []graphql.SchemaOpt{
		graphql.UseFieldResolvers(),
	}
	if cfg.MaxDepth > 0 {
		opts = append(opts, graphql.MaxDepth(cfg.MaxDepth))
	}

	schema := graphql.MustParseSchema(schemaSDL, resolver, opts...)
	h := &relay.Handler{Schema: schema}

	var handler http.Handler = h
	if cfg.Pretty {
		handler = prettyHandler{h}
	}

	return &server{handler: handler}
}

// Handler returns the http.Handler that serves GraphQL POST requests.
func (s *server) Handler() http.Handler {
	return s.handler
}

// prettyHandler wraps relay.Handler and sets the Accept header to request
// pretty-printed JSON from the relay handler (relay checks for this).
type prettyHandler struct {
	h *relay.Handler
}

func (p prettyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Accept", "application/json")
	p.h.ServeHTTP(w, r)
}
