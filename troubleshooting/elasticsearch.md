# Elasticsearch

## `[500 Internal Server Error]`

### 문제

golang elasticsearch로 문서를 인덱스 하려고 할 때 에러 발생

```go
func (app *Application) Index(config *GenerateJsonConfig) {
    fmt.Println("Index called")
    doc, _ := app.ReadJson(app.GenerateFilePath(config))

    // idAsInterface, ok := doc["id"]
    // if !ok {
    //     panic("no id exists")
    // }
    // id := idAsInterface.(string)
    marshaled, _ := json.Marshal(doc)

    exists, err := app.EsClient.Exists(PaymentsV2, "store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-merchant_1679558308933")
    if err != nil {
        return
    }
    log.Println("exists is", exists)
    documentId := DEMO_NICE_STORE_ID + "-rody_large_transactions_test"

    exists, err = app.EsClient.Exists(PaymentsV2, documentId)
    if err != nil {
        return
    }
    log.Println(documentId, "exists is", exists)

    response, err := app.EsClient.Index(PaymentsV2,
        bytes.NewReader(marshaled),
        app.EsClient.Index.WithDocumentID(documentId),
        app.EsClient.Index.WithRefresh("true"),
    )
    if err != nil {
        log.Fatal(fmt.Sprintf("[Index] failed %v", err))
    }
    log.Println("response is", response)
}
```

### 원인

- 로거 붙임

```go
func newElasticsearchClient(cfg *ApplicationConfig) (*elasticsearch.Client, error) {
    header := http.Header{}
    header.Set("Content-Type", "application/json; charset=UTF-8")
    client, err := elasticsearch.NewClient(elasticsearch.Config{
        Logger:                  &elastictransport.ColorLogger{Output: os.Stdout},
```

- 결과 로그

```log
  HEAD https://es.some.domain.co/payments-v2/_doc/store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-merchant_1679558308933 200 OK 155ms
2023/09/07 14:32:55 exists is [200 OK]
  HEAD https://es.some.domain.co/payments-v2/_doc/store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test 404 Not Found 52ms
2023/09/07 14:32:55 store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test exists is [404 Not Found]
   PUT https://es.some.domain.co/payments-v2/_doc/store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test?refresh=true 500 Server Error 38ms
2023/09/07 14:32:56 response is [500 Internal Server Error]
```

`PUT`으로 문서 생성하려고 해서 실패한 것으로 보임 -> 아니네

```shell
curl -X GET "https://es.some.domain.co/_security/user/_has_privileges" -u 'orgname-elasticsearch:bFkTWBT@ehyV9kknX*!G' -H "Content-Type: application/json" -d '{
  "index": [
    {
      "names": ["payments-v2"],
      "privileges": ["create", "index", "write"]
    }
  ]
}'
```

```shell
curl -X GET "https://es.some.domain.co/_cluster/health" -u 'orgname-elasticsearch:bFkTWBT@ehyV9kknX*!G
```

```shell
curl -X GET "https://es.some.domain.co/_security/_authenticate" -u 'orgname-elasticsearch:bFkTWBT@ehyV9kknX*!G'
```

```shell
curl -X DELETE "https://es.some.domain.co/payments-v2/_doc/store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test" -u 'orgname-elasticsearch:bFkTWBT@ehyV9kknX*!G'
```

```shell
❯ curl -X PUT "https://es.some.domain.co/payments-v2/_doc/store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test" \
    -u 'orgname-elasticsearch:bFkTWBT@ehyV9kknX*!G' \
    -H 'Content-Type: application/json' \
    --data-binary "@body.json"

# {"_index":"payments-v2","_id":"store-caae780e-6edf-4b3b-8b35-a2ac1c3f5fd0-rody_large_transactions_test","_version":3,"result":"updated","_shards":{"total":2,"successful":2,"failed":0},"_seq_no":250587,"_primary_term":2}
```

`x-elastic-client-meta`

### 해결

- [오픈서치와 ES의 클라이언트가 달라서](https://github.com/elastic/elasticsearch-py/issues/1933#issuecomment-1073830411)...인 거 같다

```log
You shouldn't use the latest version (8.1.0) of the Elasticsearch client against a 6.x version of the Elasticsearch server. Instead you should use the same major version client (6.x) against the major version of Elasticsearch that you're using.
```

- `NewTypedClient` 사용하면 안된다

## search_phase_execution_exception: all shards failed

### 문제

```log
github.com/some-qwerty-org.io/go/some-qwerty-api-service/payments.(*paymentsAggregator).SetQualifiedTransactionsAndGetDocIds
    /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/aggregator.go:69
github.com/some-qwerty-org.io/go/some-qwerty-api-service/payments.(*paymentsService).GetPayments
    /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/service.go:45
github.com/some-qwerty-org.io/go/some-qwerty-api-service.(*Application).GetPaymentsByStatus
    /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payment.go:108
github.com/gofiber/fiber/v2.(*App).next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/router.go:145
github.com/gofiber/fiber/v2.(*Ctx).Next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/ctx.go:1014
github.com/some-qwerty-org.io/go/some-qwerty-api-service.Authenticator.func1
    /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/authentication.go:47
github.com/gofiber/fiber/v2.(*App).next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/router.go:145
github.com/gofiber/fiber/v2.(*Ctx).Next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/ctx.go:1014
github.com/gofiber/fiber/v2/middleware/recover.New.func1
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/middleware/recover/recover.go:43
github.com/gofiber/fiber/v2.(*Ctx).Next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/ctx.go:1011
github.com/some-qwerty-org.io/go/util/log.New.func1
    /Users/rody/IdeaProjects/some-qwerty-org.io-go/util/log/middleware.go:131
github.com/gofiber/fiber/v2.(*Ctx).Next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/ctx.go:1011
gopkg.in/DataDog/dd-trace-go.v1/contrib/gofiber/fiber%2ev2.Middleware.func1
    /Users/rody/go/pkg/mod/gopkg.in/!data!dog/dd-trace-go.v1@v1.58.0/contrib/gofiber/fiber.v2/fiber.go:76
github.com/gofiber/fiber/v2.(*App).next
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/router.go:145
github.com/gofiber/fiber/v2.(*App).handler
    /Users/rody/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/router.go:172
github.com/valyala/fasthttp.(*Server).serveConn
    /Users/rody/go/pkg/mod/github.com/valyala/fasthttp@v1.50.0/server.go:2359
github.com/valyala/fasthttp.(*workerPool).workerFunc
    /Users/rody/go/pkg/mod/github.com/valyala/fasthttp@v1.50.0/workerpool.go:224
github.com/valyala/fasthttp.(*workerPool).getCh.func1
    /Users/rody/go/pkg/mod/github.com/valyala/fasthttp@v1.50.0/workerpool.go:196
```

### 원인

### 해결
