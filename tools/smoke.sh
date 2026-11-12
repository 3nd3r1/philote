#!/bin/sh
# Builds the horse Tinder test spec, installs and typechecks the generated backend
# and frontend, runs the backend and checks real requests, both with curl and
# through the frontend's generated client. Needs network for npm install.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
out=${1:-$(mktemp -d)}
go run "$root/cmd/philote" build --frozen --out "$out" "$root/cmd/philote/testdata/neighbr/neighbr.philote"
cd "$out/api"
npm install --silent
npx tsc
export PORT=3999 DATA_FILE="$out/data.json"
rm -f "$DATA_FILE"
npx tsx src/server.ts >"$out/server.log" 2>&1 &
pid=$!
trap 'kill $pid' EXIT
until curl -s "localhost:$PORT" >/dev/null; do sleep 0.2; done

fails=0
# req METHOD PATH WANT_STATUS [BODY]: sends a request and sets $body.
req() {
  res=$(curl -s -w '\n%{http_code}' -X "$1" -H 'content-type: application/json' ${4:+-d "$4"} "localhost:$PORT$2")
  status=$(printf '%s' "$res" | tail -n1)
  body=$(printf '%s' "$res" | sed '$d')
  if [ "$status" = "$3" ]; then mark=ok; else mark=FAIL; fails=$((fails + 1)); fi
  printf '%-4s %-6s %-28s %s %s\n' "$mark" "$1" "$2" "$status" "$body"
}
field() { printf '%s' "$body" | node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>console.log(JSON.parse(s)[process.argv[1]]))' "$1"; }

req POST /horses 201 '{"name":"Sea Biscuit","breed":"mustang","age":7,"bio":"Fast. Loves apples."}'
biscuit=$(field id)
[ "$(field handle)" = "sea-biscuit" ] || { echo "FAIL handle is $(field handle)"; fails=$((fails + 1)); }
req POST /horses 400 '{"name":"Sea Biscuit","breed":"arabian","age":9,"bio":"The other one."}'
req POST /horses 400 '{"name":"Sparkle","breed":"unicorn","age":0,"bio":"Magic.","wings":2}'
req POST /horses 201 '{"name":"Mister Ed","breed":"other","age":12,"bio":"Talks a lot."}'
ed=$(field id)
req POST /swipes 201 "{\"liked\":true,\"horse_id\":\"$biscuit\",\"target_id\":\"$ed\"}"
req POST /swipes 400 "{\"liked\":true,\"horse_id\":\"$biscuit\",\"target_id\":\"nobody\"}"
req GET /swipes 200
req GET "/horses/$biscuit/swipes" 200
[ "$(printf '%s' "$body" | grep -c "$ed")" = 1 ] || { echo "FAIL biscuit's swipes don't include the swipe on ed: $body"; fails=$((fails + 1)); }
req GET /horses/nope/swipes 404
# The swipe endpoint runs the model-written steps pinned in philote.lock.
req POST /swipe 422 "{\"horse\":\"$biscuit\",\"target\":\"$biscuit\",\"liked\":true}"
req POST /swipe 422 "{\"horse\":\"$biscuit\",\"target\":\"$ed\",\"liked\":true}"
req GET /matches 200
[ "$body" = "[]" ] || { echo "FAIL a one-sided like made a match: $body"; fails=$((fails + 1)); }
req POST /swipe 200 "{\"horse\":\"$ed\",\"target\":\"$biscuit\",\"liked\":true}"
req GET /matches 200
count=$(printf '%s' "$body" | node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>console.log(JSON.parse(s).length))')
[ "$count" = 1 ] || { echo "FAIL a mutual like made $count matches"; fails=$((fails + 1)); }
req PATCH "/horses/$biscuit" 200 '{"age":8}'
req GET /horses/nope 404
req POST /matches 404 '{}'
# A delete is refused while anything points at the record (D-020).
req DELETE "/horses/$biscuit" 409
req POST /horses 201 '{"name":"Trigger","breed":"other","age":20,"bio":"Retired."}'
trigger=$(field id)
req DELETE "/horses/$trigger" 204
req GET "/horses/$trigger" 404

[ "$fails" = 0 ] || { echo "smoke: $fails failed"; exit 1; }

cd "$out/web"
npm install --silent
npm run lint --silent >/dev/null
# Typechecking the script too proves the client's result types narrow:
# comparing an error to one the endpoint doesn't declare fails here.
cp "$root/tools/smoke-client.ts" smoke.mts
npx next typegen >/dev/null
npx tsc --noEmit
# react-server is the condition Next.js resolves server-only with on the server.
API_URL="http://localhost:$PORT" "$out/api/node_modules/.bin/tsx" --conditions=react-server smoke.mts

# The pages, rendered by the built app against the running backend.
NEXT_TELEMETRY_DISABLED=1 npx next build >"$out/next-build.log" 2>&1 || { cat "$out/next-build.log"; exit 1; }
API_URL="http://localhost:$PORT" npx next start -p 3998 >"$out/web.log" 2>&1 &
web=$!
trap 'kill $pid $web' EXIT
until curl -s localhost:3998 >/dev/null; do sleep 0.2; done
# page PATH TEXT: the page must render TEXT once its data streams in.
page() {
  if curl -s "localhost:3998$1" | grep -q "$2"; then echo "ok   page $1 shows $2"; else echo "FAIL page $1 doesn't show $2"; fails=$((fails + 1)); fi
}
page /horses "Sea Biscuit"
page /horses "/horses/$biscuit"
page "/horses/$biscuit" "sea-biscuit"
page "/horses/$biscuit" "Loves apples"
page /horses/nope "That horse doesn"
page /matches "Mister Ed"
page /matches "Shadowfax"
[ "$fails" = 0 ] || { echo "smoke: $fails failed"; exit 1; }
echo "smoke: all requests behaved"
