package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"study/open_source_internals/workbench/internal/run"
	"study/open_source_internals/workbench/internal/server"
	"study/open_source_internals/workbench/internal/source"
	"study/open_source_internals/workbench/internal/subject"
	"study/open_source_internals/workbench/web"
)

// serve 함수는 127.0.0.1에 작업대 서버를 열고, 인터럽트 신호를 받을 때까지 요청을 처리합니다.
//
// 주소는 127.0.0.1만 받습니다. 다른 주소에 열면 같은 네트워크의 누구나 실습 컨테이너에서 SQL을 실행할 수 있기 때문입니다.
// 신호를 받으면 진행 중인 실행을 취소하고(실행기가 컨테이너 안의 gdb와 psql을 정리합니다) 20초 안에 서버를 닫습니다.
func serve(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	subjPath := fs.String("subject", "", "주제 정의 workbench.json의 경로")
	addr := fs.String("addr", "127.0.0.1:8742", "서버 주소. 호스트는 127.0.0.1만 받고, 포트 0은 빈 포트를 고릅니다")
	timeout := fs.Duration("timeout", 90*time.Second, "추적 한 번의 시간 한도. 시간 초과 처리를 실제로 시험할 때 줄입니다")
	if err := parse(fs, args, map[string]*string{"subject": subjPath}); err != nil {
		return err
	}
	if *timeout <= 0 {
		return fmt.Errorf("serve: -timeout은 0보다 커야 합니다: %s", *timeout)
	}
	if host, _, err := net.SplitHostPort(*addr); err != nil || host != "127.0.0.1" {
		return fmt.Errorf("serve: -addr의 호스트는 127.0.0.1이어야 합니다: %q", *addr)
	}

	subj, err := subject.Load(*subjPath)
	if err != nil {
		return err
	}
	// 소스는 켜져 있는 실습 컨테이너에서 읽습니다. 파일마다 새 컨테이너를 띄우는 것보다 빠르고,
	// 컨테이너의 postgres 사용자는 소스 트리에 쓸 수 없으므로(쓰기 가능한 항목 0개) 이미지와 같은 내용입니다.
	entry := &server.Entry{
		Subject: subj,
		Runner:  &run.Runner{Subject: subj, Timeout: *timeout},
		Source:  &source.Docker{Command: []string{"docker", "exec", subj.Container, "cat"}, Root: subj.SourceRoot},
	}

	// 포트 0이면 운영체제가 빈 포트를 고르므로, 먼저 연 뒤 실제 포트로 Host 검사 목록을 만듭니다.
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	s, err := server.New([]*server.Entry{entry}, port, web.FS())
	if err != nil {
		_ = ln.Close()
		return err
	}
	// 쓰기 시간 한도는 추적 시간 한도에 소스 읽기와 정리 시간 1분을 더한 값입니다.
	srv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      *timeout + time.Minute,
		IdleTimeout:       60 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	fmt.Printf("작업대: http://127.0.0.1:%d/ (끝내려면 Ctrl+C)\n", port)

	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
