# Golang e2e

- [Golang e2e](#golang-e2e)
    - [iframe, postMessage, form 렌더링 등 테스트 방법](#iframe-postmessage-form-렌더링-등-테스트-방법)
        - [ChromeDP를 사용한 테스트](#chromedp를-사용한-테스트)
        - [goquery를 사용한 HTML 파싱](#goquery를-사용한-html-파싱)
        - [테스트 서버 구축](#테스트-서버-구축)
        - [Rod over ChromeDP](#rod-over-chromedp)
            - [유저 의견들(240726 기준)](#유저-의견들240726-기준)

## iframe, postMessage, form 렌더링 등 테스트 방법

Golang을 사용하여 iframe, postMessage, form 렌더링 등을 테스트하는 방법이 있습니다.
이는 주로 웹 애플리케이션의 프론트엔드 테스팅에 해당하는 영역이지만, Golang을 사용하여 이러한 테스트를 수행할 수 있습니다.

주요 원리와 방법은 다음과 같습니다:

1. 헤드리스 브라우저 사용:
   - Golang에서 헤드리스 브라우저를 제어하여 웹 페이지를 로드하고 조작할 수 있습니다.
   - 주로 사용되는 도구: ChromeDP (Chrome DevTools Protocol)

2. HTML 파싱 및 조작:
   - Golang의 HTML 파싱 라이브러리를 사용하여 DOM을 분석하고 조작할 수 있습니다.
   - 주요 라이브러리: goquery

3. 테스트 서버 구축:
   - Golang으로 테스트용 웹 서버를 만들어 실제 환경을 시뮬레이션할 수 있습니다.

4. JavaScript 실행:
   - 헤드리스 브라우저를 통해 JavaScript를 실행하고 결과를 확인할 수 있습니다.

구체적인 구현 방법:

### [ChromeDP](https://github.com/chromedp/chromedp)를 사용한 테스트

헤드리스 브라우저 시뮬레이션:
- ChromeDP는 실제 Chrome 브라우저의 기능을 프로그래밍 방식으로 제어할 수 있게 해줍니다.
- 이를 통해 JavaScript 실행, DOM 조작, 이벤트 시뮬레이션 등이 가능해집니다.
    - ChromeDP를 통해 JavaScript를 실행하고 그 결과를 Go 코드에서 확인할 수 있습니다.
    - 이를 통해 postMessage와 같은 브라우저 API의 동작을 테스트할 수 있습니다.

```go
package main

import (
    "context"
    "log"
    "github.com/chromedp/chromedp"
)

func main() {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    var res string
    err := chromedp.Run(ctx,
        chromedp.Navigate("https://example.com"),
        chromedp.Evaluate(`
            // iframe 생성
            var iframe = document.createElement('iframe');
            iframe.src = 'about:blank';
            document.body.appendChild(iframe);

            // postMessage 전송
            iframe.contentWindow.postMessage('Hello from parent', '*');

            // form 렌더링
            var form = document.createElement('form');
            form.innerHTML = '<input type="text" name="test">';
            document.body.appendChild(form);

            // 결과 반환
            "Test completed"
        `, &res),
    )

    if err != nil {
        log.Fatal(err)
    }

    log.Println(res)
}
```

각 웹 페지이의 소스 html 얻는 방법

```go
func scrapIt(url string, str *string) chromedp.Tasks { 
    return chromedp.Tasks{
         chromedp.Navigate(url), 
         chromedp.ActionFunc(func(ctx context.Context) error { 
            node, err := dom.GetDocument().Do(ctx) 
            if err != nil { 
                return err 
            } 
            
            *str, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx) 
            return err
        }), 
    } 
}
```

### goquery를 사용한 HTML 파싱

DOM 파싱 및 조작:
- goquery는 jQuery와 유사한 문법으로 HTML을 파싱하고 DOM 요소를 선택할 수 있게 해줍니다.
- 이를 통해 서버 사이드에서 HTML 구조를 분석하고 테스트할 수 있습니다.

```go
package main

import (
    "strings"
    "github.com/PuerkitoBio/goquery"
)

func TestHTMLParsing() {
    html := `
    <html>
        <body>
            <iframe src="about:blank"></iframe>
            <form>
                <input type="text" name="test">
            </form>
        </body>
    </html>
    `

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        log.Fatal(err)
    }

    // iframe 확인
    iframeCount := doc.Find("iframe").Length()
    log.Printf("Number of iframes: %d", iframeCount)

    // form 확인
    formCount := doc.Find("form").Length()
    log.Printf("Number of forms: %d", formCount)
}
```

### 테스트 서버 구축

테스트 환경 시뮬레이션:
- httptest 패키지를 사용하여 실제 서버 환경을 시뮬레이션할 수 있습니다.
- 이를 통해 네트워크 요청, 응답 처리 등을 테스트할 수 있습니다.

```go
package main

import (
    "net/http"
    "net/http/httptest"
)

func TestServer() {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(`
            <html>
                <body>
                    <iframe src="about:blank"></iframe>
                    <script>
                        window.addEventListener('message', function(event) {
                            console.log('Received message:', event.data);
                        });
                    </script>
                    <form>
                        <input type="text" name="test">
                    </form>
                </body>
            </html>
        `))
    }))
    defer ts.Close()

    // 여기서 ChromeDP나 http.Client를 사용하여 테스트 서버에 접근하고 테스트를 수행합니다.
}
```

이러한 방법들을 조합하여 복잡한 웹 애플리케이션의 다양한 측면을 Golang 환경에서 테스트할 수 있습니다. 이는 백엔드와 프론트엔드 통합 테스트, E2E 테스트 등에 유용하게 활용될 수 있습니다.

### [Rod](https://github.com/go-rod/rod) over ChromeDP

#### 유저 의견들(240726 기준)

- [reddit에 올라와 있는 의견](https://www.reddit.com/r/golang/comments/181ebuq/anybody_who_has_used_chromedp_similar_libraries/)

    1. 문서 낙후

        > 한 페이지에서 수행되는 모든 요청의 목록을 얻고 싶었습니다.
        > 크롬에서 이를 수행하는 방법을 찾을 수 없었고, 일부 문서를 찾아보니 구식이었습니다.
        > Rod는 제가 원하는 것에 대해 더 간단하고 더 나은 예제를 제공했습니다.

    2. 더 적은 리소스 사용

        > I also have used Orcgen based on Rod instead of Chromedp
        > because its uses less resources on Server than chromedp
        > and shows Similar results as far as i have experienced.

    3. ChromeDP 통한 아이프레임 작업의 한계

        > I also switched from chromedp to go-rod.
        > I ran into limitations for what I was doing specifically with iframes.
