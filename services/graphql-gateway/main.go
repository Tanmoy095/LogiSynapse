package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Tanmoy095/LogiSynapse/graphql-gateway/graph"
	"github.com/Tanmoy095/LogiSynapse/graphql-gateway/graph/generated"
)

func main() {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL", "/query"))
	http.Handle("/query", srv)

	log.Println("🚀 Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
