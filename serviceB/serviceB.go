package main

import (
	"log"
	"net/http"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

var Schema = `
	schema {
		query: Query
	}

	interface Node {
		id: ID!
	}

	type User implements Node {
		id: ID!
		lastName: String!
	}

	type Query {
		node(id: ID!): Node
		allUsers: [User!]!
	}


	type UnionTest1 {
		test1string: String
	}

	type UnionTest2 {
		test2string: String
	}
	  

	union additional_details = UnionTest1 | UnionTest2
`

// the users by id
var users = map[string]*User{
	"1": {
		id:       "1",
		lastName: "Aivazis",
	},
}

// type resolvers

type UnionTest1 struct {
	test1string string
}

type UnionTest2 struct {
	test2string string
}

type User struct {
	id       graphql.ID
	lastName string
}

func (u *User) ID() graphql.ID {
	return u.id
}

func (u *User) LastName() string {
	return u.lastName
}

type Node interface {
	ID() graphql.ID
}

type NodeResolver struct {
	node Node
}

func (n *NodeResolver) ID() graphql.ID {
	return n.node.ID()
}

func (n *NodeResolver) ToUser() (*User, bool) {
	user, ok := n.node.(*User)
	return user, ok
}

// func (n *NodeResolver) To

// query resolvers

type queryB struct{}

func (q *queryB) Node(args struct{ ID string }) *NodeResolver {
	user := users[args.ID]

	if user != nil {
		return &NodeResolver{user}
	} else {
		return nil
	}
}

func (q *queryB) AllUsers() []*User {
	// build up a list of all the users
	userSlice := []*User{}

	for _, user := range users {
		userSlice = append(userSlice, user)
	}

	return userSlice
}

func main() {
	// attach the schema to the resolver object
	schema := graphql.MustParseSchema(Schema, &queryB{})

	// make sure we add the user info to the execution context
	http.Handle("/", &relay.Handler{Schema: schema})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
