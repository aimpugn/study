#!/usr/bin/env python3
"""Build the root interview curriculum from raw source chunks.

Raw source files stay under ``interviews/source``. This builder writes the
human-facing learning documents directly under ``interviews`` so the new
curriculum is the default study surface.
"""

from __future__ import annotations

import hashlib
import importlib.util
import json
import re
import sys
from collections import Counter, defaultdict
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CURRICULUM_DIR = ROOT
SPLIT_SCRIPT = ROOT / "tools" / "split_interview_sources.py"


@dataclass(frozen=True)
class Major:
    key: str
    file_name: str
    title: str
    summary: str


@dataclass(frozen=True)
class Placement:
    major: str
    mid: str
    sub: str
    reason: str


MAJORS: list[Major] = [
    Major(
        "language-runtime",
        "language-runtime.md",
        "언어와 런타임",
        "언어 문법이 아니라 코드가 실행 단위가 되어 메모리, GC, 클래스 로딩, 런타임 스케줄러를 만나는 지점을 다룹니다.",
    ),
    Major(
        "concurrency-async-io",
        "concurrency-async-io.md",
        "동시성, 비동기, I/O",
        "스레드, 락, 대기, 이벤트 루프, 논블로킹 I/O처럼 많은 요청을 안전하고 효율적으로 처리하는 실행 모델을 다룹니다.",
    ),
    Major(
        "os-kernel-computer-architecture",
        "os-kernel-computer-architecture.md",
        "OS, 커널, 컴퓨터 구조",
        "프로세스, 커널, 부팅, CPU, 메모리, 파일 디스크립터처럼 애플리케이션 아래층의 실행 조건을 다룹니다.",
    ),
    Major(
        "network-web-protocols",
        "network-web-protocols.md",
        "네트워크와 웹 프로토콜",
        "TCP/IP, HTTP, keep-alive, gRPC, proxy, Nginx처럼 프로세스 밖으로 나간 요청이 흐르는 경로를 다룹니다.",
    ),
    Major(
        "security-cryptography",
        "security-cryptography.md",
        "보안과 암호학",
        "TLS, HTTPS, Diffie-Hellman, OAuth, token, privacy처럼 통신과 인증을 안전하게 만드는 기술 단위를 다룹니다.",
    ),
    Major(
        "database-storage-search-nosql",
        "database-storage-search-nosql.md",
        "데이터베이스, 저장소, 검색/NoSQL",
        "트랜잭션, 인덱스, 락, 복제, 파티셔닝, Elasticsearch, Couchbase처럼 데이터를 저장하고 찾는 전체 축을 함께 다룹니다.",
    ),
    Major(
        "messaging-event-driven",
        "messaging-event-driven.md",
        "메시징과 이벤트 기반 구조",
        "RabbitMQ, Kafka, AMQP, consumer, client library처럼 서비스 사이의 비동기 메시지 흐름을 다룹니다.",
    ),
    Major(
        "distributed-systems-architecture",
        "distributed-systems-architecture.md",
        "분산 시스템과 아키텍처",
        "MSA, consistency, availability, topology, saga, idempotency처럼 작은 기술 단위를 큰 시스템 설계로 조립하는 판단 축을 다룹니다.",
    ),
    Major(
        "spring-backend-frameworks",
        "spring-backend-frameworks.md",
        "Spring과 백엔드 프레임워크",
        "Spring Boot, IoC, Bean, Transaction, Servlet/Tomcat, WebClient/WebFlux처럼 백엔드 프레임워크가 런타임 위에 올리는 실행 모델을 다룹니다.",
    ),
    Major(
        "problem-solving-code-quality",
        "problem-solving-code-quality.md",
        "문제 해결, 코드 품질, 운영 실천",
        "알고리즘, 복잡도, 테스트, mock, 패턴, Docker/배포처럼 문제를 풀고 코드를 신뢰 가능하게 만드는 실천 축을 다룹니다.",
    ),
]

