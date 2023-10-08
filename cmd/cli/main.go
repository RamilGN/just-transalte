package main

import (
	"context"
	"fmt"
	"log"

	"github.com/RamilGN/just-translate/internal/translateclients"
)

func main() {
	client := translateclients.NewGoogleTranslateClient()

	res, err := client.Translate(context.TODO(), "привет", "ru", "en")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Translation, res.From, res.To)
}
