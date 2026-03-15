package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"shorter/internal/http-server/handlers/url/save"
	"shorter/internal/lib/api"
	"shorter/internal/lib/random"
	"testing"
)

const (
	host     = "localhost:8080"
	user     = "user"
	password = "user"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/v1/url").WithJSON(save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}).WithBasicAuth(user, password).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("alias").ContainsKey("status").NotContainsKey("error")
}

func TestURLShortener_Unauthorized(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/v1/url").WithJSON(save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}
			e := httpexpect.Default(t, u.String())
			resp := e.POST("/v1/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth(user, password).
				Expect().Status(http.StatusOK).
				JSON().Object()
			if tc.error != "" {
				resp.NotContainsKey("alias")

				resp.Value("error").String().IsEqual(tc.error)

				return
			}
			alias := tc.alias
			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}
			testRedirect(t, alias, tc.url)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}
	redirectedUrl, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectedUrl)
}
