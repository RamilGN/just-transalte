package internal

import (
	"context"
	"errors"
	"fmt"
)

var errTranslator = errors.New("[translator]")

type Language string

const (
	RU               Language = "ru"
	EN               Language = "en"
	LanguageCodeAuto Language = "auto"
)

func languages() map[Language]struct{} {
	return map[Language]struct{}{
		RU: {},
		EN: {},
	}
}

type LanguagePair struct {
	Lang1 Language
	Lang2 Language
}

type translation struct {
	detected    Language
	translation string
}

type TranslateClient interface {
	translate(ctx context.Context, from, to Language, text string) (translation, error)
}

type Translator struct {
	client             TranslateClient
	availableLanguages map[Language]struct{}
	languagePair       LanguagePair
}

func NewTranslator(languagePair LanguagePair) *Translator {
	translator := new(Translator)
	translator.client = newGoogleTranslateClient(nil)
	translator.availableLanguages = languages()
	translator.languagePair = languagePair

	return translator
}

func (t *Translator) Translate(ctx context.Context, text string) (string, error) {
	_, ok := t.availableLanguages[t.languagePair.Lang1]
	if !ok {
		return "", fmt.Errorf("%w undefined Lang1: %s", errTranslator, t.languagePair.Lang1)
	}

	_, ok = t.availableLanguages[t.languagePair.Lang2]
	if !ok {
		return "", fmt.Errorf("%w undefined Lang2: %s", errTranslator, t.languagePair.Lang2)
	}

	res1, err := t.client.translate(ctx, t.languagePair.Lang1, t.languagePair.Lang2, text)
	if err != nil {
		return "", fmt.Errorf("%w %w", errTranslator, err)
	}

	res2, err := t.client.translate(ctx, LanguageCodeAuto, t.languagePair.Lang1, text)
	if err != nil {
		return "", fmt.Errorf("%w %w", errTranslator, err)
	}

	var translation string
	if res2.detected == t.languagePair.Lang2 {
		translation = res2.translation
	} else {
		translation = res1.translation
	}

	return translation, nil
}
