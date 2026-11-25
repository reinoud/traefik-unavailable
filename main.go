// Vangnet: antwoordt op elk verzoek op :8080 met 503 (HTML of JSON) en op :8081 met 200 voor de healthcheck.
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const csp = "default-src 'none'; style-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// html/template escapet elke waarde uit de omgeving, ook in het attribuut lang.
var pagina = template.Must(template.New("pagina").Parse(`<!doctype html>
<html lang="{{.Taal}}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Titel}}</title>
</head>
<body>
<h1>{{.Titel}}</h1>
<p>{{.Tekst}}</p>
</body>
</html>
`))

type vangnet struct {
	html, json []byte
	retry      string
}

func waarde(env func(string) string, naam, standaard string) string {
	if v := env(naam); v != "" {
		return v
	}
	return standaard
}

// nieuw leest de instellingen één keer bij de start; een fout stopt het container.
func nieuw(env func(string) string, eigenPagina string) (*vangnet, error) {
	retry, err := strconv.Atoi(waarde(env, "RETRY_AFTER", "30"))
	if err != nil || retry < 0 {
		return nil, fmt.Errorf("RETRY_AFTER moet een geheel aantal seconden zijn (0 of meer)")
	}
	tekst := waarde(env, "TEKST", "De applicatie is even niet bereikbaar. Probeer het over een minuut opnieuw.")
	j, err := json.Marshal(map[string]string{waarde(env, "JSON_SLEUTEL", "fout"): tekst})
	if err != nil {
		return nil, err
	}
	v := &vangnet{json: j, retry: strconv.Itoa(retry)}

	if v.html, err = os.ReadFile(eigenPagina); err == nil {
		return v, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	var b strings.Builder
	err = pagina.Execute(&b, map[string]string{
		"Taal":  waarde(env, "TAAL", "nl"),
		"Titel": waarde(env, "TITEL", "Even geduld"),
		"Tekst": tekst,
	})
	v.html = []byte(b.String())
	return v, err
}

func beveilig(h http.Header) {
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Security-Policy", csp)
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
}

// wilHTML is waar als Accept text/html noemt met een q groter dan 0.
func wilHTML(accept []string) bool {
	for _, regel := range accept {
		for _, deel := range strings.Split(regel, ",") {
			mt, p, err := mime.ParseMediaType(deel)
			if err != nil || mt != "text/html" {
				continue
			}
			if p["q"] == "" {
				return true
			}
			q, err := strconv.ParseFloat(p["q"], 64)
			return err == nil && q > 0
		}
	}
	return false
}

// ServeHTTP logt niets: geen pad, querystring of headers.
func (v *vangnet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	beveilig(h)
	h.Set("Retry-After", v.retry)
	body := v.json
	h.Set("Content-Type", "application/json")
	if wilHTML(r.Header.Values("Accept")) {
		body = v.html
		h.Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write(body)
}

func gezond(w http.ResponseWriter, _ *http.Request) {
	beveilig(w.Header())
	w.Write([]byte("ok\n"))
}

// Geen ServeMux: die stuurt redirects bij paden als //a of /a/../b.
func luister(adres string, h http.Handler) error {
	s := &http.Server{Addr: adres, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	return s.ListenAndServe()
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-health" { // HEALTHCHECK van het image; scratch heeft geen curl
		r, err := (&http.Client{Timeout: 2 * time.Second}).Get("http://127.0.0.1:8081/")
		if err != nil || r.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	v, err := nieuw(os.Getenv, "/pagina/index.html")
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(luister(":8081", http.HandlerFunc(gezond))) }()
	log.Print("vangnet luistert op :8080, healthcheck op :8081")
	log.Fatal(luister(":8080", v))
}