LEGACY_MAJOR_FILES = [
    "01-language-runtime.md",
    "02-concurrency-async-io.md",
    "03-os-kernel-computer-architecture.md",
    "04-network-web-protocols.md",
    "05-security-cryptography.md",
    "06-database-storage-search-nosql.md",
    "07-messaging-event-driven.md",
    "08-distributed-systems-architecture.md",
    "09-spring-backend-frameworks.md",
    "10-problem-solving-code-quality.md",
]

APPENDIX_KEY = "source-context"
APPENDIX_FILE = "source/_source-context-and-question-bank.md"


def load_split_module():
    spec = importlib.util.spec_from_file_location("split_interview_sources", SPLIT_SCRIPT)
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


def normalize(text: str) -> str:
    return text.casefold()


def has_any(text: str, *keywords: str) -> bool:
    haystack = normalize(text)
    return any(normalize(keyword) in haystack for keyword in keywords)


def clean_title(title: str) -> str:
    cleaned = re.sub(r"^#{1,6}\s+", "", title).strip()
    cleaned = cleaned.strip("*").strip()
    return cleaned or title


def placement_for(chunk) -> Placement:
    title = normalize(clean_title(chunk.title))
    body_head = normalize(chunk.body[:2400])
    text = title + "\n" + body_head

    if chunk.title == "[source front matter]":
        return Placement(APPENDIX_KEY, "원재료 메타데이터", "파일 앞부분과 생성 맥락", "source front matter")

    if has_any(title, "실제 동작 시나리오") and has_any(text, "wait()", "notify()", "synchronized", "모니터"):
        return Placement("concurrency-async-io", "Java monitor와 wait/notify", clean_title(chunk.title), "monitor/wait-notify scenario")

    if title in ("interview", "interviews") or has_any(title, "시나리오", "질문 목록", "프로젝트에서 사용한 오픈소스"):
        return Placement(APPENDIX_KEY, "질문 은행과 면접 시나리오", clean_title(chunk.title), "broad question-bank context")

    if has_any(text, "면접관", "지원자") and has_any(text, "질문", "답변", "시뮬레이션"):
        return Placement(APPENDIX_KEY, "질문 은행과 면접 시나리오", clean_title(chunk.title), "interview simulation")

    if has_any(title, "java 관련 질문"):
        return Placement("language-runtime", "JVM 실행 모델", clean_title(chunk.title), "java question cluster")

    if has_any(title, "docker", "container", "컨테이너", "kubernetes", "gitlab", "ci/cd", "배포"):
        return Placement("problem-solving-code-quality", "운영 실천과 배포 기반", clean_title(chunk.title), "devops/deployment practice")

    if has_any(title, "정규표현식", "algorithm", "알고리즘", "복잡도", "big o", "퀵소트", "자료 구조", "재귀", "귀납", "그래프"):
        return Placement("problem-solving-code-quality", "알고리즘과 문제 해결", clean_title(chunk.title), "algorithm/problem-solving")

    if has_any(title, "테스트", "test", "mock", "stub", "tdd", "solid", "디자인 패턴", "프록시 패턴", "proxy pattern"):
        if has_any(title, "프록시 패턴", "proxy pattern", "solid", "디자인 패턴"):
            return Placement("problem-solving-code-quality", "설계 원칙과 패턴", clean_title(chunk.title), "code design")
        return Placement("problem-solving-code-quality", "테스트와 대역 객체", clean_title(chunk.title), "testing")

    if has_any(title, "rabbitmq", "rabbit mq", "kafka", "amqp", "stomp", "mqtt", "메시지", "메세징", "amqplib"):
        if has_any(title, "amqp", "stomp", "mqtt", "프로토콜"):
            return Placement("messaging-event-driven", "메시징 프로토콜", clean_title(chunk.title), "messaging protocol")
        if has_any(title, "amqplib", "node.js 앱", "전달"):
            return Placement("messaging-event-driven", "Consumer와 클라이언트 구현", clean_title(chunk.title), "consumer/client implementation")
        return Placement("messaging-event-driven", "Broker 비교와 선택", clean_title(chunk.title), "broker comparison")

    if has_any(title, "go의 goroutine", "goroutine"):
        return Placement("concurrency-async-io", "Goroutine과 런타임 스케줄링", clean_title(chunk.title), "goroutine")

    if has_any(title, "pthread_create"):
        return Placement("concurrency-async-io", "스레드 생성과 스케줄링", clean_title(chunk.title), "thread creation")

    if has_any(title, "경쟁 조건", "race condition"):
        return Placement("concurrency-async-io", "공유 상태와 경쟁 조건", clean_title(chunk.title), "race condition")

    if has_any(title, "deadlock", "교착"):
        return Placement("concurrency-async-io", "교착과 진행 보장", clean_title(chunk.title), "deadlock")

    if has_any(title, "wait", "notify", "모니터", "synchronized", "blocked", "waiting"):
        return Placement("concurrency-async-io", "Java monitor와 wait/notify", clean_title(chunk.title), "monitor/wait-notify")

    if has_any(title, "스레드 상태", "쓰레드 상태", "thread state"):
        return Placement("concurrency-async-io", "스레드 상태와 스케줄링", clean_title(chunk.title), "thread state")

    if has_any(title, "netty", "epoll", "io_uring", "aio_", "libuv", "file descriptor", "fd"):
        if has_any(title, "netty", "event loop", "이벤트 루프"):
            return Placement("concurrency-async-io", "Event loop와 네트워크 런타임", clean_title(chunk.title), "event loop")
        return Placement("concurrency-async-io", "OS I/O multiplexing", clean_title(chunk.title), "io multiplexing")

    if has_any(title, "비동기", "async", "논블로킹", "블로킹", "i/o", "멀티플렉싱"):
        return Placement("concurrency-async-io", "Blocking, non-blocking, async 구분", clean_title(chunk.title), "async/non-blocking")

    if has_any(title, "코루틴", "coroutine"):
        return Placement("concurrency-async-io", "Coroutine과 협력형 실행", clean_title(chunk.title), "coroutine")

    if has_any(title, "서버에서 100gb 파일", "스트리밍 방식"):
        return Placement("network-web-protocols", "HTTP/gRPC와 스트리밍", clean_title(chunk.title), "streaming")

    if has_any(title, "spring jdbc", "@transactional", "transactional", "jdbc", "hikari", "리파지토리"):
        return Placement("spring-backend-frameworks", "트랜잭션과 데이터 접근", clean_title(chunk.title), "spring data access")

    if has_any(title, "apache + mod_php", "php-fpm"):
        return Placement("language-runtime", "언어 런타임 비교", clean_title(chunk.title), "runtime comparison")

    if has_any(title, "elasticsearch", "elastic", "샤드", "hot/warm", "카우치베이스", "couchbase", "nosql"):
        if has_any(title, "카우치베이스", "couchbase", "nosql"):
            return Placement("database-storage-search-nosql", "검색/NoSQL 저장소", clean_title(chunk.title), "NoSQL storage")
        return Placement("database-storage-search-nosql", "검색 엔진과 샤딩", clean_title(chunk.title), "search engine")

    if has_any(title, "acid", "mvcc", "isolation", "격리", "락", "트랜잭션", "2단계", "2pc", "2pl", "스냅숏"):
        return Placement("database-storage-search-nosql", "트랜잭션, 락, 격리", clean_title(chunk.title), "transaction/isolation")

    if has_any(title, "b tree", "b+ tree", "b+tree", "b+ 트리", "인덱스", "카디널리티", "cardinality"):
        return Placement("database-storage-search-nosql", "인덱스와 조회 성능", clean_title(chunk.title), "index/query performance")

    if has_any(title, "수억건", "3억건", "테이블 처리", "파티셔닝", "아카이빙"):
        return Placement("database-storage-search-nosql", "대용량 테이블 운영", clean_title(chunk.title), "large table operation")

    if has_any(title, "db", "database", "데이터베이스", "mysql", "postgresql", "replication", "레플리케이션", "pdo"):
        return Placement("database-storage-search-nosql", "DB 접근과 저장소 운영", clean_title(chunk.title), "database/storage")

    if has_any(title, "oauth", "pkce", "csrf", "token", "토큰", "프라이버시", "보안"):
        return Placement("security-cryptography", "인증, 토큰, 프라이버시", clean_title(chunk.title), "auth/privacy")

    if has_any(title, "tls", "https", "dhe", "ecdhe", "전방 비밀성", "암호", "crypt", "ssl", "pkcs"):
        if has_any(title, "dhe", "ecdhe", "전방 비밀성"):
            return Placement("security-cryptography", "키 교환과 전방 비밀성", clean_title(chunk.title), "key exchange")
        return Placement("security-cryptography", "TLS/HTTPS 핸드셰이크", clean_title(chunk.title), "secure transport")

    if has_any(title, "eventual consistency", "cap", "saga", "사가", "일관성", "consistency"):
        return Placement("distributed-systems-architecture", "일관성과 분산 트랜잭션", clean_title(chunk.title), "distributed consistency")

    if has_any(title, "msa", "마이크로서비스", "결제", "멱등", "idempot", "분산 트랜잭션"):
        return Placement("distributed-systems-architecture", "MSA와 결제 시스템 설계", clean_title(chunk.title), "service architecture")

    if has_any(title, "고가용성", "availability", "토폴로지", "로드 밸런싱", "확장성", "architecture", "아키텍처"):
        return Placement("distributed-systems-architecture", "가용성, 토폴로지, 확장성", clean_title(chunk.title), "availability/topology")

    if has_any(title, "resttemplate", "webclient", "webflux", "mono", "flux"):
        return Placement("spring-backend-frameworks", "Spring HTTP 클라이언트와 Reactive", clean_title(chunk.title), "spring reactive/http client")

    if has_any(title, "tomcat", "servlet", "dispatcher", "nginx", "war", "spring web"):
        return Placement("spring-backend-frameworks", "Servlet/Tomcat 요청 처리", clean_title(chunk.title), "servlet request path")

    if has_any(title, "spring", "스프링", "bean", "빈", "ioc", "dependency injection", "di", "applicationcontext", "beanfactory", "어노테이션 프로세싱", "@bean", "@component", "aop"):
        if has_any(title, "bean", "빈", "@bean", "@component"):
            return Placement("spring-backend-frameworks", "Bean 생명주기와 선택", clean_title(chunk.title), "bean lifecycle")
        if has_any(title, "ioc", "dependency injection", "di", "applicationcontext", "beanfactory"):
            return Placement("spring-backend-frameworks", "IoC와 DI", clean_title(chunk.title), "ioc/di")
        if has_any(title, "어노테이션", "aop"):
            return Placement("spring-backend-frameworks", "AOP와 어노테이션 처리", clean_title(chunk.title), "aop/annotation")
        return Placement("spring-backend-frameworks", "Spring Boot 실행 모델", clean_title(chunk.title), "spring boot")

    if has_any(title, "tcp", "ip 구조", "nat", "http", "keep-alive", "grpc", "protocol buffers", "소켓 통신", "컨제스천", "congestion"):
        if has_any(title, "http", "keep-alive", "grpc", "protocol buffers", "스트리밍", "streaming"):
            return Placement("network-web-protocols", "HTTP/gRPC와 스트리밍", clean_title(chunk.title), "web protocol")
        return Placement("network-web-protocols", "TCP/IP와 소켓 통신", clean_title(chunk.title), "tcp/ip socket")

    if has_any(title, "proxy", "reverse proxy", "nginx"):
        return Placement("network-web-protocols", "Proxy와 L7 진입 경로", clean_title(chunk.title), "proxy")

    if has_any(title, "network", "네트워크", "소켓"):
        return Placement("network-web-protocols", "TCP/IP와 소켓 통신", clean_title(chunk.title), "network")

    if has_any(title, "경쟁 조건", "race condition", "deadlock", "교착", "synchronized", "wait", "notify", "모니터", "멀티스레딩", "스레드", "쓰레드", "thread", "pthread"):
        if has_any(title, "deadlock", "교착"):
            return Placement("concurrency-async-io", "교착과 진행 보장", clean_title(chunk.title), "deadlock")
        if has_any(title, "wait", "notify", "모니터", "synchronized", "blocked", "waiting"):
            return Placement("concurrency-async-io", "Java monitor와 wait/notify", clean_title(chunk.title), "monitor/wait-notify")
        return Placement("concurrency-async-io", "공유 상태와 경쟁 조건", clean_title(chunk.title), "shared state")

    if has_any(text, "wait()", "notify()", "notifyall", "wait set", "synchronized", "모니터", "blocked", "waiting"):
        return Placement("concurrency-async-io", "Java monitor와 wait/notify", clean_title(chunk.title), "monitor/wait-notify")

    if has_any(title, "부동소수점", "cpu"):
        return Placement("os-kernel-computer-architecture", "CPU와 숫자 표현", clean_title(chunk.title), "cpu/numeric representation")

    if has_any(title, "boot", "부팅", "systemd"):
        return Placement("os-kernel-computer-architecture", "부팅과 init 시스템", clean_title(chunk.title), "boot/init")

    if has_any(title, "fork", "exec", "프로세스", "process", "background process", "`&`"):
        return Placement("os-kernel-computer-architecture", "프로세스 생성과 실행", clean_title(chunk.title), "process execution")

    if has_any(title, "운영체제", "커널", "kernel", "메모리", "memory", "파일 시스템", "선점형", "pre-emption", "서버용 컴퓨터", "노트북도 서버"):
        if has_any(title, "서버용 컴퓨터", "노트북도 서버"):
            return Placement("os-kernel-computer-architecture", "서버 하드웨어와 운영 환경", clean_title(chunk.title), "server environment")
        if has_any(title, "선점형", "pre-emption"):
            return Placement("os-kernel-computer-architecture", "스케줄링과 선점", clean_title(chunk.title), "scheduling")
        return Placement("os-kernel-computer-architecture", "커널, 메모리, 파일 시스템", clean_title(chunk.title), "kernel/memory/fs")

    if has_any(title, "jvm", "java", "classloader", "클래스 로더", "부모 위임", "메타스페이스", "jar", "java -jar", "바이트코드"):
        if has_any(title, "classloader", "클래스 로더", "부모 위임"):
            return Placement("language-runtime", "Class loading과 실행 준비", clean_title(chunk.title), "class loading")
        if has_any(title, "gc", "garbage", "가비지", "메타스페이스", "메모리"):
            return Placement("language-runtime", "GC와 메모리 관리", clean_title(chunk.title), "jvm memory/gc")
        return Placement("language-runtime", "JVM 실행 모델", clean_title(chunk.title), "jvm execution")

    if has_any(title, "kotlin", "infix", "inline", "noinline", "crossinline", "when"):
        return Placement("language-runtime", "Kotlin 언어 기능", clean_title(chunk.title), "kotlin language")

    if has_any(title, "go", "golang", "goroutine", "php", "php-fpm", "runtime", "런타임", "pdo", "node.js"):
        return Placement("language-runtime", "언어 런타임 비교", clean_title(chunk.title), "runtime comparison")

    if has_any(title, "함수형", "functional", "functor", "monad", "either", "arrow"):
        return Placement("language-runtime", "함수형 프로그래밍 모델", clean_title(chunk.title), "functional programming")

    return Placement("problem-solving-code-quality", "추가 분류 필요 항목", clean_title(chunk.title), "fallback placement")


