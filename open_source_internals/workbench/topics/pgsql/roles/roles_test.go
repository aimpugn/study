package roles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyEvidence 함수는 판정 기준 파일(표시 파일 하나와 대응표)을 임시 폴더의 같은 상대 경로로 옮겨 쓰고, edit가 있으면 대응표 글을 바꿉니다.
// 돌려주는 것은 그 임시 저장소 뿌리입니다.
func copyEvidence(t *testing.T, mark string, edit func(string) string) string {
	t.Helper()
	src, err := findEvidence()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	dst := filepath.Join(root, evidenceDir)
	if err := os.MkdirAll(filepath.Join(dst, "t1-inventory"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{filepath.Join("t1-inventory", mark), "t1-structural-roles.tsv"} {
		b, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		text := string(b)
		if edit != nil && name == "t1-structural-roles.tsv" {
			text = edit(text)
		}
		if err := os.WriteFile(filepath.Join(dst, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestLoadTakesOnlyGeneratorFixedProvenance 함수는 대응표의 출처가 표시 파일과 다를 때, 생성기가 고친 행이면 그 출처를 쓰고 표를 손으로
// 고친 행이면 멈추는지 확인합니다(Spring 재검토 N2).
func TestLoadTakesOnlyGeneratorFixedProvenance(t *testing.T) {
	// Given: 트랜잭션 실험의 판정 기준을 그대로 옮긴 폴더
	t.Chdir(copyEvidence(t, "marks_tx.txt", nil))

	// When: 읽으면
	rs, err := Load("tx")

	// Then: 생성기가 고친 읽기 구간의 출처는 도출이고, 고치지 않은 경로는 표시 파일의 출처다
	if err != nil {
		t.Fatal(err)
	}
	if p := rs["$.captures[].reads.pages"].Prov; p != "도출" {
		t.Fatalf("고친 읽기 구간의 출처 %q, 기대 도출", p)
	}
	if p := rs["$.steps[]"].Prov; p != "관찰" {
		t.Fatalf("단계의 출처 %q, 기대 관찰", p)
	}

	// Given: 생성기를 거치지 않고 단계 행의 출처 칸만 도출로 바꾼 대응표
	t.Chdir(copyEvidence(t, "marks_tx.txt", func(s string) string {
		lines := strings.Split(s, "\n")
		for i, l := range lines {
			if strings.HasPrefix(l, "tx\t$.steps[]\t") {
				lines[i] = strings.Replace(l, "\t관찰\t", "\t도출\t", 1)
			}
		}
		return strings.Join(lines, "\n")
	}))

	// When/Then: 읽으면 멈춘다
	if _, err := Load("tx"); err == nil || !strings.Contains(err.Error(), "고침 표시가 없음") {
		t.Fatalf("손으로 고친 출처를 받으면 안 됨: %v", err)
	}
}
