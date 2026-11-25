# ADR 0001: statische Go-binary op scratch

Datum: 2026-11-25. Status: voorgesteld. Issue: #1.

## Context

Het vangnet moet op elk pad en elke methode 503 geven, met headers die het zelf zet. Het kiest per verzoek HTML of JSON op basis van `Accept`. Teksten uit de omgeving moeten in HTML en in JSON ge-escaped worden. De healthcheck valt buiten de 503. Het image draait als niet-root op een alleen-lezen bestandssysteem, voor amd64 en arm64.

Issue #1 noemt twee kandidaten: `nginx:<versie>-alpine` met een vaste configuratie, of een statische Go-binary op `scratch` of `distroless`.

## Besluit

Een statische Go-binary, alleen met de standaardbibliotheek, op `scratch`. De builder is `golang:<versie>-alpine`, vastgepind op digest. Het image draait als uid 65532 en luistert op 8080 (vangnet) en 8081 (healthcheck).

## Overwegingen

Escaping beslist de keuze. nginx kan een omgevingsvariabele invullen met `envsubst` in de entrypoint, maar escapet die niet voor HTML of JSON. Daarvoor is njs nodig of een eigen script dat bij de start een pagina schrijft. Dat script is eigen code in shell, op het punt waar fouten tot XSS leiden. In Go doen `html/template` en `encoding/json` het escapen, en `mime.ParseMediaType` leest de `Accept`-header. Het geheel is ongeveer 140 regels Go, met evenveel regels test.

Alleen-lezen pleit ook voor Go. De entrypoint van het nginx-image schrijft bij de start naar `/etc/nginx/conf.d` en nginx schrijft naar `/var/cache/nginx` en `/var/run`. Met `read_only` vraagt dat `tmpfs`-mounts in elk compose-bestand. De Go-binary schrijft niets en heeft geen `tmpfs` nodig.

Op `scratch` staat alleen de binary in het image (ongeveer 9 MB). Er zijn geen OS-pakketten die Trivy kan melden; alleen de Go-standaardbibliotheek telt. Een melding daarin lost een nieuwe Go-patch op. Bij het bouwen meldde Trivy vier HIGH-kwetsbaarheden in Go 1.27.1 en nog één in 1.27.2; met 1.27.4 waren het er nul.

`distroless/static` voegt `/etc/passwd`, certificaten en tijdzonedata toe. Het vangnet maakt geen uitgaande TLS-verbindingen en gebruikt geen tijdzones, dus `scratch` volstaat. De gebruiker staat als getal in `USER`.

## Gevolgen

- De `HEALTHCHECK` kan geen `curl` gebruiken. De binary heeft daarom een optie `-health` die poort 8081 opvraagt.
- `http.ServeMux` wordt niet gebruikt, omdat die redirects stuurt bij paden als `//a`. Eén handler beantwoordt alles.
- Go-patches gaan via Dependabot (Docker-ecosysteem) naar de builder. Een nieuwe release is nodig om een melding in de standaardbibliotheek op te lossen.
- Multi-arch bouwt zonder emulatie: de builder draait op het eigen platform en compileert voor het doelplatform.