def demote_headings(body: str) -> str:
    def replace(match: re.Match[str]) -> str:
        hashes = match.group(1)
        rest = match.group(2)
        new_level = min(6, max(5, len(hashes) + 3))
        return f"{'#' * new_level} {rest}"

    return re.sub(r"^(#{1,6})\s+(.+)$", replace, body.rstrip("\n"), flags=re.MULTILINE)


def source_ref(split_module, chunk) -> str:
    return split_module.source_ref(chunk)


def clear_generated_outputs() -> None:
    for major in MAJORS:
        (CURRICULUM_DIR / major.file_name).unlink(missing_ok=True)
    for file_name in LEGACY_MAJOR_FILES:
        (CURRICULUM_DIR / file_name).unlink(missing_ok=True)
    (CURRICULUM_DIR / "_question-index.md").unlink(missing_ok=True)
    (CURRICULUM_DIR / "_curriculum_manifest.json").unlink(missing_ok=True)
    (CURRICULUM_DIR / APPENDIX_FILE).unlink(missing_ok=True)


def write_curriculum_doc(major: Major, chunks: list, placements: dict[str, Placement], aliases_by_hash, split_module) -> None:
    path = CURRICULUM_DIR / major.file_name
    lines = [
        f"# {major.title}",
        "",
        major.summary,
        "",
        "> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.",
        "",
    ]

    mids: dict[str, dict[str, list]] = defaultdict(lambda: defaultdict(list))
    for chunk in chunks:
        placement = placements[chunk.sha256]
        mids[placement.mid][placement.sub].append(chunk)

    for mid in sorted(mids):
        lines.extend([f"## {mid}", ""])
        for sub in sorted(mids[mid]):
            lines.extend([f"### {sub}", ""])
            for chunk in mids[mid][sub]:
                placement = placements[chunk.sha256]
                aliases = aliases_by_hash[chunk.sha256]
                alias_refs = ", ".join(source_ref(split_module, alias) for alias in aliases)
                lines.extend(
                    [
                        f"#### 원문: {clean_title(chunk.title)}",
                        "",
                        f"<!-- curriculum-chunk: sha256={chunk.sha256} major={placement.major} mid={placement.mid} sub={placement.sub} sources={alias_refs} -->",
                        "",
                        f"> Source: `{source_ref(split_module, chunk)}`",
                        f"> Classification reason: {placement.reason}",
                    ]
                )
                if len(aliases) > 1:
                    lines.append(f"> Duplicate source aliases: `{alias_refs}`")
                lines.extend(["", demote_headings(chunk.body), "", "<!-- /curriculum-chunk -->", ""])

    path.write_text("\n".join(lines).rstrip() + "\n")


