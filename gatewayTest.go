package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nautilus/gateway"
	"github.com/nautilus/graphql"
)

func witContentfulAuthInfo(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// look up the value of the Authorization header
		tokenValue := r.Header.Get("Authorization")

		// here is where you would perform some kind of validation on the token
		// but we're going to skip that for this example and just save it as the
		// id directly. PLEASE, DO NOT DO THIS IN PRODUCTION.

		// invoke the handler with the new context
		handler.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), "Authorization", tokenValue),
		))
	})
}

// the next thing we need to do is to modify the network requests to our services.
// To do this, we have to define a middleware that pulls the id of the user out
// of the context of the incoming request and sets it as the USER_ID header.
var forwardAuthContentful = gateway.RequestMiddleware(func(r *http.Request) error {
	// the initial context of the request is set as the same context
	// provided by net/http

	// we are safe to extract the value we saved in context and set it as the outbound header
	if bearerToken := r.Context().Value("Authorization"); bearerToken != nil {
		r.Header.Set("Authorization", "Bearer "+bearerToken.(string))
	}

	// return the modified request
	return nil
})

func main() {
	// introspect the apis
	schemas, err := graphql.IntrospectRemoteSchemas(
		"http://localhost:8080/",
		"http://localhost:8081/",
	)
	if err != nil {
		fmt.Println("error ", err)
		panic(err)
	}

	// create the gateway instance
	gw, err := gateway.New(schemas, gateway.WithMiddlewares(forwardAuthContentful))
	if err != nil {
		panic(err)
	}

	// add the playground endpoint to the router
	http.HandleFunc("/graphql", witContentfulAuthInfo(gw.PlaygroundHandler))

	// start the server
	fmt.Println("Starting server")
	err = http.ListenAndServe(":3001", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
