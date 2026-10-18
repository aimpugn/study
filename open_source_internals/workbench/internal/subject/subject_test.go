package subject

import (
	"os"
	"path/filepath"
	"testing"
)

// 실제 PostgreSQL 주제 정의를 읽고, 파일 하나에 붙인 층 이름이 디렉터리 전체의 이름보다 앞서는지 확인합니다.
func TestLoadRealSubjectAndClassify(t *testing.T) {
	// Given
	s, err := Load("../../../database/postgresql/workbench.json")
	if err != nil {
		t.Fatal(err)
	}

	// When / Then: printtup.c는 파일 접두사, heaptuple.c는 디렉터리 접두사, 모르는 경로는 분류 없음
	cases := map[string]string{
		"src/backend/access/common/printtup.c":  "결과 전송",
		"src/backend/access/common/heaptuple.c": "공통 행 처리",
		"src/backend/tcop/postgres.c":           "명령 처리",
		"src/include/access/tableam.h":          "테이블 접근",
		"src/backend/utils/mmgr/aset.c":         Unclassified,
	}
	for path, want := range cases {
		if got := s.Classify(path); got != want {
			t.Errorf("%s의 층 %q, %q여야 합니다", path, got, want)
		}
	}
	if _, ok := s.Watch("select"); !ok {
		t.Fatal("select 감시 목록이 있어야 합니다")
	}
}

// 같은 이름의 층은 세로줄 하나를 함께 쓰고, 분류 없음 줄이 맨 끝에 오는지 확인합니다.
func TestLanesDeduplicateLabels(t *testing.T) {
	// Given
	s := Subject{Layers: []Layer{{"a/", "가"}, {"b/", "나"}, {"c/", "가"}}}

	// When
	lanes := s.Lanes()

	// Then
	if len(lanes) != 3 || lanes[0].Label != "가" || lanes[1].Label != "나" || lanes[2].Label != Unclassified {
		t.Fatalf("세로줄 %+v, 가·나·분류 없음이어야 합니다", lanes)
	}
}

// 필수 값이 비었거나 감시 목록 파일이 없는 정의를 거부하는지 확인합니다.
func TestLoadRejectsIncompleteSubject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "run.sh"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"container 없음":     `{"id":"x","image":"i","sourceRoot":"/s","runScript":"run.sh","watches":[{"id":"w","file":"run.sh"}]}`,
		"감시 목록 없음":         `{"id":"x","container":"c","image":"i","sourceRoot":"/s","runScript":"run.sh"}`,
		"감시 목록 파일 없음":      `{"id":"x","container":"c","image":"i","sourceRoot":"/s","runScript":"run.sh","watches":[{"id":"w","file":"none.watch"}]}`,
		"감시 목록 id 겹침":      `{"id":"x","container":"c","image":"i","sourceRoot":"/s","runScript":"run.sh","watches":[{"id":"w","file":"run.sh"},{"id":"w","file":"run.sh"}]}`,
		"JSON 형식 아님":       `{"id":`,
		"pgOptions에 셸 문자":  `{"id":"x","container":"c","image":"i","sourceRoot":"/s","runScript":"run.sh","watches":[{"id":"w","file":"run.sh","pgOptions":"-c a=b; rm -rf /"}]}`,
		"pgOptions에 -c 빠짐": `{"id":"x","container":"c","image":"i","sourceRoot":"/s","runScript":"run.sh","watches":[{"id":"w","file":"run.sh","pgOptions":"enable_seqscan=off"}]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			// Given
			path := filepath.Join(dir, "workbench.json")
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}

			// When
			_, err := Load(path)

			// Then
			if err == nil {
				t.Fatalf("오류여야 합니다: %s", body)
			}
		})
	}
}