def write_appendix(chunks: list, placements: dict[str, Placement], aliases_by_hash, split_module) -> None:
    path = CURRICULUM_DIR / APPENDIX_FILE
    path.parent.mkdir(exist_ok=True)
    lines = [
        "# Source Context And Question Bank",
        "",
        "이 파일은 10개 기술 대분류에 바로 넣기보다 면접 맥락, 질문 목록, 시나리오로 보존해야 하는 원문을 모읍니다.",
        "정식 답변 문서로 승격할 때는 여기의 질문을 `_question-index.md`에서 적절한 대주제와 연결합니다.",
        "",
    ]
    mids: dict[str, dict[str, list]] = defaultdict(lambda: defaultdict(list))
    for chunk in chunks:
        placement = placements[chunk.sha256]
        mids[placement.mid][placement.sub].append(chunk)

    for mid in sorted(mids):
        lines.extend([f"## {mid}", ""])
        for sub in sorted(mids[mid]):
            lines.extend([f"### {sub}", ""])
            for chunk in mids[mid][sub]:
                placement = placements[chunk.sha256]
                aliases = aliases_by_hash[chunk.sha256]
                alias_refs = ", ".join(source_ref(split_module, alias) for alias in aliases)
                lines.extend(
                    [
                        f"#### 원문: {clean_title(chunk.title)}",
                        "",
                        f"<!-- curriculum-chunk: sha256={chunk.sha256} major={APPENDIX_KEY} mid={placement.mid} sub={placement.sub} sources={alias_refs} -->",
                        "",
                        f"> Source: `{source_ref(split_module, chunk)}`",
                    ]
                )
                if len(aliases) > 1:
                    lines.append(f"> Duplicate source aliases: `{alias_refs}`")
                lines.extend(["", demote_headings(chunk.body), "", "<!-- /curriculum-chunk -->", ""])

    path.write_text("\n".join(lines).rstrip() + "\n")


