package source

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// validPath 값은 받아들일 소스 상대경로의 꼴입니다. 소스 트리의 src/와 contrib/ 아래만 읽게 합니다.
var validPath = regexp.MustCompile(`^(src|contrib)/[A-Za-z0-9_][A-Za-z0-9_./-]*$`)

// ValidatePath 함수는 화면이나 기록에서 온 경로가 소스 트리 안의 상대경로인지 확인합니다.
//
// src/ 또는 contrib/로 시작하고, 영문·숫자·'_'·'.'·'/'·'-'만 쓰며, ".." 조각과 빈 조각이 없어야 합니다.
// 어기면 이유를 담은 오류를 돌려줍니다. 컨테이너 안의 다른 파일을 읽지 못하게 하는 경계입니다.
func ValidatePath(path string) error {
	if !validPath.MatchString(path) {
		return fmt.Errorf("소스 경로는 src/ 또는 contrib/로 시작하는 상대경로여야 합니다: %q", path)
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("소스 경로에 빈 조각이나 '.', '..'이 있습니다: %q", path)
		}
	}
	return nil
}

// Docker 값은 Docker 컨테이너나 이미지 안의 소스 트리에서 파일을 읽고, 읽은 파일을 기억해 둡니다.
//
// Command는 파일 경로 앞에 붙일 명령입니다. 실행 중인 컨테이너에서 읽을 때는 {"docker", "exec", "pg-lab", "cat"}이고,
// 이미지에서 바로 읽을 때는 {"docker", "run", "--rm", "--network", "none", "pg-internals:18.6", "cat"}입니다.
// Root는 컨테이너 안 소스 트리의 위치입니다(예: /src/postgresql-18.6). 여러 고루틴이 함께 써도 됩니다.
type Docker struct {
	Command []string
	Root    string
	Timeout time.Duration

	mu    sync.Mutex
	cache map[string][]string
}

// Lines 함수는 소스 상대경로 path의 파일 전체를 줄 목록으로 돌려줍니다. 결과의 0번이 1번 줄입니다.
//
// 경로가 ValidatePath를 통과하지 못하거나, docker 명령이 실패하거나, Timeout 안에 끝나지 않으면 오류를 돌려줍니다.
// 한 번 읽은 파일은 기억해 두고 다시 읽지 않습니다. 소스 트리는 이미지에 고정되어 있어 실행 중에 바뀌지 않기 때문입니다.
func (d *Docker) Lines(ctx context.Context, path string) ([]string, error) {
	if err := ValidatePath(path); err != nil {
		return nil, err
	}
	d.mu.Lock()
	if lines, ok := d.cache[path]; ok {
		d.mu.Unlock()
		return lines, nil
	}
	d.mu.Unlock()

	timeout := d.Timeout
	if timeout == 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := append(append([]string{}, d.Command[1:]...), strings.TrimSuffix(d.Root, "/")+"/"+path)
	cmd := exec.CommandContext(ctx, d.Command[0], args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("소스 파일 %s 읽기 실패: %w: %s", path, err, strings.TrimSpace(stderr.String()))
	}

	// 파일 끝의 줄바꿈 때문에 생기는 빈 마지막 조각은 줄이 아니므로 뺍니다.
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	d.mu.Lock()
	if d.cache == nil {
		d.cache = map[string][]string{}
	}
	d.cache[path] = lines
	d.mu.Unlock()
	return lines, nil
}
