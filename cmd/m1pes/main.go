package main

import (
	"context"
	"log"

	"m1pes/internal/app"
)

func main() {
	ctx := context.Background()

	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err = a.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
