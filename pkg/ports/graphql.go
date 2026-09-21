package ports

import "net/http"

// GraphQLHandler returns an http.Handler that serves a GraphQL endpoint.
// Mount it at any path using your HTTP router.
type GraphQLHandler interface {
	Handler() http.Handler
}
