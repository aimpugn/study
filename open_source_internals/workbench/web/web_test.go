package web

import (
	"bytes"
	"strings"
	"testing"
)

// 재생본이 외부 파일을 부르지 않고, 소스 줄의 "</script>"가 스크립트를 일찍 닫지 않는지 확인합니다.
func TestStandaloneInlinesEverythingAndEscapesScriptClose(t *testing.T) {
	// Given: 소스 줄에 스크립트를 닫는 글자가 든 자료
	bundle := map[string]string{"line": `printf("</script><b>")`}

	// When
	page, err := Standalone(bundle)
	// Then
	if err != nil {
		t.Fatal(err)
	}
	// 브라우저가 무언가를 내려받게 만드는 꼴만 찾습니다. SVG 이름공간 문자열("http://www.w3.org/2000/svg")은 요청이 아닙니다.
	for _, external := range []string{`href="app.css"`, `src="app.js"`, `src="http`, `href="http`, `url(http`, `@import`, "fetch('http"} {
		if strings.Contains(string(page), external) {
			t.Fatalf("재생본이 외부 자원을 부릅니다: %s", external)
		}
	}
	if bytes.Count(page, []byte("</script>")) != 2 {
		t.Fatalf("</script>는 자료 스크립트와 화면 스크립트를 닫는 두 번만 나와야 합니다: %d번", bytes.Count(page, []byte("</script>")))
	}
	if !bytes.Contains(page, []byte(`\u003c/script\u003e`)) {
		t.Fatal("자료 안의 </script>는 \\u003c/script\\u003e로 바뀌어야 합니다")
	}
}
