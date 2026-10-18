package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"study/open_source_internals/workbench/internal/run"
	"study/open_source_internals/workbench/internal/subject"
)

// port 값은 시험용 서버가 열렸다고 가정하는 포트입니다. Host와 Origin 검사가 이 값을 씁니다.
const port = 8742

// sampleTrace 값은 run-trace.sh 출력의 최소 꼴입니다.
const sampleTrace = `STOP 1 exec_simple_query src/backend/tcop/postgres.c:1014 query_string=0xaaaa "SELECT 1;"
STOP 2 ReadyForQuery src/backend/tcop/dest.c:258
END ReadyForQuery=1 exec_simple_query=1
--- psql
1
1
`

// fakeDocker 값은 시험용 명령 실행기입니다. docker inspect에는 running을, 추적 스크립트에는 sampleTrace를 돌려줍니다.
// block이 있으면 추적 스크립트는 block이 닫힐 때까지 끝나지 않습니다.
type fakeDocker struct {
	running bool
	block   chan struct{}

	mu      sync.Mutex
	scripts int
}

// exec 함수는 run.Executor 꼴의 가짜 실행입니다.
func (f *fakeDocker) exec(ctx context.Context, _ []string, name string, args ...string) ([]byte, []byte, error) {
	switch {
	case name == "docker" && args[0] == "inspect":
		return []byte(fmt.Sprint(f.running)), nil, nil
	case name == "docker" && strings.Contains(strings.Join(args, " "), "echo busy"):
		return []byte("free"), nil, nil
	case name == "sh":
		f.mu.Lock()
		f.scripts++
		f.mu.Unlock()
		if f.block != nil {
			select {
			case <-f.block:
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
		}
		return []byte(sampleTrace), nil, nil
	}
	return nil, nil, nil
}

// started 함수는 추적 스크립트를 부른 횟수를 돌려줍니다.
func (f *fakeDocker) started() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.scripts
}

// fakeSource 값은 어떤 파일이든 "줄 N" 꼴의 2000줄을 돌려주는 소스 읽기입니다.
type fakeSource struct{}

// Lines 함수는 replay.LineReader를 채웁니다.
func (fakeSource) Lines(_ context.Context, _ string) ([]string, error) {
	lines := make([]string, 2000)
	for i := range lines {
		lines[i] = fmt.Sprintf("줄 %d", i+1)
	}
	return lines, nil
}

// newServer 함수는 실제 PostgreSQL 주제 정의와 가짜 실행기로 서버 처리기를 만듭니다.
func newServer(t *testing.T, f *fakeDocker) http.Handler {
	t.Helper()
	subj, err := subject.Load("../../../database/postgresql/workbench.json")
	if err != nil {
		t.Fatal(err)
	}
	static := fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>작업대</title>")}}
	s, err := New([]*Entry{{Subject: subj, Runner: &run.Runner{Subject: subj, Exec: f.exec, Timeout: 2 * time.Second}, Source: fakeSource{}}}, port, static)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 10, 12, 23, 0, 0, 0, time.UTC) }
	return s.Handler()
}

// request 값은 시험 요청 하나의 꼴입니다. 비어 있는 머리글은 보내지 않습니다.
type request struct {
	method, path, host, body, contentType string
	workbench, origin, fetchSite          string
}

