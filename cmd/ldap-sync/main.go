package main

import (
	"context"
	"fmt"
	"ldap-sync/internal/ldap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	client := ldap.GenerateNewClientWithAuthToken("https://lldap.picklemustard.dev")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	results, err := client.FetchAllUsers(ctx)

	if err != nil {
		fmt.Printf("Got an error: %w", err)
	}

	for _, user := range results.Users {
		fmt.Println(user)
	}

}
