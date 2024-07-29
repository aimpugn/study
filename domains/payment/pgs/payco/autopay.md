# autopay

## 자동 결제 삭제

```bash
curl --location 'https://alpha-api-bill.payco.com/outseller/autoPayment/delete' \
--header 'Content-Type: application/json' \
--data-raw '{
    "sellerKey": "AUTOPAY",
    "sellerAutoPaymentReferenceKey": "rody@some.domain",
    "autoPaymentCertifyKey": "yR9a2melqtgeuekyQh82Dp96DMzmRWDwTcI8GI2z8If4XcFZcPMhD5CvROJ_Mzb7LnQSBZaPuOpXRVCdv96odQ=="
}'
```

```json
// 응답
{"code":0,"message":"success"}
```
