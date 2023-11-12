package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var errGoogleClient = errors.New("[google-client]")

const (
	headerAccept      = "Accept"
	headerContentType = "Content-Type"
	mediaTypeJSON     = "application/json"

	googleClientTimeout    = 1 * time.Second
	googleTranslateBaseURL = "https://translate.googleapis.com/translate_a/single"

	googleDtTranslateT  = "t"  // Translation of source.
	googleDtTranslateAT = "at" // Alternative translation of source.
	googleDtTranslateRM = "rm" // Transliteration of source.
	googleDtTranslateBD = "bd" // Synonyms of source.
)

type googleTranslateClient struct {
	httpclient  *http.Client
	url         *url.URL
	queryParams url.Values
}

func newGoogleTranslateClient(httpClient *http.Client) *googleTranslateClient {
	client := new(googleTranslateClient)
	if httpClient == nil {
		client.httpclient = &http.Client{Timeout: googleClientTimeout}
	} else {
		client.httpclient = httpClient
	}

	clientURL, err := url.Parse(googleTranslateBaseURL)
	if err != nil {
		log.Fatal(err)
	}

	client.url = clientURL
	client.queryParams = url.Values{}
	client.queryParams.Set("client", "gtx")
	client.queryParams.Set("dt", googleDtTranslateT)

	return client
}

func (g *googleTranslateClient) translate(
	ctx context.Context,
	from, to Language,
	text string,
) (translation, error) {
	g.queryParams.Set("q", text)
	g.queryParams.Set("sl", string(from))
	g.queryParams.Set("tl", string(to))

	url, _ := url.Parse(googleTranslateBaseURL)
	url.RawQuery = g.queryParams.Encode()

	resp, err := g.get(ctx, url.String())
	if err != nil {
		return translation{}, err
	}

	level1, ok := resp[0].([]any)
	if !ok {
		return translation{}, fmt.Errorf("%w tried to get level 1", errGoogleClient)
	}

	translations := []string{}

	for _, v := range level1 {
		level2, ok := v.([]any)
		if !ok {
			return translation{}, fmt.Errorf("%w tried to get level 2", errGoogleClient)
		}

		translationString, ok := level2[0].(string)
		if !ok {
			return translation{}, fmt.Errorf("%w tried to get level 2", errGoogleClient)
		}

		translations = append(translations, translationString)
	}

	detectedLanguage, ok := resp[2].(string)
	if !ok {
		return translation{}, fmt.Errorf("%w tried to get detected language", errGoogleClient)
	}

	return translation{
		detected:    Language(detectedLanguage),
		translation: strings.Join(translations, ""),
	}, nil
}

func (g *googleTranslateClient) get(ctx context.Context, url string) ([]any, error) {
	req, err := g.newRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	translation, err := g.do(req)
	if err != nil {
		return nil, err
	}

	return translation, nil
}

func (g *googleTranslateClient) newRequest(
	ctx context.Context,
	method, url string,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %w", errGoogleClient, err)
	}

	req.Header.Set(headerAccept, mediaTypeJSON)
	req.Header.Set(headerContentType, mediaTypeJSON)

	return req, nil
}

func (g *googleTranslateClient) do(req *http.Request) ([]any, error) {
	resp, err := g.httpclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %w", errGoogleClient, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w response status code is %d", errGoogleClient, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("%w %w", errGoogleClient, err)
	}

	translation := new([]any)

	err = json.Unmarshal(data, translation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal google translation: %w", err)
	}

	return *translation, nil
}
