package main

import (
	"context"
	"fmt"

	"github.com/unipost-dev/sdk-go/unipost"
)

func main() {
	// Reads UNIPOST_API_KEY from environment automatically
	client := unipost.NewClient()

	ctx := context.Background()

	// List connected accounts
	accounts, err := client.Accounts.List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d connected accounts\n", len(accounts))

	if len(accounts) == 0 {
		fmt.Println("Connect an account first at https://app.unipost.dev")
		return
	}

	// Collect account IDs
	ids := make([]string, len(accounts))
	for i, a := range accounts {
		ids[i] = a.ID
	}

	// Create a post
	post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
		Caption:    "Hello from UniPost Go SDK!",
		AccountIDs: ids,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Post created: %s (status: %s)\n", post.ID, post.Status)
}
