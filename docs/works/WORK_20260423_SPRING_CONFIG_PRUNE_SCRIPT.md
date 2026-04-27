# WORK_20260423_SPRING_CONFIG_PRUNE_SCRIPT

## 0. Meta

- 작업 제목: Spring 설정 민감정보 prune 스크립트
- WORK 파일 경로: `WORK_20260423_SPRING_CONFIG_PRUNE_SCRIPT.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute`
- 작업 깊이: `standard`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: Spring `application.yml`, `application.yaml`, `application.properties` 계열에서 회사 IP, 비밀번호, API key, token, profile 전용 설정을 안전하게 제거하거나 치환하는 재사용 스크립트를 만든다.
- scope: `jvm/spring/tools/prune_spring_config.py`, `jvm/spring/tools/test_prune_spring_config.py`
- mode: `execute`
- run_mode: `normal`
- finish: `test+commit`
- must_keep: `common`, `junit`, `local` profile은 보존한다. `kcf-outbound-gateway-{dev,real,stg}.properties`처럼 파일명만으로 제거 대상인 파일은 내용을 읽지 않는다.
- extra_checks: 기본 실행은 `dry-run`이어야 하며, 원본 secret 값을 stdout에 출력하지 않아야 한다.

## 2. Frozen Checklist

- [ ] profile suffix가 `dev`, `real`, `stg`인 제거 대상 파일은 읽지 않고 제거 대상으로 보고한다.
- [ ] `application-{profile}` 파일은 `common`, `junit`, `local` 외 profile이면 제거 대상으로 본다.
- [ ] YAML multi-document에서 `spring.config.activate.on-profile` 또는 `spring.profiles`가 허용 profile 밖이면 해당 document를 제거한다.
- [ ] properties multi-document에서 `spring.config.activate.on-profile` 또는 `spring.profiles`가 허용 profile 밖이면 해당 document를 제거한다.
- [ ] 보존 파일 안의 password, username, token, key, apik류 값과 IPv4 값은 치환한다.
- [ ] 기본 실행은 dry-run이고 `--write`에서만 파일을 변경한다.
- [ ] build output 디렉터리(`target`, `build` 등)는 기본 scan에서 제외한다.
- [ ] 테스트가 사용자 예시의 properties/YAML 형태와 파일 제거 요구를 포함한다.

## 3. Evidence / Decision Ledger

- E-01: 사용자 예시는 properties와 YAML 양쪽에서 URL 내부 IP, password, token, key, `kmsg-sndr-sys-apik`, `server-ip`, `sftp-ip`가 섞이는 형태다. 따라서 단순 key 삭제가 아니라 value 내부 IPv4 치환과 key 기반 redaction이 필요하다.
- E-02: 사용자가 `common`, `junit`, `local` 외 profile 제거를 명시했다. 따라서 profile file name과 Spring multi-document profile selector가 모두 처리 대상이다.
- D-01: 기본을 dry-run으로 둔다. 실제 삭제와 치환은 `--write`에서만 적용한다. 이유는 secret prune 도구가 잘못 작동하면 설정 손실이 크기 때문이다.
- D-02: 제거 대상 profile suffix 파일은 내용을 읽지 않는다. 이유는 사용자가 직접 "읽지도 말고 제거"를 요구했고, 읽기 실패 또는 secret 출력 위험을 줄일 수 있기 때문이다.

## 4. Verification Plan

- `python3 -m unittest jvm/spring/tools/test_prune_spring_config.py`
- `python3 jvm/spring/tools/prune_spring_config.py --help`
- `python3 jvm/spring/tools/prune_spring_config.py /tmp/...` 형태의 테스트 fixture dry-run 확인

## 5. Final Re-Judgement

- [x] profile suffix가 `dev`, `real`, `stg`인 제거 대상 파일은 읽지 않고 제거 대상으로 보고한다.
- [x] `application-{profile}` 파일은 `common`, `junit`, `local` 외 profile이면 제거 대상으로 본다.
- [x] YAML multi-document에서 `spring.config.activate.on-profile` 또는 `spring.profiles`가 허용 profile 밖이면 해당 document를 제거한다.
- [x] properties multi-document에서 `spring.config.activate.on-profile` 또는 `spring.profiles`가 허용 profile 밖이면 해당 document를 제거한다.
- [x] 보존 파일 안의 password, username, token, key, apik류 값과 IPv4 값은 치환한다.
- [x] 기본 실행은 dry-run이고 `--write`에서만 파일을 변경한다.
- [x] build output 디렉터리(`target`, `build` 등)는 기본 scan에서 제외한다.
- [x] 테스트가 사용자 예시의 properties/YAML 형태와 파일 제거 요구를 포함한다.

## 6. Verification Result

- `python3 -m unittest jvm/spring/tools/test_prune_spring_config.py`: PASS, 7 tests.
- `python3 -m py_compile jvm/spring/tools/prune_spring_config.py jvm/spring/tools/test_prune_spring_config.py`: PASS.
- `python3 -m unittest discover -s jvm/spring/tools -p 'test_*.py'`: PASS, 7 tests.
- `python3 jvm/spring/tools/prune_spring_config.py .`: PASS. 현재 저장소 기준 dry-run은 source application config 2개만 대상으로 잡고, build output은 제외하며, 원본 secret 값을 출력하지 않는다.

## 7. Closure

- requested closure scope: Spring 설정 prune 스크립트와 검증 테스트 추가.
- achieved closure scope: requested scope 완료.
- remaining open items: 없음.
- unrelated open work: 기존 dirty worktree의 다른 문서/예제 변경은 이번 작업 범위 밖이며 건드리지 않았다.
- verdict: `WHOLE_COMPLETE`.
