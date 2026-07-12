#!/usr/bin/env bash
set -euo pipefail

base=${BASE_URL:-http://localhost:8080}
resp=""

say() { printf '\n== %s\n' "$1"; }

req() {
    local expect=$1 method=$2 path=$3 body=${4:-}
    local out status
    if [ -n "$body" ]; then
        out=$(curl -s -w '\n%{http_code}' -X "$method" "$base$path" \
            -H 'Content-Type: application/json' -d "$body")
    else
        out=$(curl -s -w '\n%{http_code}' -X "$method" "$base$path")
    fi
    status=${out##*$'\n'}
    resp=${out%$'\n'*}
    if [ "$status" != "$expect" ]; then
        echo "FAIL: $method $path returned $status, expected $expect" >&2
        echo "$resp" >&2
        exit 1
    fi
    echo "   $method $path -> $status"
}

id_of() { echo "$resp" | sed -E 's/.*"id":"([^"]+)".*/\1/'; }

confirm() {
    local expect=$1 payment=$2 file=$3
    local out status
    out=$(curl -s -w '\n%{http_code}' -X PUT "$base/api/payments/$payment/confirm" \
        -F proof=@"$file")
    status=${out##*$'\n'}
    resp=${out%$'\n'*}
    if [ "$status" != "$expect" ]; then
        echo "FAIL: confirm payment $payment returned $status, expected $expect" >&2
        echo "$resp" >&2
        exit 1
    fi
    echo "   PUT /api/payments/$payment/confirm -> $status"
}

pickup_wait() { sleep 1.2; }

say "waiting for the api"
for i in $(seq 1 30); do
    if curl -sf "$base/health" > /dev/null 2>&1; then
        break
    fi
    if [ "$i" = 30 ]; then
        echo "FAIL: api did not become healthy at $base" >&2
        exit 1
    fi
    sleep 1
done
echo "   api is up"

tomorrow=$(date -u -v+1d +%Y-%m-%dT09:00:00Z 2>/dev/null || date -u -d '+1 day' +%Y-%m-%dT09:00:00Z)

say "registering households"
req 201 POST /api/households '{"owner_name":"Budi Santoso","address":"Jl. Melati No. 1, Surabaya"}'
budi=$(id_of)
req 201 POST /api/households '{"owner_name":"Sari Lestari","address":"Jl. Anggrek No. 2, Surabaya"}'
sari=$(id_of)
req 201 POST /api/households '{"owner_name":"Andi Wijaya","address":"Jl. Kenanga No. 3, Surabaya"}'
andi=$(id_of)
req 200 GET "/api/households?page=1&limit=10"
req 200 GET "/api/households/$budi"

say "a household without records can be deleted, and is gone afterwards"
req 201 POST /api/households '{"owner_name":"Dewi Kusuma","address":"Jl. Mawar No. 4, Surabaya"}'
dewi=$(id_of)
req 200 DELETE "/api/households/$dewi"
req 404 GET "/api/households/$dewi"

say "budi: full happy path, pickup completed and payment confirmed with proof"
req 201 POST /api/pickups "{\"household_id\":\"$budi\",\"type\":\"paper\"}"
budi_paper=$(id_of)
req 200 PUT "/api/pickups/$budi_paper/schedule" "{\"pickup_date\":\"$tomorrow\"}"
req 200 PUT "/api/pickups/$budi_paper/complete"

budi_payment=$(curl -s "$base/api/payments?household_id=$budi&status=pending" \
    | sed -E 's/.*"id":"([^"]+)".*/\1/')
proof_dir=$(mktemp -d)
trap 'rm -rf "$proof_dir"' EXIT
printf 'transfer receipt' > "$proof_dir/proof.jpg"
confirm 200 "$budi_payment" "$proof_dir/proof.jpg"

say "budi: a second pickup, scheduled and waiting"
pickup_wait
req 201 POST /api/pickups "{\"household_id\":\"$budi\",\"type\":\"plastic\"}"
budi_plastic=$(id_of)
req 200 PUT "/api/pickups/$budi_plastic/schedule" "{\"pickup_date\":\"$tomorrow\"}"
req 200 GET "/api/pickups?household_id=$budi&status=scheduled"

say "sari: completed pickup leaves a pending payment behind"
pickup_wait
req 201 POST /api/pickups "{\"household_id\":\"$sari\",\"type\":\"plastic\"}"
sari_plastic=$(id_of)
req 200 PUT "/api/pickups/$sari_plastic/schedule" "{\"pickup_date\":\"$tomorrow\"}"
req 200 PUT "/api/pickups/$sari_plastic/complete"
sari_payment=$(curl -s "$base/api/payments?household_id=$sari&status=pending" \
    | sed -E 's/.*"id":"([^"]+)".*/\1/')

say "sari: the pending payment now blocks her next pickup"
pickup_wait
req 422 POST /api/pickups "{\"household_id\":\"$sari\",\"type\":\"paper\"}"

say "electronic pickups need a safety check"
pickup_wait
req 400 POST /api/pickups "{\"household_id\":\"$budi\",\"type\":\"electronic\"}"
pickup_wait
req 201 POST /api/pickups "{\"household_id\":\"$budi\",\"type\":\"electronic\",\"safety_check\":true}"

say "andi: a fresh organic request and a canceled one"
pickup_wait
req 201 POST /api/pickups "{\"household_id\":\"$andi\",\"type\":\"organic\"}"
pickup_wait
req 201 POST /api/pickups "{\"household_id\":\"$andi\",\"type\":\"paper\"}"
andi_paper=$(id_of)
req 200 PUT "/api/pickups/$andi_paper/cancel"

say "auto cancel: an organic request left unscheduled past 72 hours"
pickup_wait
req 201 POST /api/pickups "{\"household_id\":\"$andi\",\"type\":\"organic\"}"
stale=$(id_of)
docker compose exec -T postgres psql -q -U postgres -d community_waste \
    -c "UPDATE waste_pickups SET created_at = now() - interval '73 hours' WHERE id = '$stale'" > /dev/null
echo "   backdated created_at of $stale by 73 hours"
echo "   waiting for the worker (runs every AUTOCANCEL_INTERVAL)"
canceled=""
for i in $(seq 1 90); do
    found=$(curl -s "$base/api/pickups?household_id=$andi&status=canceled")
    if echo "$found" | grep -q "\"id\":\"$stale\""; then
        canceled=yes
        break
    fi
    sleep 1
done
if [ -z "$canceled" ]; then
    echo "FAIL: pickup $stale was not auto canceled within 90s" >&2
    exit 1
fi
echo "   worker canceled the stale organic pickup"

say "wrong states and blocked deletes are rejected"
req 409 PUT "/api/pickups/$andi_paper/schedule" "{\"pickup_date\":\"$tomorrow\"}"
confirm 409 "$budi_payment" "$proof_dir/proof.jpg"
req 409 DELETE "/api/households/$budi"

say "manual payment entry for the scheduled plastic, confirmed with proof"
req 422 POST /api/payments "{\"household_id\":\"$sari\",\"waste_id\":\"$budi_plastic\",\"amount\":50000}"
req 201 POST /api/payments "{\"household_id\":\"$budi\",\"waste_id\":\"$budi_plastic\",\"amount\":50000}"
manual_payment=$(id_of)
req 200 GET "/api/payments?status=pending&household_id=$budi"
confirm 200 "$manual_payment" "$proof_dir/proof.jpg"

say "rate limit kicks in after the burst"
sleep 5
limited=""
for i in $(seq 1 6); do
    code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$base/api/pickups" \
        -H 'Content-Type: application/json' \
        -d "{\"household_id\":\"$sari\",\"type\":\"paper\"}")
    printf '   attempt %s -> %s\n' "$i" "$code"
    if [ "$code" = 429 ]; then
        limited=yes
    fi
done
if [ -z "$limited" ]; then
    echo "FAIL: never saw a 429 while hammering pickup creation" >&2
    exit 1
fi

say "reports over the simulated data"
req 200 GET /api/reports/waste-summary
echo "   $resp"
req 200 GET /api/reports/payment-summary
echo "   $resp"
req 200 GET "/api/reports/households/$budi/history"
req 404 GET /api/reports/households/00000000-0000-0000-0000-000000000001/history

say "done, postman variables"
echo "   household_id = $budi"
echo "   pickup_id    = $budi_plastic"
echo "   payment_id   = $sari_payment"
