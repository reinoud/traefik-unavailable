# traefik-unavailable

Klein, generiek container dat achter Traefik een nette 503-pagina levert zolang een applicatie niet bereikbaar is. Draait in een eigen compose-stack, los van de applicatie.

- Open werk staat in GitHub-issues van `reinoud/traefik-unavailable`.
- Werkwijze zoals bij botenloodsbeheer: elke wijziging met code in een eigen branch (`feature/<issue>-<kort>`), review door een verse agent vóór de commit (Sentinel; Aegis voor headers en hardening), pull request naar `main` met `Closes #<nr>`, de reviewers en de testuitkomst; de sessie mergt na akkoord zelf (`gh pr merge --squash --delete-branch`).
- Release: CHANGELOG, semver-tag, trivy-scan (CRITICAL/HIGH met oplossing stopt de release), SBOM (CycloneDX JSON) bij de release.
- Code schrijven en reviewen met de skill `ponytail`; teksten volgen de skill `schrijfstijl`.
- Agents schrijven hun terugblik in `werk/terugblik/` (niet in git).
- Dit is een publieke repository: geen tokens, serverpaden, domeinen of andere installatiegegevens.

## Briefingchecklist (lessen uit terugblikken)

- Een containertest vergelijkt op de verwachte waarde (`= 65532`), nooit op ongelijkheid; een lege uitkomst moet falen.
- Multi-arch lokaal: tijdelijke builder (`docker buildx create --driver docker-container`), bouw naar een OCI-tar in de scratchpad, scan met `trivy --input`, verwijder de builder daarna.
- Tijdelijke mappen in de scratchpad met `mktemp -d`; geen `rm -rf`. Tests bij voorkeur op een alleen-lezen mount van de repo.
- Noem in een reviewopdracht welke tags een release moet opleveren, en bij een nareview welke stappen de gewijzigde instelling gebruiken.
- `actionlint` draait via `docker run --rm -v "$PWD":/repo:ro -w /repo rhysd/actionlint:1.7.7 -no-color`.
