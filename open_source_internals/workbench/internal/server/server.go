// Package server 는 작업대 화면 파일과 실행 API를 제공하는 HTTP 처리기입니다.
//
// 이 패키지가 맡는 보안 경계는 다른 사이트가 학습자의 브라우저를 통해 보낸 요청을 막는 것입니다.
// 실습의 psql 사용자는 슈퍼유저라서 SQL을 걸러서는 컨테이너 안의 명령 실행을 막을 수 없고,
// 그 피해의 범위는 컨테이너 격리(--network none, 읽기 전용 마운트)가 정합니다(설계 기록의 안전 경계 절).
// 따라서 여기서는 "이 요청이 정말 이 서버의 화면에서 왔는가"만 엄격하게 확인합니다.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"study/open_source_internals/workbench/internal/replay"
	"study/open_source_internals/workbench/internal/run"
	"study/open_source_internals/workbench/internal/subject"
)

// maxRequestBytes 값은 실행 요청 본문의 최대 크기입니다. SQL 한도(run.MaxSQLBytes)에 JSON 꾸밈을 더해도 넉넉합니다.
const maxRequestBytes = 8 << 10

// contentSecurityPolicy 값은 화면이 이 서버의 파일만 불러오고, 다른 사이트의 틀(iframe) 안에 들어가지 못하게 합니다.
// 틀 안에 넣지 못하게 하는 것은, 다른 사이트가 작업대를 투명하게 덮어 학습자가 모르게 실행 단추를 누르게 하는 일을 막기 위해서입니다.
const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self'; " +
	"base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// Entry 값은 서버가 다루는 주제 하나와, 그 주제의 실행기와 소스 읽기입니다.
type Entry struct {
	Subject subject.Subject
	Runner  *run.Runner
	Source  replay.LineReader
}

// Server 값은 작업대의 HTTP 처리기입니다. New로 만듭니다.
type Server struct {
	entries map[string]*Entry
	order   []string
	hosts   map[string]bool
	static  fs.FS
	now     func() time.Time
}

// New 함수는 port에 열린 작업대 서버를 만듭니다. 처리기는 Handler로 꺼냅니다.
//
// 인자: entries는 다룰 주제들이고 화면은 첫 주제를 씁니다. port는 서버가 실제로 연 포트로, Host와 Origin 검사에 씁니다.
// static은 화면 파일(index.html, app.css, app.js)이 든 파일 시스템입니다.
// 실패: 주제가 없거나 주제 id가 겹치면 오류를 돌려줍니다.
func New(entries []*Entry, port int, static fs.FS) (*Server, error) {
	if len(entries) == 0 {
		return nil, errors.New("주제가 하나도 없습니다")
	}
	s := &Server{
		entries: map[string]*Entry{},
		hosts: map[string]bool{
			fmt.Sprintf("127.0.0.1:%d", port): true,
			fmt.Sprintf("localhost:%d", port): true,
		},
		static: static,
		now:    time.Now,
	}
	for _, e := range entries {
		if _, dup := s.entries[e.Subject.ID]; dup {
			return nil, fmt.Errorf("주제 id %q가 겹칩니다", e.Subject.ID)
		}
		s.entries[e.Subject.ID] = e
		s.order = append(s.order, e.Subject.ID)
	}
	return s, nil
}

// Handler 함수는 모든 요청에 출처 검사(guard)를 한 뒤 화면 파일, GET /api/subjects, POST /api/runs를 제공하는 처리기를 돌려줍니다.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(s.static))
	mux.HandleFunc("GET /api/subjects", s.subjects)
	mux.HandleFunc("POST /api/runs", s.runs)
	return s.guard(mux)
}

// guard 함수는 모든 요청에 출처 검사를 하고, 통과한 요청만 next로 넘깁니다.
//
// 모든 요청: Host가 127.0.0.1:포트 또는 localhost:포트여야 합니다. 다른 이름이 이 주소를 가리키게 바꾸는
// DNS 재바인딩으로 다른 사이트가 같은 출처처럼 행세하는 일을 막습니다.
// API 요청과 GET·HEAD가 아닌 요청: X-Workbench: 1 머리글이 있어야 하고, Origin이 있으면 이 서버의 주소와 같아야 하며,
// Sec-Fetch-Site가 있으면 same-origin이어야 합니다. 다른 사이트의 스크립트는 브라우저의 사전 확인(preflight) 없이
// 이런 머리글을 붙일 수 없고, 이 서버는 어떤 응답에도 CORS 허용 머리글을 붙이지 않으므로 사전 확인은 언제나 실패합니다.
func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("Referrer-Policy", "no-referrer")

		if !s.hosts[r.Host] {
			fail(w, http.StatusForbidden, fmt.Sprintf("Host %q는 이 서버의 주소가 아닙니다", r.Host))
			return
		}

		// 경로는 정리한 뒤에 봅니다. "/x/../api/runs"처럼 돌아 들어오는 경로도 API 요청으로 다루기 위해서입니다.
		p := path.Clean("/" + r.URL.Path)
		isAPI := p == "/api" || strings.HasPrefix(p, "/api/")
		if isAPI || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
			h.Set("Cache-Control", "no-store")
			if r.Header.Get("X-Workbench") != "1" {
				fail(w, http.StatusForbidden, "X-Workbench 머리글이 없습니다. 작업대 화면에서 보낸 요청만 받습니다")
				return
			}
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
				fail(w, http.StatusForbidden, fmt.Sprintf("Origin %q는 이 서버가 아닙니다", o))
				return
			}
			if f := r.Header.Get("Sec-Fetch-Site"); f != "" && f != "same-origin" {
				fail(w, http.StatusForbidden, fmt.Sprintf("Sec-Fetch-Site %q 요청은 받지 않습니다", f))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// watchView 값은 화면에 보내는 감시 목록입니다. 감시 목록 파일의 경로는 보내지 않습니다.
type watchView struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Examples []string `json:"examples"`
}

