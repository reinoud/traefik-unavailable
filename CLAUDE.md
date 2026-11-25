# traefik-unavailable

Klein, generiek container dat achter Traefik een nette 503-pagina levert zolang een applicatie niet bereikbaar is. Draait in een eigen compose-stack, los van de applicatie.

- Open werk staat in GitHub-issues van `reinoud/traefik-unavailable`.
- Werkwijze zoals bij botenloodsbeheer: elke wijziging met code in een eigen branch (`feature/<issue>-<kort>`), review door een verse agent vóór de commit (Sentinel; Aegis voor headers en hardening), pull request naar `main` met `Closes #<nr>`, de reviewers en de testuitkomst; de sessie mergt na akkoord zelf (`gh pr merge --squash --delete-branch`).
- Release: CHANGELOG, semver-tag, trivy-scan (CRITICAL/HIGH met oplossing stopt de release), SBOM (CycloneDX JSON) bij de release.
- Code schrijven en reviewen met de skill `ponytail`; teksten volgen de skill `schrijfstijl`.
- Agents schrijven hun terugblik in `werk/terugblik/` (niet in git).
- Dit is een publieke repository: geen tokens, serverpaden, domeinen of andere installatiegegevens.
