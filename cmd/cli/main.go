package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/RamilGN/just-translate/internal"
)

func main() {
	ctx := context.Background()

	var (
		language1 string
		language2 string
		text      string
	)

	flag.StringVar(&language1, "l1", "", "language 1")
	flag.StringVar(&language2, "l2", "", "language 2")
	flag.StringVar(&text, "t", "", "text")
	flag.Parse()

	translator := internal.NewTranslator(nil, internal.LanguagePair{
		Lang1: internal.Language(language1),
		Lang2: internal.Language(language2),
	})

	translation, err := translator.Translate(ctx, text)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%v\n", translation)
}
