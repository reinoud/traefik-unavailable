# traefik-unavailable

Een klein container dat achter Traefik een 503-pagina levert zolang een applicatie niet bereikbaar is. Traefik haalt de routers van een applicatie weg zodra die niet healthy is en antwoordt dan met een kale `404 page not found`. Dit container vangt dat op met een router met prioriteit 1: elke router van de applicatie wint, zolang die bestaat.

Image: `ghcr.io/reinoud/traefik-unavailable`, voor `linux/amd64` en `linux/arm64`. Tags: de versie (`1.0.0`) en de major (`1`).

## Gedrag

Op poort 8080 antwoordt het container op elk pad en elke methode met `503 Service Unavailable`. Een browser die `text/html` accepteert, krijgt een HTML-pagina. Elke andere client krijgt JSON, bijvoorbeeld `{"fout":"De applicatie is even niet bereikbaar."}`. Ook `Accept: */*` en een verzoek zonder `Accept` krijgen JSON.

Elk antwoord heeft deze headers:

- `Retry-After` (standaard 30 seconden);
- `Cache-Control: no-store`;
- `Content-Security-Policy: default-src 'none'; style-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'`;
- `X-Content-Type-Options: nosniff` en `Referrer-Policy: no-referrer`.

Het container zet geen `Strict-Transport-Security`: dat beleid hoort bij de applicatie en het domein. Het stuurt geen redirects en geen cookies, en logt geen verzoeken. In de log staat alleen de startregel of een fout in de instellingen.

De healthcheck luistert op poort 8081 en geeft op elk pad 200. Die poort hoort niet bij de router in Traefik. Het image heeft een eigen `HEALTHCHECK`.

## Instellingen

| Variabele | Standaard | Betekenis |
|-----------|-----------|-----------|
| `TITEL` | `Even geduld` | Titel en kop van de pagina |
| `TEKST` | `De applicatie is even niet bereikbaar. Probeer het over een minuut opnieuw.` | Alinea op de pagina en tekst in de JSON |
| `TAAL` | `nl` | Attribuut `lang` van de pagina |
| `RETRY_AFTER` | `30` | Seconden in `Retry-After`; een ongeldige waarde stopt het container bij de start |
| `JSON_SLEUTEL` | `fout` | Sleutel in de JSON |

Teksten worden ge-escaped: HTML in `TITEL` of `TEKST` verschijnt als tekst. De pagina heeft geen CSS, geen script en geen externe bronnen. Ze gebruikt de standaardopmaak van de browser en is leesbaar op 360 px.

Een eigen pagina kan als bestand worden gemount op `/pagina/index.html`. Browsers krijgen dan dat bestand, ongewijzigd en nog steeds met status 503 en dezelfde headers. `TITEL` en `TAAL` gelden dan niet; de JSON gebruikt nog wel `TEKST` en `JSON_SLEUTEL`. De CSP staat geen inline CSS of script toe en het container levert geen andere bestanden, dus een eigen pagina is HTML zonder opmaak.

## Compose-voorbeeld

Het vangnet draait in een eigen compose-stack, zodat het blijft draaien als de stack van de applicatie herstart. Het sluit aan op het netwerk dat de applicatie met Traefik deelt. Dat netwerk is voor deze stack extern. Vul `<installatie>` in met de naam van de applicatiestack, en zet `DOMEIN` in een `.env` naast dit bestand.

```yaml
name: vangnet

services:
  vangnet:
    image: ghcr.io/reinoud/traefik-unavailable:1
    restart: unless-stopped
    read_only: true
    cap_drop: [ALL]
    security_opt: [no-new-privileges:true]
    environment:
      TITEL: "Even geduld"
      TEKST: "De applicatie start op. Probeer het over een minuut opnieuw."
    networks: [voorkant]
    labels:
      - traefik.enable=true
      - traefik.docker.network=<installatie>_voorkant
      - traefik.http.services.<installatie>-vangnet.loadbalancer.server.port=8080
      - traefik.http.routers.<installatie>-vangnet.rule=Host(`${DOMEIN}`)
      - traefik.http.routers.<installatie>-vangnet.priority=1
      - traefik.http.routers.<installatie>-vangnet.entrypoints=websecure
      - traefik.http.routers.<installatie>-vangnet.tls=true

networks:
  voorkant:
    external: true
    name: <installatie>_voorkant
```

De stack van de applicatie moet eerst draaien, anders bestaat het netwerk nog niet. Een router zonder expliciete prioriteit krijgt in Traefik de lengte van zijn regel als prioriteit, en die is altijd groter dan 1.

Controle na `docker compose up -d`:

```sh
curl -i -H 'Accept: text/html' https://<domein>/willekeurig
curl -i -H 'Accept: application/json' https://<domein>/willekeurig
```

Beide geven 503 met `Retry-After`, `Cache-Control: no-store` en de beveiligingsheaders, zolang de applicatie niet bereikbaar is.

## Bouwen en testen

```sh
docker build -t traefik-unavailable:test .
sh test/container.sh traefik-unavailable:test
```

De Go-tests draaien tijdens `docker build`. `test/container.sh` start het image alleen-lezen, zonder capabilities en met `no-new-privileges`, en controleert het van buitenaf. Een release (`.github/workflows/release.yml`) scant het image per platform met Trivy en zet per platform een SBOM in CycloneDX JSON bij de release. De keuze voor de basis staat in [docs/adr/0001-basis.md](docs/adr/0001-basis.md).