def write_readme(counts: Counter, appendix_count: int) -> None:
    lines = [
        "# interviews",
        "",
        "이 디렉터리는 경력 기술 인터뷰를 준비하기 위한 공간입니다.",
        "목표는 예상 질문을 많이 모으는 데서 끝나지 않고, 질문이 들어왔을 때 짧은 시간 안에 먼저 핵심을 답하고, 이어서 필요한 만큼 깊게 설명할 수 있는 상태를 만드는 것입니다.",
        "",
        "프로젝트 의도는 [PROJECT_INTENT.md](PROJECT_INTENT.md)에서 고정합니다.",
        "대표 사용 장면과 산출물 역할은 [USECASE.md](USECASE.md)에서 고정합니다.",
        "",
        "현재 학습용 기본 진입점은 이 디렉터리 바로 아래의 10개 대주제 문서입니다.",
        "원문 질문 은행과 시나리오는 [source](source/)에 보관하고, 앞으로의 정리와 학습은 아래 새 curriculum 문서를 기준으로 진행합니다.",
        "`검색/NoSQL`은 독립 문서로 분리하지 않고 `데이터베이스, 저장소, 검색/NoSQL` 안에 포함합니다.",
        "",
        "현재 파일은 아직 딥 리라이트 완료본이 아니라 `원문 배치본`입니다.",
        "다음 단계에서 각 소주제를 `짧은 직답 -> 깊은 메커니즘 -> 예시 -> 꼬리 질문 -> 검증/근거` 구조로 승격합니다.",
        "",
        "## 대분류",
        "",
    ]
    for major in MAJORS:
        lines.append(f"- [{major.title}]({major.file_name}): {counts[major.key]} source chunks")
    lines.extend(
        [
            f"- [Source Context And Question Bank]({APPENDIX_FILE}): {appendix_count} context chunks",
            "",
            "## 보존 규칙",
            "",
            "- 원본 `intervie*.md` 파일은 `source/` 아래 source reservoir로 유지합니다.",
            "- 원문 chunk는 source span, duplicate alias, SHA-256으로 추적합니다.",
            "- root curriculum 문서에서는 계층을 맞추기 위해 heading depth만 조정합니다.",
            "- `_curriculum_manifest.json`은 모든 unique chunk가 어느 대분류/중분류/소분류에 놓였는지 기록합니다.",
            "",
            "## 재생성",
            "",
            "```sh",
            "python3 interviews/tools/build_interview_curriculum.py",
            "```",
        ]
    )
    (CURRICULUM_DIR / "README.md").write_text("\n".join(lines) + "\n")


