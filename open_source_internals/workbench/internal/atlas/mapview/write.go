package mapview

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// screenFiles 는 화면 파일입니다. index.html 하나에 map.css·range.js·map.js와 보기 모형을 넣어 파일 하나로 씁니다(file://로 열림).
//
//go:embed screen/index.html screen/map.css screen/range.js screen/map.js
var screenFiles embed.FS

// WriteHTML 함수는 보기 모형을 품은 정적 화면 파일 하나를 씁니다. 화면은 공급자를 부르지 않습니다(정적 재생본, 설계 6).
//
// 화면 파일의 자리 표시(/*CSS*/, /*RANGE*/, /*MAP*/, /*DATA*/)를 채웁니다. 자리 표시가 없거나 쓰기가 실패하면 오류를 돌려줍니다.
func WriteHTML(w io.Writer, v *View) error {
	read := func(name string) (string, error) {
		b, err := screenFiles.ReadFile("screen/" + name)
		return string(b), err
	}
	page, err := read("index.html")
	if err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// 보기 모형 안의 글이 스크립트를 닫지 못하게 "</"를 바꿉니다.
	safe := strings.ReplaceAll(string(data), "</", "<\\/")
	for _, part := range []struct{ mark, file, text string }{{"/*CSS*/", "map.css", ""}, {"/*RANGE*/", "range.js", ""}, {"/*MAP*/", "map.js", ""}, {"/*DATA*/", "", safe}} {
		if !strings.Contains(page, part.mark) {
			return fmt.Errorf("화면 파일에 자리 표시 %s가 없음", part.mark)
		}
		text := part.text
		if part.file != "" {
			if text, err = read(part.file); err != nil {
				return err
			}
		}
		page = strings.Replace(page, part.mark, text, 1)
	}
	_, err = io.Copy(w, bytes.NewBufferString(page))
	return err
}
