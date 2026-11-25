package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func start(t *testing.T, m map[string]string, eigenPagina string) *httptest.Server {
	t.Helper()
	v, err := nieuw(env(m), eigenPagina)
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(v)
	t.Cleanup(s.Close)
	return s
}

func vraag(t *testing.T, methode, url, accept string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(methode, url, nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	c := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	r, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	return r, string(b)
}

func controleerHeaders(t *testing.T, r *http.Response) {
	t.Helper()
	for k, v := range map[string]string{
		"Cache-Control":           "no-store",
		"Content-Security-Policy": csp,
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "no-referrer",
	} {
		if got := r.Header.Get(k); got != v {
			t.Errorf("%s = %q, wil %q", k, got, v)
		}
	}
	for _, k := range []string{"Strict-Transport-Security", "Set-Cookie", "Location"} {
		if r.Header.Get(k) != "" {
			t.Errorf("onverwachte header %s", k)
		}
	}
}

func TestElkPadElkeMethode503(t *testing.T) {
	s := start(t, map[string]string{"RETRY_AFTER": "45"}, "")
	for _, pad := range []string{"/", "/a/b?zoek=geheim", "//dubbel", "/x/../y", "/healthz", "/style.css"} {
		for _, m := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "FOO"} {
			for _, accept := range []string{"", "text/html"} {
				r, _ := vraag(t, m, s.URL+pad, accept)
				if r.StatusCode != 503 || r.Header.Get("Retry-After") != "45" {
					t.Errorf("%s %s: %d, Retry-After %q", m, pad, r.StatusCode, r.Header.Get("Retry-After"))
				}
				controleerHeaders(t, r)
			}
		}
	}
}

func TestFormaatVolgtAccept(t *testing.T) {
	s := start(t, map[string]string{"TEKST": "Even weg", "JSON_SLEUTEL": "melding"}, "")
	for accept, html := range map[string]bool{
		"text/html,application/xhtml+xml,*/*;q=0.8": true,
		"application/json, TEXT/HTML;q=0.5":         true,
		"text/html;q=0":                             false,
		"application/json":                          false,
		"*/*":                                       false,
		"":                                          false,
	} {
		r, body := vraag(t, "GET", s.URL+"/", accept)
		ct := r.Header.Get("Content-Type")
		if html {
			if ct != "text/html; charset=utf-8" || !strings.Contains(body, "<h1>") {
				t.Errorf("Accept %q: wil HTML, kreeg %s", accept, ct)
			}
			continue
		}
		var j map[string]string
		if ct != "application/json" || json.Unmarshal([]byte(body), &j) != nil || j["melding"] != "Even weg" {
			t.Errorf("Accept %q: wil JSON, kreeg %s %s", accept, ct, body)
		}
	}
}

func TestTekstenWordenGeEscaped(t *testing.T) {
	kwaad := `<script>alert(1)</script>"&'`
	s := start(t, map[string]string{"TITEL": kwaad, "TEKST": kwaad, "TAAL": `nl" onload="x`}, "")
	_, html := vraag(t, "GET", s.URL+"/", "text/html")
	if strings.Contains(html, "<script>") || strings.Contains(html, `" onload`) || !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("HTML niet ge-escaped:\n%s", html)
	}
	_, body := vraag(t, "GET", s.URL+"/", "")
	var j map[string]string
	if strings.Contains(body, "<") || json.Unmarshal([]byte(body), &j) != nil || j["fout"] != kwaad {
		t.Errorf("JSON niet ge-escaped: %s", body)
	}
}

func TestEigenPagina(t *testing.T) {
	f := filepath.Join(t.TempDir(), "index.html")
	os.WriteFile(f, []byte("<p>eigen</p>"), 0o644)
	s := start(t, map[string]string{"TITEL": "genegeerd"}, f)
	r, body := vraag(t, "GET", s.URL+"/", "text/html")
	if r.StatusCode != 503 || body != "<p>eigen</p>" {
		t.Errorf("eigen pagina: %d %q", r.StatusCode, body)
	}
}

func TestOngeldigeRetryAfter(t *testing.T) {
	for _, w := range []string{"-1", "abc", "1.5"} {
		if _, err := nieuw(env(map[string]string{"RETRY_AFTER": w}), ""); err == nil {
			t.Errorf("RETRY_AFTER=%q geaccepteerd", w)
		}
	}
}

func TestHealthcheck(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(gezond))
	defer s.Close()
	r, _ := vraag(t, "GET", s.URL+"/", "")
	if r.StatusCode != 200 {
		t.Errorf("healthcheck: %d", r.StatusCode)
	}
	controleerHeaders(t, r)
}
