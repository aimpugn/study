package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// 맥 쪽 자리입니다. 기록 빌드는 컨테이너의 /src에서 맥 공유 경로로 한 번 옮겨 두고, 컨테이너는 그것을 /lab으로 봅니다.
const (
	container  = "pg-lab"
	machine    = "tracelab"
	srcBin     = "/src/postgresql-18.6/src/backend/postgres"
	tracefs    = "/sys/kernel/tracing"
	instance   = tracefs + "/instances/txlab"
	inLabBin   = "/lab/.lab-tmp/postgres"
	inLabTxlab = "/tmp/txlab"
)

// cmdRecord 는 맥에서 시나리오 하나를 기록합니다.
//
// 순서: 디스크 여유 확인 → 자기 자신을 linux/arm64로 빌드해 컨테이너에 넣음 → 컨테이너 안 serve(기록 서버) →
// tracelab에 uprobe 정의를 넣고 txlab 인스턴스에서 켬 → 컨테이너 안 run(시나리오) → uprobe 기록을 끄고 읽음 →
// 컨테이너 안 stop → 기록 파일을 <out>으로 옮기고 컨테이너의 함수 기록을 지움. tracelab 설정은 끝에서(실패해도) 되돌립니다.
//
// Errors: 어느 단계든 실패하면 그 단계 이름을 붙여 돌려줍니다. ftrace 버퍼가 넘쳐 사건을 잃었으면 기록을 남긴 채 오류로 끝냅니다.
func cmdRecord(args []string) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	scPath := fs.String("scenario", "", "시나리오 파일")
	out := fs.String("out", "", "출력 디렉터리(맥)")
	probes := fs.String("probes", "probes.txt", "관측 정의 파일")
	labBin := fs.String("labbin", "../../../../../database/postgresql/scripts/.lab-tmp/postgres", "맥에서 본 기록 빌드(컨테이너의 /lab/.lab-tmp/postgres와 같은 파일)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	sc, err := parseScenario(*scPath)
	if err != nil {
		return err
	}
	bin, err := filepath.Abs(*labBin)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(*out, "trace"), 0o755); err != nil {
		return err
	}
	if err := checkDisk(); err != nil {
		return err
	}
	if err := syncLabBinary(bin); err != nil {
		return fmt.Errorf("기록 빌드 옮기기: %w", err)
	}
	defs, err := parseProbes(*probes)
	if err != nil {
		return err
	}
	lines, err := uprobeLines(bin, bin, defs)
	if err != nil {
		return fmt.Errorf("uprobe 정의: %w", err)
	}

	// 컨테이너 안에서 돌 같은 프로그램을 빌드해 넣습니다. 시나리오도 함께 넣습니다.
	tmp, err := os.MkdirTemp("", "txlab")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	build := exec.Command("go", "build", "-o", filepath.Join(tmp, "txlab"), ".")
	build.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm64", "CGO_ENABLED=0")
	if b, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("linux 빌드: %v\n%s", err, b)
	}
	for _, cp := range [][2]string{{filepath.Join(tmp, "txlab"), inLabTxlab}, {*scPath, "/tmp/txlab-scenario.txt"}} {
		if err := run("docker", "cp", cp[0], container+":"+cp[1]); err != nil {
			return err
		}
	}
	if err := run("docker", "exec", container, inLabTxlab, "serve", "-bin", inLabBin, "-scenario", "/tmp/txlab-scenario.txt"); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	// uprobe는 txlab 무리(group)로만 넣고 txlab 인스턴스에서만 켜서, 최상위 추적기와 다른 실험의 설정을 건드리지 않습니다.
	defer restoreMachine(defs)
	var setup strings.Builder
	setup.WriteString(fmt.Sprintf("set -e; T=%s; mkdir -p %s; echo 0 > %s/tracing_on; echo 32768 > %s/buffer_size_kb; echo 0 > %s/options/overwrite; echo tai > %s/trace_clock\n",
		tracefs, instance, instance, instance, instance, instance))
	for _, l := range lines {
		setup.WriteString(fmt.Sprintf("echo '%s' >> $T/uprobe_events\n", l))
	}
	setup.WriteString(fmt.Sprintf("echo 1 > %s/events/txlab/enable; echo > %s/trace; echo 1 > %s/tracing_on\n", instance, instance, instance))
	if err := machineSh(setup.String()); err != nil {
		return fmt.Errorf("uprobe 켜기: %w", err)
	}

	// 실험 1~4는 컨테이너 안 run이 psql 세션을 띄우고, 실험 5는 맥의 Java 앱이 bridge를 거쳐 붙습니다(spring.go).
	var runErr error
	if sc.Spring != "" {
		runErr = springRun(sc, *out, filepath.Dir(bin))
	} else {
		runErr = run("docker", "exec", container, inLabTxlab, "run", "-scenario", "/tmp/txlab-scenario.txt", "-out", labDir+"/out")
	}

	if err := machineSh(fmt.Sprintf("echo 0 > %s/tracing_on", instance)); err != nil {
		return err
	}
	trace, err := machineOut(fmt.Sprintf("cat %s/trace", instance))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "uprobe.trace"), trace, 0o644); err != nil {
		return err
	}
	stats, err := machineOut(fmt.Sprintf("for c in %s/per_cpu/cpu*; do echo $c $(grep -E '^(overrun|dropped events)' $c/stats | tr '\\n' ' '); done", instance))
	if err != nil {
		return err
	}
	if err := run("docker", "exec", container, inLabTxlab, "stop"); err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	for _, cp := range [][2]string{{labDir + "/out/steps.json", "steps.json"}, {labDir + "/all.strace", "all.strace"}, {labDir + "/server.log", "server.log"}} {
		if err := run("docker", "cp", container+":"+cp[0], filepath.Join(*out, cp[1])); err != nil && runErr == nil {
			return err
		}
	}
	if sc.Spring != "" {
		if err := copyWireLogs(*out); err != nil && runErr == nil {
			return err
		}
	}
	if err := run("docker", "cp", container+":"+traceDir+"/.", filepath.Join(*out, "trace")); err != nil {
		return err
	}
	if err := run("docker", "exec", container, "rm", "-rf", traceDir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "ftrace-stats.txt"), stats, 0o644); err != nil {
		return err
	}
	if runErr != nil {
		return fmt.Errorf("run: %w", runErr)
	}
	for _, l := range strings.Split(strings.TrimSpace(string(stats)), "\n") {
		f := strings.Fields(l)
		for i := 0; i+1 < len(f); i++ {
			if (f[i] == "overrun:" || f[i] == "events:") && f[i+1] != "0" {
				return fmt.Errorf("ftrace 버퍼가 넘쳐 사건을 잃었습니다: %s", l)
			}
		}
	}
	fmt.Println("기록:", *out)
	return nil
}

