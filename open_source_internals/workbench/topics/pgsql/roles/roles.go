// Package roles 는 T1의 판정 기준(재고 조사의 표시 파일과 구조 역할 대응표)을 atlas.Roles로 읽습니다.
//
// 표시 파일(marks_*.txt)은 형식을 설계하기 전에 다른 에이전트가 원본 경로마다 붙인 역할과 출처이고,
// 대응표(t1-structural-roles.tsv)는 그 가운데 구조 역할 경로의 목적지를 구현 전에 정해 둔 것입니다.
// 두 파일의 정본은 docs/works/evidence/learning-map-20261017/에 있고, 이 패키지는 그것을 고치지 않고 읽기만 합니다.
package roles

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
)

// evidenceDir 는 저장소 뿌리에서 판정 기준 파일이 있는 폴더까지의 경로입니다.
const evidenceDir = "docs/works/evidence/learning-map-20261017"

// markFile 은 입력 이름마다 그 입력의 표시 파일입니다.
var markFile = map[string]string{
	"e05": "marks_e05.txt", "c3": "marks_c3.txt", "c4": "marks_c4.txt", "tx": "marks_tx.txt", "e03": "marks_e03.txt",
}

// Load 함수는 입력 하나(e05·c3·c4·tx·e03)의 판정 기준을 읽습니다.
//
// 표시 파일의 모든 경로에 역할과 출처를 넣고, 대응표에 있는 경로에는 결정을 더합니다. 대응표의 출처 칸은 표시 파일의 출처이거나,
// 검토가 틀렸다고 보인 경로에서 생성기(roles.py의 PROV_FIXES)가 고친 출처라(까닭은 규칙 칸) 대응표의 것을 씁니다. 대응표에 아직 TBD가 남은
// 경로는 결정이 비어 있어, 그 키가 attrs에 있으면 구조 숨김 검사가 실패합니다(정하기 전에는 통과할 수 없음).
// 저장소 뿌리를 찾지 못하거나 파일 형식이 틀리면 오류를 돌려줍니다.
func Load(input string) (atlas.Roles, error) {
	name, ok := markFile[input]
	if !ok {
		return nil, fmt.Errorf("모르는 입력 %q", input)
	}
	dir, err := findEvidence()
	if err != nil {
		return nil, err
	}

	// 표시 파일 한 줄은 "<경로>  n=… 타입…  | 역할 | 출처 | 규칙"입니다.
	roles := atlas.Roles{}
	if err := eachLine(filepath.Join(dir, "t1-inventory", name), func(line string) error {
		if !strings.HasPrefix(line, "$") {
			return nil
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			return fmt.Errorf("표시 줄 형식: %q", line)
		}
		path := strings.Fields(parts[0])[0]
		roles[path] = atlas.Role{Role: strings.TrimSpace(parts[1]), Prov: strings.TrimSpace(parts[2])}
		return nil
	}); err != nil {
		return nil, err
	}

	// 대응표 한 줄은 "input path role prov decision rule"(탭 구분)입니다. TBD는 결정으로 받지 않습니다.
	if err := eachLine(filepath.Join(dir, "t1-structural-roles.tsv"), func(line string) error {
		f := strings.Split(line, "\t")
		if len(f) != 6 || f[0] == "input" || f[0] != input {
			return nil
		}
		r, ok := roles[f[1]]
		if !ok {
			return fmt.Errorf("대응표의 %s %s가 표시 파일에 없음", f[0], f[1])
		}
		r.Prov = f[3]
		if f[4] != "TBD" {
			r.Decision = f[4]
		}
		roles[f[1]] = r
		return nil
	}); err != nil {
		return nil, err
	}
	return roles, nil
}

// findEvidence 함수는 지금 폴더에서 위로 올라가며 판정 기준 폴더를 찾습니다. 시험은 패키지 폴더에서 돌기 때문입니다.
func findEvidence() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		cand := filepath.Join(dir, evidenceDir)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s를 찾지 못함", evidenceDir)
		}
		dir = parent
	}
}

// eachLine 함수는 파일의 줄마다 fn을 부릅니다. fn이 오류를 내면 멈추고 그 오류를 줄 번호와 함께 돌려줍니다.
func eachLine(path string, fn func(string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	n := 0
	for sc.Scan() {
		n++
		if err := fn(sc.Text()); err != nil {
			return fmt.Errorf("%s:%d: %w", path, n, err)
		}
	}
	return sc.Err()
}
