package internal

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestTranslate(t *testing.T) {
	ctx := context.Background()

	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Request:    req,
			Body: io.NopCloser(
				strings.NewReader(
					`[[["Привет","Hello",null,null,10]],null,"en",null,null,null,1,[],[["en"],null,[1],["en"]]]`,
				),
			),
		}
	})

	gclient := newGoogleTranslateClient(httpClient)

	actual, err := gclient.translate(ctx, EN, RU, "Hello")
	if err != nil {
		t.Fatal(err)
	}

	expected := translation{detected: "en", translation: "Привет"}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, actual %v", expected, actual)
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}
