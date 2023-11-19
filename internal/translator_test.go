package internal_test

import (
	"context"
	"testing"

	"github.com/RamilGN/just-translate/internal"
)

func Test_Translator_Translate(t *testing.T) {
	ctx := context.Background()

	translator := internal.NewTranslator(
		nil,
		internal.LanguagePair{Lang1: internal.RU, Lang2: internal.EN},
	)

	actual, err := translator.Translate(ctx, "Hello")
	if err != nil {
		t.Fatal(err)
	}

	expected := "Привет"
	if expected != actual {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	actual, err = translator.Translate(ctx, "Привет")
	if err != nil {
		t.Fatal(err)
	}

	expected = "Hello"
	if expected != actual {
		t.Errorf("expected %v, actual %v", expected, actual)
	}
}
