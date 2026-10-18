package roles

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// generatorInputs 는 대응표를 만들 때 생성기에 주는 입력의 차례입니다. 표의 행 차례가 이 차례를 따릅니다(roles.py의 Usage, 메모 6).
var generatorInputs = []string{"e05", "c3", "c4", "tx", "e03"}

// generatorDiff 함수는 저장소 뿌리 root의 표시 파일로 생성기(roles.py)를 다시 돌린 출력이 커밋된 대응표와 바이트까지 같은지 봅니다.
// 다르거나 생성기가 실패하면 오류를 돌려줍니다. python3가 없으면 이 판정을 할 수 없어 오류입니다.
func generatorDiff(root string) error {
	dir := filepath.Join(root, evidenceDir)
	args := []string{"-I", "roles.py"}
	for _, in := range generatorInputs {
		args = append(args, markFile[in], in)
	}
	cmd := exec.Command("python3", args...)
	cmd.Dir = filepath.Join(dir, "t1-inventory")
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("생성기 실패: %v: %s", err, stderr.String())
	}
	table, err := os.ReadFile(filepath.Join(dir, "t1-structural-roles.tsv"))
	if err != nil {
		return err
	}
	if !bytes.Equal(out.Bytes(), table) {
		return fmt.Errorf("커밋된 대응표가 생성기 출력과 다름(생성기 %d바이트, 표 %d바이트)", out.Len(), len(table))
	}
	return nil
}

// TestTableIsTheGeneratorOutput 함수는 커밋된 대응표가 생성기의 출력 그대로인지 확인합니다. 로더는 규칙 칸의 고침 표시만 보므로,
// 표시 글까지 손으로 넣은 표는 로더를 통과합니다. 그런 표가 검사 4의 기대를 바꾸지 못하게 하는 것이 이 시험입니다(Spring 바깥 검토 지적 2).
func TestTableIsTheGeneratorOutput(t *testing.T) {
	// Given/When/Then: 저장소의 판정 기준은 생성기 출력과 같다
	src, err := findEvidence()
	if err != nil {
		t.Fatal(err)
	}
	repo := strings.TrimSuffix(src, string(filepath.Separator)+filepath.FromSlash(evidenceDir))
	if err := generatorDiff(repo); err != nil {
		t.Fatal(err)
	}

	// Given: 생성기와 표시 파일을 옮기고, 단계 행의 출처를 도출로 바꾸며 고침 표시 글까지 손으로 넣은 대응표
	root := t.TempDir()
	dst := filepath.Join(root, evidenceDir)
	if err := os.MkdirAll(filepath.Join(dst, "t1-inventory"), 0o755); err != nil {
		t.Fatal(err)
	}
	names := []string{filepath.Join("t1-inventory", "roles.py"), "t1-structural-roles.tsv"}
	for _, in := range generatorInputs {
		names = append(names, filepath.Join("t1-inventory", markFile[in]))
	}
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		text := string(b)
		if name == "t1-structural-roles.tsv" {
			lines := strings.Split(text, "\n")
			for i, l := range lines {
				if strings.HasPrefix(l, "tx\t$.steps[]\t") {
					lines[i] = strings.Replace(l, "\t관찰\t", "\t도출\t", 1) + "; 출처 고침(표시 관찰 → 도출)"
				}
			}
			text = strings.Join(lines, "\n")
		}
		if err := os.WriteFile(filepath.Join(dst, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// When: 로더로 읽으면
	t.Chdir(root)
	rs, err := Load("tx")

	// Then: 로더는 받지만(규칙 칸만 보므로) 생성기 출력과 견주면 다르다
	if err != nil || rs["$.steps[]"].Prov != "도출" {
		t.Fatalf("이 변이는 로더를 통과하는 꼴이어야 함: %v, %q", err, rs["$.steps[]"].Prov)
	}
	if err := generatorDiff(root); err == nil || !strings.Contains(err.Error(), "생성기 출력과 다름") {
		t.Fatalf("손으로 고친 표를 잡아야 함: %v", err)
	}
}

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