def write_question_index(unique_chunks: list, placements: dict[str, Placement], aliases_by_hash, split_module) -> None:
    lines = [
        "# Interview Question And Source Index",
        "",
        "이 파일은 원문 chunk가 최종 curriculum의 어느 위치에 들어갔는지 빠르게 찾기 위한 색인입니다.",
        "",
        "| 대분류 | 중분류 | 소분류 | 원문 제목 | Source | Aliases |",
        "|---|---|---|---|---|---|",
    ]
    major_titles = {major.key: major.title for major in MAJORS}
    major_titles[APPENDIX_KEY] = "Source Context And Question Bank"
    for chunk in unique_chunks:
        placement = placements[chunk.sha256]
        aliases = aliases_by_hash[chunk.sha256]
        alias_refs = "<br>".join(f"`{source_ref(split_module, alias)}`" for alias in aliases[1:])
        lines.append(
            "| "
            + " | ".join(
                [
                    major_titles[placement.major],
                    placement.mid,
                    placement.sub,
                    clean_title(chunk.title).replace("|", "\\|"),
                    f"`{source_ref(split_module, chunk)}`",
                    alias_refs or "",
                ]
            )
            + " |"
        )
    (CURRICULUM_DIR / "_question-index.md").write_text("\n".join(lines) + "\n")


