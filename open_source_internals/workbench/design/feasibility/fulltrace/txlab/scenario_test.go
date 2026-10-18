package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseScenarioKeepsPauseOutOfSteps 는 쉼 줄이 단계로 세지지 않고, 쉼 뒤에 올 단계의 순번에 붙는지 봅니다.
// 해석기와 화면은 단계마다 세션과 SQL이 있다고 보므로, 쉼이 단계에 섞이면 단계 번호가 기록과 어긋납니다.
func TestParseScenarioKeepsPauseOutOfSteps(t *testing.T) {
	// Given: 두 번째 단계 앞에 쉼 1.5초, 마지막에 남는 쉼은 없음
	f := t.TempDir() + "/s.txt"
	text := "# 제목\nwatch: acct\nS1: BEGIN;\npause: 1500ms\nS2: SELECT 1;\n"
	if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	// When
	sc, err := parseScenario(f)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Steps) != 2 || sc.Steps[1].Session != "S2" {
		t.Fatalf("steps = %+v; want S1, S2 두 단계", sc.Steps)
	}
	if len(sc.Pauses) != 1 || sc.Pauses[1] != 1500*time.Millisecond {
		t.Fatalf("pauses = %v; want map[1:1.5s]", sc.Pauses)
	}
}

// TestParseScenarioRejectsBadPause 는 단위가 없거나 0인 쉼을 조용히 넘기지 않는지 봅니다(인자를 직접 풀다 잘못된 값이
// 통과해 측정이 틀린 조건으로 돌았던 일이 있어, 읽지 못한 값은 오류로 멈춥니다).
func TestParseScenarioRejectsBadPause(t *testing.T) {
	for _, bad := range []string{"pause: 1500", "pause: 0s", "pause: -1s"} {
		// Given
		f := t.TempDir() + "/s.txt"
		if err := os.WriteFile(f, []byte("S1: BEGIN;\n"+bad+"\nS1: COMMIT;\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// When
		_, err := parseScenario(f)

		// Then
		if err == nil {
			t.Fatalf("%q를 받아들였습니다", bad)
		}
	}
}

// TestParseScenarioKeepsKillOutOfSteps 는 kill 줄이 단계로 세지지 않고, 그 뒤에 올 단계의 순번에 붙는지, 같은 자리의 쉼과
// 함께 놓일 수 있는지, 세션마다 처음 쓰인 구간이 남는지 봅니다. 실험 3-crash-sync-off-wait가 "쉼 11초 → kill → 단계" 꼴입니다.
func TestParseScenarioKeepsKillOutOfSteps(t *testing.T) {
	// Given: S1의 단계 하나, 쉼, S1을 죽임, 재시작 뒤 처음 쓰는 S2의 단계 둘
	f := t.TempDir() + "/s.txt"
	text := "S1: INSERT INTO acct VALUES (2, 'lee', 200);\npause: 11s\nkill: S1\nS2: SELECT * FROM acct;\nS2: SELECT 1;\n"
	if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	// When
	sc, err := parseScenario(f)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Steps) != 3 {
		t.Fatalf("steps = %+v; want 단계 셋(kill은 단계가 아님)", sc.Steps)
	}
	if len(sc.Kills) != 1 || sc.Kills[1] != "S1" {
		t.Fatalf("kills = %v; want map[1:S1]", sc.Kills)
	}
	if sc.Pauses[1] != 11*time.Second {
		t.Fatalf("pauses = %v; want map[1:11s]", sc.Pauses)
	}
	if sc.Epoch["S1"] != 0 || sc.Epoch["S2"] != 1 {
		t.Fatalf("epoch = %v; want S1 0, S2 1", sc.Epoch)
	}
}

// TestParseScenarioRejectsBadKill 는 실행하면 조용히 틀어질 kill 꼴을 파일을 읽을 때 멈추는지 봅니다.
func TestParseScenarioRejectsBadKill(t *testing.T) {
	for _, c := range []struct{ why, text string }{
		{"쓰인 적 없는 세션", "S1: BEGIN;\nkill: S9\nS2: SELECT 1;\n"},
		{"이미 죽인 세션을 다시 죽임", "S1: BEGIN;\nkill: S1\nkill: S1\nS2: SELECT 1;\n"},
		{"kill 앞의 세션을 그 뒤에 씀", "S1: BEGIN;\nkill: S1\nS1: SELECT 1;\n"},
		{"죽이지 않은 세션도 재시작 뒤에는 끊김", "S1: BEGIN;\nS2: BEGIN;\nkill: S1\nS2: SELECT 1;\n"},
		{"kill 뒤 같은 단계 앞의 쉼", "S1: BEGIN;\nkill: S1\npause: 1s\nS2: SELECT 1;\n"},
	} {
		// Given
		f := t.TempDir() + "/s.txt"
		if err := os.WriteFile(f, []byte(c.text), 0o644); err != nil {
			t.Fatal(err)
		}

		// When
		_, err := parseScenario(f)

		// Then
		if err == nil {
			t.Errorf("%s: 받아들였습니다", c.why)
		}
	}
}

// TestParseScenarioReadsSpringScenario 는 Spring 시나리오(실험 5)가 세션 줄 없이 받아들여지고, 풀 크기를 읽거나 기본값 2를 갖는지,
// 세션 줄이 섞이거나 풀 크기를 읽지 못하면 오류로 멈추는지 봅니다. 섞인 줄을 조용히 받으면 psql 단계와 앱의 멈춤 자리가 한 기록에
// 겹쳐 단계 번호가 무엇을 가리키는지 정해지지 않습니다.
func TestParseScenarioReadsSpringScenario(t *testing.T) {
	for _, c := range []struct {
		name, text string
		pool       int
		blocked    time.Duration
		ok         bool
	}{
		{"풀 크기와 기다림 있음", "# 전파\nsetup: CREATE TABLE acct(id int);\nwatch: acct\nspring: requires-new\npool: 1\nblocked: 500ms\n", 1, 500 * time.Millisecond, true},
		{"풀 크기와 기다림 기본값", "spring: required\n", 2, 1500 * time.Millisecond, true},
		{"세션 줄이 섞임", "spring: required\nS1: SELECT 1;\n", 0, 0, false},
		{"풀 크기가 숫자가 아님", "spring: required\npool: two\n", 0, 0, false},
		{"풀 크기가 0", "spring: required\npool: 0\n", 0, 0, false},
		{"기다림이 시간이 아님", "spring: required\nblocked: soon\n", 0, 0, false},
		{"Spring이 아닌데 기다림이 있음", "S1: SELECT 1;\nblocked: 500ms\n", 0, 0, false},
	} {
		// Given
		f := t.TempDir() + "/s.txt"
		if err := os.WriteFile(f, []byte(c.text), 0o644); err != nil {
			t.Fatal(err)
		}

		// When
		sc, err := parseScenario(f)

		// Then
		if (err == nil) != c.ok {
			t.Fatalf("%s: 오류 %v; want ok=%v", c.name, err, c.ok)
		}
		if c.ok && (sc.Spring == "" || sc.Pool != c.pool || sc.Blocked != c.blocked) {
			t.Fatalf("%s: Spring %q, Pool %d, Blocked %v; want 장면 이름과 풀 %d, 기다림 %v", c.name, sc.Spring, sc.Pool, sc.Blocked, c.pool, c.blocked)
		}
	}
}

// TestScenarioFilesParse 는 저장소의 시나리오 파일이 모두 읽히는지 봅니다. 기록은 맥에서 시작해 컨테이너 안 serve까지 간 뒤에야
// 시나리오를 다시 읽으므로, 형식이 틀린 파일은 서버를 한 번 띄운 뒤에야 드러납니다.
func TestScenarioFilesParse(t *testing.T) {
	// Given
	files, err := filepath.Glob("../../../../../database/postgresql/transactions/*.txt")
	if err != nil || len(files) == 0 {
		t.Fatalf("시나리오 파일을 찾지 못했습니다: %v", err)
	}
	for _, f := range files {
		// When
		sc, err := parseScenario(f)

		// Then
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(f), err)
			continue
		}
		if strings.HasPrefix(filepath.Base(f), "5-") && (sc.Spring == "" || len(sc.Watch) == 0) {
			t.Errorf("%s: 실험 5 시나리오인데 Spring %q, watch %v", filepath.Base(f), sc.Spring, sc.Watch)
		}
	}
}