// copyWireLogs 는 컨테이너 쪽 relay가 남긴 메시지 기록(/tmp/pgi/out/wire-*.jsonl)을 <out>으로 옮깁니다.
func copyWireLogs(out string) error {
	b, err := exec.Command("docker", "exec", container, "sh", "-c", "ls "+labDir+"/out").Output()
	if err != nil {
		return err
	}
	for _, f := range strings.Fields(string(b)) {
		if strings.HasPrefix(f, "wire-") && strings.HasSuffix(f, ".jsonl") {
			if err := run("docker", "cp", container+":"+labDir+"/out/"+f, filepath.Join(out, f)); err != nil {
				return err
			}
		}
	}
	return nil
}

// restoreMachine 은 txlab 인스턴스를 지우고 넣었던 uprobe 정의를 뺍니다. 실패하면 알리기만 합니다(기록은 이미 남김).
func restoreMachine(defs []probeDef) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("T=%s; [ -d %s ] && { echo 0 > %s/tracing_on; echo 0 > %s/events/txlab/enable 2>/dev/null; rmdir %s; }\n", tracefs, instance, instance, instance, instance))
	for _, d := range defs {
		b.WriteString(fmt.Sprintf("echo '-:txlab/%s' >> $T/uprobe_events 2>/dev/null\n", d.Name))
	}
	b.WriteString("grep -c txlab/ $T/uprobe_events || true\n")
	out, err := machineOut(b.String())
	if err != nil || strings.TrimSpace(lastLine(out)) != "0" {
		fmt.Fprintf(os.Stderr, "tracelab 되돌리기를 확인하지 못했습니다: %v %s\n", err, out)
	}
}

func lastLine(b []byte) string {
	l := strings.Split(strings.TrimSpace(string(b)), "\n")
	return l[len(l)-1]
}

// syncLabBinary 는 컨테이너의 기록 빌드와 맥 공유 경로의 사본 크기가 다르면 다시 옮깁니다.
func syncLabBinary(bin string) error {
	b, err := exec.Command("docker", "exec", container, "stat", "-c", "%s", srcBin).Output()
	if err != nil {
		return err
	}
	want, _ := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if fi, err := os.Stat(bin); err == nil && fi.Size() == want {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		return err
	}
	return run("docker", "cp", container+":"+srcBin, bin)
}

// checkDisk 는 맥 데이터 볼륨의 여유가 5 GB 이상인지 봅니다. 이 맥은 디스크가 차면 OrbStack VM이 멈춰 다른 컨테이너까지 멈춥니다.
func checkDisk() error {
	out, err := exec.Command("df", "-k", "/System/Volumes/Data").Output()
	if err != nil {
		return err
	}
	l := strings.Split(strings.TrimSpace(string(out)), "\n")
	f := strings.Fields(l[len(l)-1])
	if len(f) < 4 {
		return fmt.Errorf("df 출력을 읽지 못했습니다: %s", out)
	}
	kb, _ := strconv.ParseInt(f[3], 10, 64)
	// 2026-10-14에 기록과 무관한 설치가 20분 사이 15 GB를 써서 여유가 0이 되었고, OrbStack VM이 멈췄습니다.
	// 기록 한 벌은 지우기 전 최대 약 0.5 GB를 쓰지만, 다른 프로그램이 쓰는 몫까지 남기려고 10 GB를 요구합니다.
	if kb < 10<<20 {
		return fmt.Errorf("맥 디스크 여유가 %d MB라 기록하지 않습니다(10 GB 이상 필요)", kb>>10)
	}
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func machineSh(script string) error {
	_, err := machineOut(script)
	return err
}

// machineOut 은 tracelab에서 root로 sh 스크립트를 실행하고 표준 출력을 돌려줍니다.
func machineOut(script string) ([]byte, error) {
	cmd := exec.Command("orb", "-m", machine, "-u", "root", "sh", "-c", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%v: %s", err, stderr.String())
	}
	return out, nil
}