def write_manifest(all_chunks: list, unique_chunks: list, placements: dict[str, Placement], aliases_by_hash, split_module) -> None:
    representative_by_hash = {chunk.sha256: chunk for chunk in unique_chunks}
    major_files = {major.key: major.file_name for major in MAJORS}
    major_files[APPENDIX_KEY] = APPENDIX_FILE
    manifest = {
        "source_files": split_module.SOURCE_NAMES,
        "source_file_hashes": {
            name: hashlib.sha256((split_module.SOURCE_DIR / name).read_bytes()).hexdigest()
            for name in split_module.SOURCE_NAMES
        },
        "source_chunk_count": len(all_chunks),
        "unique_chunk_count": len(unique_chunks),
        "major_count": len(MAJORS),
        "majors": [
            {
                "key": major.key,
                "file": major.file_name,
                "title": major.title,
                "summary": major.summary,
            }
            for major in MAJORS
        ],
        "appendix": {
            "key": APPENDIX_KEY,
            "file": APPENDIX_FILE,
            "title": "Source Context And Question Bank",
        },
        "chunks": [
            {
                "source_ref": source_ref(split_module, chunk),
                "representative_ref": source_ref(split_module, representative_by_hash[chunk.sha256]),
                "sha256": chunk.sha256,
                "title": chunk.title,
                "source_topic": chunk.topic,
                "is_representative": chunk is representative_by_hash[chunk.sha256],
                "curriculum": {
                    "major": placements[chunk.sha256].major,
                    "file": major_files[placements[chunk.sha256].major],
                    "mid": placements[chunk.sha256].mid,
                    "sub": placements[chunk.sha256].sub,
                    "reason": placements[chunk.sha256].reason,
                },
            }
            for chunk in all_chunks
        ],
        "deduplicated_groups": [
            {
                "sha256": digest,
                "representative_ref": source_ref(split_module, representative_by_hash[digest]),
                "source_refs": [source_ref(split_module, chunk) for chunk in aliases],
                "curriculum": {
                    "major": placements[digest].major,
                    "file": major_files[placements[digest].major],
                    "mid": placements[digest].mid,
                    "sub": placements[digest].sub,
                },
            }
            for digest, aliases in sorted(aliases_by_hash.items())
            if len(aliases) > 1
        ],
    }
    (CURRICULUM_DIR / "_curriculum_manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n"
    )


