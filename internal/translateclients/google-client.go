package translateclients

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

var errGoogleClient = errors.New("google-client")

const (
	googleClientTimeout    = 1 * time.Second
	googleTranslateBaseURL = "https://translate.googleapis.com/translate_a/single"

	googleDtTranslateT  = "t"  // Translation of source.
	googleDtTranslateAT = "at" // Alternative translation of source.
	googleDtTranslateRM = "rm" // Transliteration of source.
	googleDtTranslateBD = "bd" // Synonyms of source.
)

type GoogleTranslation struct {
	From        string
	To          string
	Translation string
}

type GoogleTranslateClient struct {
	httpclient  *http.Client
	url         *url.URL
	queryParams url.Values
}

func NewGoogleTranslateClient() *GoogleTranslateClient {
	client := new(GoogleTranslateClient)
	client.httpclient = &http.Client{Timeout: googleClientTimeout}

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

func (g *GoogleTranslateClient) Translate(ctx context.Context, q, sl, tl string) (GoogleTranslation, error) {
	g.queryParams.Set("q", q)
	g.queryParams.Set("sl", sl)
	g.queryParams.Set("tl", tl)

	url, _ := url.Parse(googleTranslateBaseURL)
	url.RawQuery = g.queryParams.Encode()

	resp, err := g.get(ctx, url.String())
	if err != nil {
		return GoogleTranslation{}, err
	}

	level1, ok := resp[0].([]any)
	if !ok {
		return GoogleTranslation{}, fmt.Errorf("%w: tried to get level 1", errGoogleClient)
	}

	translations := []string{}

	for _, v := range level1 {
		level2, ok := v.([]any)
		if !ok {
			return GoogleTranslation{}, fmt.Errorf("%w: tried to get level 2", errGoogleClient)
		}

		translation, ok := level2[0].(string)
		if !ok {
			return GoogleTranslation{}, fmt.Errorf("%w: tried to get level 2", errGoogleClient)
		}

		translations = append(translations, translation)
	}

	return GoogleTranslation{
		From:        sl,
		To:          tl,
		Translation: strings.Join(translations, ""),
	}, nil
}

func (g *GoogleTranslateClient) get(ctx context.Context, url string) ([]any, error) {
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

func (g *GoogleTranslateClient) newRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGoogleClient, err)
	}

	req.Header.Set(headerAccept, mediaTypeJSON)
	req.Header.Set(headerContentType, mediaTypeJSON)

	return req, nil
}

func (g *GoogleTranslateClient) do(req *http.Request) ([]any, error) {
	resp, err := g.httpclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGoogleClient, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: response status code is %d", errGoogleClient, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGoogleClient, err)
	}

	translation := new([]any)

	err = json.Unmarshal(data, translation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal google translation: %w", err)
	}

	return *translation, nil
}