// send 함수는 req를 처리기에 보내고 응답을 돌려줍니다.
func send(h http.Handler, req request) *httptest.ResponseRecorder {
	r := httptest.NewRequest(req.method, req.path, strings.NewReader(req.body))
	r.Host = req.host
	for k, v := range map[string]string{"X-Workbench": req.workbench, "Origin": req.origin, "Sec-Fetch-Site": req.fetchSite, "Content-Type": req.contentType} {
		if v != "" {
			r.Header.Set(k, v)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// sameOrigin 함수는 작업대 화면이 실제로 보내는 꼴의 실행 요청을 만듭니다.
func sameOrigin(body string) request {
	return request{
		method: http.MethodPost, path: "/api/runs", host: fmt.Sprintf("127.0.0.1:%d", port), body: body,
		contentType: "application/json", workbench: "1", origin: fmt.Sprintf("http://127.0.0.1:%d", port), fetchSite: "same-origin",
	}
}

// 안전 경계의 각 항목을 어기는 요청이 실제로 거부되고, 화면이 보내는 요청만 통과하는지 확인합니다(판정 기준 B5).
func TestGuardRejectsRequestsNotFromWorkbenchPage(t *testing.T) {
	good := sameOrigin(`{"subject":"postgresql","watch":"select","sql":"SELECT 1;"}`)
	with := func(change func(*request)) request { r := good; change(&r); return r }
	cases := []struct {
		name string
		req  request
		want int
	}{
		{"화면이 보내는 요청", good, http.StatusOK},
		{"localhost 이름의 같은 출처", with(func(r *request) {
			r.host, r.origin = fmt.Sprintf("localhost:%d", port), fmt.Sprintf("http://localhost:%d", port)
		}), http.StatusOK},
		{"Origin 없는 같은 출처(일부 브라우저)", with(func(r *request) { r.origin = "" }), http.StatusOK},
		{"DNS 재바인딩: 다른 Host", with(func(r *request) { r.host = fmt.Sprintf("evil.example:%d", port) }), http.StatusForbidden},
		{"다른 포트의 Host", with(func(r *request) { r.host = "127.0.0.1:9999" }), http.StatusForbidden},
		{"X-Workbench 없음", with(func(r *request) { r.workbench = "" }), http.StatusForbidden},
		{"X-Workbench 값이 다름", with(func(r *request) { r.workbench = "true" }), http.StatusForbidden},
		{"다른 사이트의 Origin", with(func(r *request) { r.origin = "http://evil.example" }), http.StatusForbidden},
		{"이름만 다른 Origin(127.0.0.1 화면에서 localhost로)", with(func(r *request) {
			r.origin = fmt.Sprintf("http://localhost:%d", port)
		}), http.StatusForbidden},
		{"Origin 값 null", with(func(r *request) { r.origin = "null" }), http.StatusForbidden},
		{"Sec-Fetch-Site cross-site", with(func(r *request) { r.fetchSite = "cross-site" }), http.StatusForbidden},
		{"Sec-Fetch-Site same-site", with(func(r *request) { r.fetchSite = "same-site" }), http.StatusForbidden},
		{"Sec-Fetch-Site none", with(func(r *request) { r.fetchSite = "none" }), http.StatusForbidden},
		{"..로 돌아 들어오는 경로, 머리글 없음", with(func(r *request) { r.path, r.workbench = "/x/../api/runs", "" }), http.StatusForbidden},
		// GET은 메서드 규칙에 걸리지 않으므로, 이 줄은 출처 검사가 정리된 경로를 보는지만 시험합니다.
		{"..로 돌아 들어오는 GET 경로, 머리글 없음", request{method: http.MethodGet, path: "/x/../api/subjects", host: good.host}, http.StatusForbidden},
		{"GET 아닌 화면 파일 요청, 머리글 없음", with(func(r *request) { r.path, r.workbench = "/index.html", "" }), http.StatusForbidden},
		{"주제 목록, 머리글 없음", request{method: http.MethodGet, path: "/api/subjects", host: good.host}, http.StatusForbidden},
		{"화면 파일, 다른 Host", request{method: http.MethodGet, path: "/", host: "evil.example"}, http.StatusForbidden},
		{"화면 파일, 주소창에서 연 요청", request{method: http.MethodGet, path: "/", host: good.host, fetchSite: "none"}, http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Given
			h := newServer(t, &fakeDocker{running: true})

			// When
			w := send(h, c.req)

			// Then
			if w.Code != c.want {
				t.Fatalf("상태 %d, 기대 %d: %s", w.Code, c.want, w.Body.String())
			}
			for k := range w.Header() {
				if strings.HasPrefix(k, "Access-Control-") {
					t.Fatalf("CORS 머리글 %s를 붙였습니다", k)
				}
			}
			if w.Header().Get("X-Frame-Options") != "DENY" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
				t.Fatal("틀 안에 넣지 못하게 하는 머리글이 없습니다")
			}
		})
	}
}

// 실행 요청의 꼴이 틀리거나 실행할 수 없을 때, 원인에 맞는 상태 코드와 이유를 돌려주는지 확인합니다.
func TestRunsRejectsMalformedRequests(t *testing.T) {
	big := `{"subject":"postgresql","watch":"select","sql":"` + strings.Repeat("a", maxRequestBytes) + `"}`
	cases := []struct {
		name, body, contentType string
		running                 bool
		want                    int
		reason                  string
	}{
		{"백슬래시 메타 명령", `{"subject":"postgresql","watch":"select","sql":"SELECT 1; \\! id"}`, "", true, http.StatusBadRequest, "백슬래시"},
		{"두 문장", `{"subject":"postgresql","watch":"select","sql":"SELECT 1; SELECT 2;"}`, "", true, http.StatusBadRequest, "한 문장"},
		{"설정에 없는 감시 목록(경로 시도)", `{"subject":"postgresql","watch":"../../etc/passwd","sql":"SELECT 1;"}`, "", true, http.StatusBadRequest, "감시 목록"},
		{"없는 주제", `{"subject":"linux","watch":"select","sql":"SELECT 1;"}`, "", true, http.StatusBadRequest, "주제"},
		{"모르는 필드", `{"subject":"postgresql","watch":"select","sql":"SELECT 1;","stackAt":"x"}`, "", true, http.StatusBadRequest, "unknown field"},
		{"JSON 값 둘", `{"subject":"postgresql","watch":"select","sql":"SELECT 1;"}{}`, "", true, http.StatusBadRequest, "하나만"},
		{"너무 큰 본문", big, "", true, http.StatusBadRequest, "too large"},
		{"JSON이 아닌 본문", `subject=postgresql`, "application/x-www-form-urlencoded", true, http.StatusUnsupportedMediaType, "application/json"},
		{"컨테이너 꺼짐", `{"subject":"postgresql","watch":"select","sql":"SELECT 1;"}`, "", false, http.StatusServiceUnavailable, "켜져 있지 않습니다"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Given
			f := &fakeDocker{running: c.running}
			h := newServer(t, f)
			req := sameOrigin(c.body)
			if c.contentType != "" {
				req.contentType = c.contentType
			}

			// When
			w := send(h, req)

			// Then
			var body struct{ Error string }
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if w.Code != c.want || !strings.Contains(body.Error, c.reason) {
				t.Fatalf("상태 %d(기대 %d), 이유 %q에 %q가 없습니다", w.Code, c.want, body.Error, c.reason)
			}
			if f.started() != 0 {
				t.Fatal("거부해야 할 요청에 추적 스크립트를 불렀습니다")
			}
		})
	}
}

