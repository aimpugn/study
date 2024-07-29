# checkout/orders

## Request by cURL

```bash
curl --location 'https://api-m.sandbox.paypal.com/v2/checkout/orders' \
--header 'Content-Type: application/json' \
--header 'Prefer: return=representation' \
--header 'PayPal-Request-Id: 845bda9b-f387-4ea7-ab18-a57bf696fb2c' \
--header 'Authorization: Bearer A21AAJmS8oEonQxZUd_HAl05140dNu1vX1PxMPPuWOREjhu6v1vFzOlz_5p72r_b_dJ3aWd9jKd5IgZqGNPED6aJArdxcxwmg' \
--data '{
    "intent": "CAPTURE",
    "purchase_units": [
        {
            "items": [
                {
                    "name": "T-Shirt",
                    "description": "Green XL",
                    "quantity": "1",
                    "unit_amount": {
                        "currency_code": "USD",
                        "value": "100.00"
                    }
                }
            ],
            "amount": {
                "currency_code": "USD",
                "value": "100.00",
                "breakdown": {
                    "item_total": {
                        "currency_code": "USD",
                        "value": "100.00"
                    }
                }
            }
        }
    ],
    "application_context": {
        "return_url": "https://example.com/return",
        "cancel_url": "https://example.com/cancel"
    }
}'
```

```bash
root@tx-gateway-service-76fc8d8d59-6xss5:/# curl --location 'https://api-m.sandbox.paypal.com/v2/checkout/orders' \
--header 'Content-Type: application/json' \
--header 'Prefer: return=representation' \
--header 'PayPal-Request-Id: 845bda9b-f387-4ea7-ab18-a57bf696fb2c' \
--header 'Authorization: Bearer A21AAJmS8oEonQxZUd_HAl05140dNu1vX1PxMPPuWOREjhu6v1vFzOlz_5p72r_b_dJ3aWd9jKd5IgZqGNPED6aJArdxcxwmg' \
--data '{
    "intent": "CAPTURE",
    "purchase_units": [
        {
            "items": [
                {
                    "name": "T-Shirt",
                    "description": "Green XL",
                    "quantity": "1",
                    "unit_amount": {
                        "currency_code": "USD",
                        "value": "100.00"
                    }
                }
            ],
            "amount": {
                "currency_code": "USD",
                "value": "100.00",
                "breakdown": {
                    "item_total": {
                        "currency_code": "USD",
                        "value": "100.00"
                    }
                }
            }
        }
    ],
    "application_context": {
        "return_url": "https://example.com/return",
        "cancel_url": "https://example.com/cancel"
    }
}' -i
HTTP/2 406
content-type: text/html
server: nginx
strict-transport-security: max-age=31536000; includeSubDomains
accept-ranges: bytes
via: 1.1 varnish, 1.1 varnish
edge-control: max-age=120
date: Wed, 10 Apr 2024 07:04:26 GMT
x-served-by: cache-nrt-rjtf7700028-NRT, cache-icn1450061-ICN
x-cache: MISS, MISS
x-cache-hits: 0, 0
x-timer: S1712732666.873080,VS0,VE406
content-length: 156
<html>
<head><title>406 Not Acceptable</title></head>
<body>
<center><h1>406 Not Acceptable</h1></center>
<hr><center>nginx</center>
</body>
</html>
root@tx-gateway-service-76fc8d8d59-6xss5:/# curl --location 'https://api-m.sandbox.paypal.com/v2/checkout/orders' --header 'Content-Type: application/json' --header 'Prefer: return=representation' --header 'PayPal-Request-Id: 845bda9b-f387-4ea7-ab18-a57bf696fb2c' --header 'Authorization: Bearer A21AAJmS8oEonQxZUd_HAl05140dNu1vX1PxMPPuWOREjhu6v1vFzOlz_5p72r_b_dJ3aWd9jKd5IgZqGNPED6aJArdxcxwmg' --data '{
    "intent": "CAPTURE",
    "purchase_units": [
        {
            "items": [
                {
                    "name": "T-Shirt",
                    "description": "Green XL",
                    "quantity": "1",
                    "unit_amount": {
                        "currency_code": "USD",
                        "value": "100.00"
                    }
                }
            ],
            "amount": {
                "currency_code": "USD",
                "value": "100.00",
                "breakdown": {
                    "item_total": {
                        "currency_code": "USD",
                        "value": "100.00"
                    }
                }
            }
        }
    ],
    "application_context": {
        "return_url": "https://example.com/return",
        "cancel_url": "https://example.com/cancel"
    }
}'
<html>
<head><title>406 Not Acceptable</title></head>
<body>
<center><h1>406 Not Acceptable</h1></center>
<hr><center>nginx</center>
</body>
</html>
```