// subjectView 값은 화면에 보내는 주제 하나의 정보와 실습 컨테이너의 상태입니다.
type subjectView struct {
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Container    string      `json:"container"`
	Running      bool        `json:"running"`
	StartCommand string      `json:"startCommand"`
	SourceNotice string      `json:"sourceNotice"`
	Watches      []watchView `json:"watches"`
}

// subjects 함수는 주제 목록과 각 실습 컨테이너가 지금 켜져 있는지를 돌려줍니다.
func (s *Server) subjects(w http.ResponseWriter, r *http.Request) {
	views := make([]subjectView, 0, len(s.order))
	for _, id := range s.order {
		e := s.entries[id]
		v := subjectView{
			ID: e.Subject.ID, Title: e.Subject.Title, Container: e.Subject.Container,
			Running: e.Runner.Running(r.Context()), StartCommand: e.Subject.StartCommand, SourceNotice: e.Subject.SourceNotice,
		}
		for _, wt := range e.Subject.Watches {
			v.Watches = append(v.Watches, watchView{ID: wt.ID, Label: wt.Label, Examples: wt.Examples})
		}
		views = append(views, v)
	}
	writeJSON(w, http.StatusOK, map[string]any{"subjects": views})
}

// runRequest 값은 POST /api/runs의 본문입니다.
type runRequest struct {
	Subject string `json:"subject"`
	Watch   string `json:"watch"`
	SQL     string `json:"sql"`
}

// runResponse 값은 실행 한 번의 결과입니다. 화면은 재생본과 같은 bundle을 그대로 그립니다.
type runResponse struct {
	Bundle   replay.Bundle `json:"bundle"`
	Warnings []string      `json:"warnings"`
}

// runs 함수는 SQL 한 문장을 추적하고, 그 기록을 화면이 그릴 bundle로 돌려줍니다.
//
// 상태 코드: 요청 꼴이 틀리면 400, JSON이 아니면 415, 이 서버나 컨테이너 안에서 다른 추적이 진행 중이면 409, 컨테이너가 꺼져 있으면 503,
// 시간 한도를 넘으면 504, 그 밖의 실패는 500입니다. 본문은 언제나 {"error": 이유} 또는 runResponse입니다.
func (s *Server) runs(w http.ResponseWriter, r *http.Request) {
	// 본문은 JSON 값 하나여야 하고, 모르는 필드가 있으면 거부합니다. 오타가 난 필드가 조용히 빈 값이 되는 일을 막습니다.
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		fail(w, http.StatusUnsupportedMediaType, "본문은 application/json이어야 합니다")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req runRequest
	if err := dec.Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, fmt.Sprintf("요청 본문을 읽지 못했습니다: %v", err))
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		fail(w, http.StatusBadRequest, "요청 본문에는 JSON 값이 하나만 있어야 합니다")
		return
	}
	e, ok := s.entries[req.Subject]
	if !ok {
		fail(w, http.StatusBadRequest, fmt.Sprintf("주제 %q가 없습니다", req.Subject))
		return
	}

	// 실행이 끝나면 기록을 재생본과 같은 bundle로 바꿔, 실시간 화면과 재생본이 같은 그리기 코드를 쓰게 합니다.
	res, err := e.Runner.Run(r.Context(), req.Watch, req.SQL)
	if err != nil {
		fail(w, statusOf(err), err.Error())
		return
	}
	watch, _ := e.Subject.Watch(req.Watch)
	meta := replay.Meta{
		SQL: req.SQL, WatchID: watch.ID, WatchLabel: watch.Label,
		Origin: "실시간 실행", RecordedAt: s.now().Format("2006-01-02 15:04:05"),
	}
	bundle, err := replay.Build(r.Context(), res.Record, e.Subject, e.Source, meta)
	if err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	warnings := res.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	writeJSON(w, http.StatusOK, runResponse{Bundle: bundle, Warnings: warnings})
}

// statusOf 함수는 run.Runner의 오류를 HTTP 상태 코드로 바꿉니다.
func statusOf(err error) int {
	var ie *run.InputError
	switch {
	case errors.As(err, &ie):
		return http.StatusBadRequest
	case errors.Is(err, run.ErrBusy), errors.Is(err, run.ErrTraceInProgress):
		return http.StatusConflict
	case errors.Is(err, run.ErrNotRunning):
		return http.StatusServiceUnavailable
	case errors.Is(err, run.ErrTimeout):
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

// fail 함수는 {"error": reason} 본문으로 실패를 응답합니다.
func fail(w http.ResponseWriter, status int, reason string) {
	writeJSON(w, status, map[string]string{"error": reason})
}

// writeJSON 함수는 v를 JSON으로 응답합니다. 이미 머리글을 보낸 뒤라 실패를 알릴 방법이 없으므로 쓰기 오류는 버립니다.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
