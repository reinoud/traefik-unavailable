# Changelog

## 1.0.0 (2026-11-25)

Eerste versie (#1).

- Antwoordt op elk pad en elke methode met 503, `Retry-After` en `Cache-Control: no-store`.
- HTML voor clients die `text/html` accepteren, anders JSON.
- Instelbaar met `TITEL`, `TEKST`, `TAAL`, `RETRY_AFTER` en `JSON_SLEUTEL`; teksten worden ge-escaped. Een eigen pagina kan op `/pagina/index.html` worden gemount.
- Zet zelf `Content-Security-Policy`, `X-Content-Type-Options` en `Referrer-Policy`; geen HSTS.
- Healthcheck op poort 8081.
- Statische Go-binary op `scratch`, als uid 65532, alleen-lezen te draaien, voor amd64 en arm64.
- Release met Trivy-scan en SBOM (CycloneDX JSON) per platform.