def main() -> None:
    split_module = load_split_module()
    all_chunks = []
    for name in split_module.SOURCE_NAMES:
        all_chunks.extend(split_module.read_chunks(split_module.SOURCE_DIR / name))

    aliases_by_hash = defaultdict(list)
    unique_by_hash = {}
    for chunk in all_chunks:
        aliases_by_hash[chunk.sha256].append(chunk)
        unique_by_hash.setdefault(chunk.sha256, chunk)
    unique_chunks = list(unique_by_hash.values())

    placements = {chunk.sha256: placement_for(chunk) for chunk in unique_chunks}
    chunks_by_major = defaultdict(list)
    for chunk in unique_chunks:
        chunks_by_major[placements[chunk.sha256].major].append(chunk)

    clear_generated_outputs()
    for major in MAJORS:
        write_curriculum_doc(major, chunks_by_major[major.key], placements, aliases_by_hash, split_module)
    write_appendix(chunks_by_major[APPENDIX_KEY], placements, aliases_by_hash, split_module)

    counts = Counter(placements[chunk.sha256].major for chunk in unique_chunks)
    write_readme(counts, counts[APPENDIX_KEY])
    write_question_index(unique_chunks, placements, aliases_by_hash, split_module)
    write_manifest(all_chunks, unique_chunks, placements, aliases_by_hash, split_module)

    print(f"source chunks: {len(all_chunks)}")
    print(f"unique chunks: {len(unique_chunks)}")
    print(f"major documents: {len(MAJORS)}")
    for major in MAJORS:
        print(f"{major.file_name}: {counts[major.key]}")
    print(f"{APPENDIX_FILE}: {counts[APPENDIX_KEY]}")


if __name__ == "__main__":
    main()
