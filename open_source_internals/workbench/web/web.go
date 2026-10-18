// Package web 은 작업대의 화면 파일(index.html, app.css, app.js)을 실행 파일 안에 담고, 재생본 HTML 하나로 묶습니다.
//
// 실시간 서버와 재생본이 같은 화면 파일을 씁니다. 화면을 고치면 두 경로가 함께 바뀌도록, 화면 파일의 정본은 이 폴더 하나입니다.
package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

// files 값은 화면 파일입니다. 서버는 이것을 그대로 내주고, Standalone은 이것을 한 파일로 묶습니다.
//
//go:embed index.html app.css app.js
var files embed.FS

// FS 함수는 화면 파일을 담은 파일 시스템을 돌려줍니다. 서버가 http.FileServer로 내줄 때 씁니다.
func FS() fs.FS {
	return files
}

// Standalone 함수는 자료 bundle을 담은 재생본 HTML 한 파일을 만듭니다.
//
// index.html의 app.css 링크와 app.js 스크립트를 파일 내용으로 바꾸고, 그 앞에 자료를 JSON 스크립트로 넣습니다.
// 그래서 결과 파일은 서버도 네트워크도 없이 열립니다. 바꿀 자리가 index.html에 정확히 한 번씩 있지 않으면 오류를 돌려줍니다.
// 화면 파일을 고치다 자리를 지웠는데 자료 없는 파일이 조용히 만들어지는 것을 막기 위해서입니다.
func Standalone(bundle any) ([]byte, error) {
	page, err := files.ReadFile("index.html")
	if err != nil {
		return nil, err
	}
	css, err := files.ReadFile("app.css")
	if err != nil {
		return nil, err
	}
	js, err := files.ReadFile("app.js")
	if err != nil {
		return nil, err
	}

	// json.Marshal은 <, >, &를 < 꼴로 바꾸므로, 소스 줄에 "</script>"가 있어도 스크립트 태그가 일찍 닫히지 않습니다.
	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("재생 자료를 JSON으로 바꾸기: %w", err)
	}

	style := []byte(`<link rel="stylesheet" href="app.css">`)
	script := []byte(`<script src="app.js"></script>`)
	for _, marker := range [][]byte{style, script} {
		if n := bytes.Count(page, marker); n != 1 {
			return nil, fmt.Errorf("index.html에 %s가 %d번 있습니다. 정확히 한 번이어야 합니다", marker, n)
		}
	}
	page = bytes.Replace(page, style, concat("<style>\n", css, "</style>"), 1)
	page = bytes.Replace(page, script, concat(`<script id="bundle" type="application/json">`, data, "</script>\n<script>\n", js, "</script>"), 1)
	return page, nil
}

// concat 함수는 문자열과 바이트 조각을 차례로 이어 붙입니다.
func concat(parts ...any) []byte {
	var b bytes.Buffer
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			b.WriteString(v)
		case []byte:
			b.Write(v)
		}
	}
	return b.Bytes()
}