// 실행 중에 온 두 번째 실행 요청을 409로 거부하는지 확인합니다.
func TestRunsRejectsConcurrentRunWith409(t *testing.T) {
	// Given: 첫 실행이 block이 닫힐 때까지 끝나지 않음
	f := &fakeDocker{running: true, block: make(chan struct{})}
	h := newServer(t, f)
	body := `{"subject":"postgresql","watch":"select","sql":"SELECT 1;"}`
	first := make(chan int, 1)
	go func() { first <- send(h, sameOrigin(body)).Code }()
	deadline := time.Now().Add(2 * time.Second)
	for f.started() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("첫 실행이 2초 안에 시작하지 않았습니다")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// When
	second := send(h, sameOrigin(body)).Code
	close(f.block)

	// Then
	if second != http.StatusConflict || <-first != http.StatusOK {
		t.Fatalf("두 번째는 409, 첫째는 200이어야 합니다: %d", second)
	}
}

// 정상 실행이 재생본과 같은 꼴의 bundle과 빈 주의 목록을 돌려주는지 확인합니다.
func TestRunsReturnsBundleLikeReplay(t *testing.T) {
	// Given
	h := newServer(t, &fakeDocker{running: true})

	// When
	w := send(h, sameOrigin(`{"subject":"postgresql","watch":"select","sql":"SELECT 1;"}`))

	// Then
	if w.Code != http.StatusOK {
		t.Fatalf("상태 %d: %s", w.Code, w.Body.String())
	}
	var got struct {
		Bundle struct {
			Steps []struct{ Seq int }
			Meta  struct{ SQL, Origin, RecordedAt, Guide string }
			Files map[string]map[string]string
		}
		Warnings []string
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	m := got.Bundle.Meta
	if len(got.Bundle.Steps) != 2 || m.SQL != "SELECT 1;" || m.Origin != "실시간 실행" || m.RecordedAt != "2026-10-12 23:00:00" || m.Guide != "" {
		t.Fatalf("bundle이 다릅니다: %+v", got.Bundle)
	}
	if got.Warnings == nil || len(got.Bundle.Files["src/backend/tcop/postgres.c"]) == 0 {
		t.Fatalf("주의 목록은 빈 배열이고, 멈춘 파일의 줄이 들어 있어야 합니다: %v", got.Warnings)
	}
	if !strings.Contains(w.Body.String(), `"warnings":[]`) {
		t.Fatal("주의가 없으면 null이 아니라 빈 배열이어야 합니다")
	}
}

// 주제 목록이 컨테이너 상태와 화면에 필요한 값만 보내고, 감시 목록 파일 경로는 보내지 않는지 확인합니다.
func TestSubjectsReportsRunningAndHidesWatchFiles(t *testing.T) {
	// Given
	h := newServer(t, &fakeDocker{running: false})

	// When
	w := send(h, request{method: http.MethodGet, path: "/api/subjects", host: fmt.Sprintf("127.0.0.1:%d", port), workbench: "1", fetchSite: "same-origin"})

	// Then
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), ".watch") {
		t.Fatalf("상태 %d, 본문에 감시 목록 파일 경로가 없어야 합니다: %s", w.Code, w.Body.String())
	}
	var got struct {
		Subjects []struct {
			ID      string
			Running bool
			Watches []struct{ ID string }
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want, err := subject.Load("../../../database/postgresql/workbench.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Subjects) != 1 || got.Subjects[0].Running || len(got.Subjects[0].Watches) != len(want.Watches) {
		t.Fatalf("주제 목록이 다릅니다. 감시 목록은 workbench.json의 %d개여야 합니다: %+v", len(want.Watches), got)
	}
	for i, w := range want.Watches {
		if got.Subjects[0].Watches[i].ID != w.ID {
			t.Fatalf("%d번 감시 목록 %q, %q여야 합니다", i, got.Subjects[0].Watches[i].ID, w.ID)
		}
	}
}

// 정해 둔 경로가 아닌 메서드는 405로 거부하는지 확인합니다.
func TestOtherMethodsAreNotAllowed(t *testing.T) {
	// Given
	h := newServer(t, &fakeDocker{running: true})
	req := sameOrigin("")
	req.method, req.path = http.MethodPut, "/api/subjects"

	// When
	w := send(h, req)

	// Then
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("상태 %d, 기대 405", w.Code)
	}
}

// 실행기의 오류마다 화면이 원인을 구분할 수 있는 상태 코드로 바뀌는지 확인합니다. 감싼 오류도 같은 코드여야 합니다.
func TestStatusOfMapsRunnerErrors(t *testing.T) {
	cases := map[error]int{
		&run.InputError{Reason: "x"}: http.StatusBadRequest,
		run.ErrBusy:                  http.StatusConflict,
		run.ErrTraceInProgress:       http.StatusConflict,
		fmt.Errorf("감쌈: %w", run.ErrTraceInProgress): http.StatusConflict,
		run.ErrNotRunning: http.StatusServiceUnavailable,
		run.ErrTimeout:    http.StatusGatewayTimeout,
		fmt.Errorf("추적 스크립트가 실패했습니다"): http.StatusInternalServerError,
	}
	for err, want := range cases {
		if got := statusOf(err); got != want {
			t.Errorf("%v: 상태 %d, 기대 %d", err, got, want)
		}
	}
}
