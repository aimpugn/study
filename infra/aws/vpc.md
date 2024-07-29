# vpc

## vpc?

## vpc 내에서 curl

```bash
# api-a3 서버 IP 주소는 10.0.5.4
ubuntu@api-a3:~$ curl http://ip-10-0-5-238.ap-northeast-2.compute.internal/api/docs -vvv
*   Trying 10.0.5.238...
* Connected to ip-10-0-5-238.ap-northeast-2.compute.internal (10.0.5.238) port 80 (#0)
> GET /api/docs HTTP/1.1
> Host: ip-10-0-5-238.ap-northeast-2.compute.internal
> User-Agent: curl/7.47.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Wed, 27 Dec 2023 06:05:06 GMT
< Server: Apache
< X-Content-Type-Options: nosniff
< X-XSS-Protection: 1; mode=block
< X-Frame-Options: SAMEORIGIN
< Transfer-Encoding: chunked
< Content-Type: application/json
<
{
    "swagger": "2.0",
```
