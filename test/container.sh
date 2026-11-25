#!/bin/sh
# Start het image zoals in het compose-voorbeeld (alleen-lezen, niet-root, zonder capabilities)
# en controleert het van buitenaf. Gebruik: sh test/container.sh <image>
set -eu
image=${1:?image}
fout=0
check() { if eval "$2"; then echo "ok   $1"; else echo "FOUT $1"; fout=$((fout + 1)); fi; }

id=$(docker run -d --read-only --cap-drop ALL --security-opt no-new-privileges:true \
  -e TITEL='<b>vet</b>' -e RETRY_AFTER=12 -p 127.0.0.1::8080 "$image")
trap 'docker rm -f "$id" >/dev/null' EXIT
poort=$(docker port "$id" 8080 | head -1 | sed 's/.*://')

i=0
until [ "$(docker inspect -f '{{.State.Health.Status}}' "$id")" = healthy ] || [ $i -ge 20 ]; do
  sleep 1; i=$((i + 1))
done
check "healthcheck van het image is healthy" '[ "$(docker inspect -f "{{.State.Health.Status}}" "$id")" = healthy ]'
check "draait alleen-lezen" '[ "$(docker inspect -f "{{.HostConfig.ReadonlyRootfs}}" "$id")" = true ]'
check "draait als uid 65532, niet als root" '[ "$(docker top "$id" -o pid,uid | awk "NR==2 {print \$2}")" = 65532 ]'

h=$(curl -s -o /dev/null -D - -H 'Accept: text/html' "http://127.0.0.1:$poort/willekeurig?x=1" | tr -d '\r')
check "503 met Retry-After en no-store" 'echo "$h" | grep -q "^HTTP/1.1 503" && echo "$h" | grep -qi "^retry-after: 12$" && echo "$h" | grep -qi "^cache-control: no-store$"'
check "beveiligingsheaders" 'echo "$h" | grep -qi "^content-security-policy: default-src .none." && echo "$h" | grep -qi "^x-content-type-options: nosniff$" && echo "$h" | grep -qi "^referrer-policy: no-referrer$"'
check "HTML met ge-escapete titel" 'curl -s -H "Accept: text/html" "http://127.0.0.1:$poort/" | grep -q "&lt;b&gt;vet&lt;/b&gt;"'
check "JSON zonder Accept" '[ "$(curl -s -o /dev/null -w "%{content_type} %{http_code}" "http://127.0.0.1:$poort/")" = "application/json 503" ]'

echo "$fout fout(en)"
[ "$fout" -eq 0 ]
