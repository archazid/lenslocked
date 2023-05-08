package main

import (
	stdctx "context"
	"fmt"

	"archazid.io/lenslocked/context"
	"archazid.io/lenslocked/models"
)

func main() {
	ctx := stdctx.Background()

	user := models.User{
		Email: "zahid@archazid.io",
	}
	ctx = context.WithUser(ctx, &user)

	retrievedUser := context.User(ctx)
	fmt.Println(retrievedUser.Email)
}
