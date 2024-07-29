# es via curl

## examples

```bash
curl -XGET \
    -u 'port-es:bFkTWBT@ehyV9kknX*!G' \
    -s \
    https://es.some.domain.co/payments-v2/_doc/store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test \
    | jq '._source'
```
