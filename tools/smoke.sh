#!/bin/sh
# Builds the shop test spec, installs and typechecks the generated backend,
# runs it and checks real requests. Needs network for npm install.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
out=${1:-$(mktemp -d)}
go run "$root/cmd/philote" build --out "$out" "$root/cmd/philote/testdata/shop/shop.philote"
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

req POST /customers 201 '{"email":"ada@example.com","name":"Ada"}'
req POST /customers 400 '{"email":"ada@example.com","name":"Ada again"}'
req POST /customers 400 '{"email":"not-an-email","name":"Bob","age":3}'
req GET /customers 200
req POST /products 201 '{"name":"Blue Mug","price_cents":1200,"stock":3}'
mug=$(field id)
[ "$(field slug)" = "blue-mug" ] || { echo "FAIL slug is $(field slug)"; fails=$((fails + 1)); }
req POST /products 400 '{"name":"Free","price_cents":-1}'
req PATCH "/products/$mug" 200 '{"stock":5}'
req GET "/products/$mug" 200
req GET /products/nope 404
req POST /orders 404 '{}'
req DELETE "/products/$mug" 204
req GET "/products/$mug" 404

[ "$fails" = 0 ] && echo "smoke: all requests behaved" || { echo "smoke: $fails failed"; exit 1; }
