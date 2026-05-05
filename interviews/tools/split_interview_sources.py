#!/usr/bin/env python3
"""Split raw interview markdown files into source-preserving topic files.

This generator deliberately does not rewrite or improve the source chunks.
It copies each unique chunk verbatim, records every original source span in a
manifest, and lets later deep-study passes work from traceable topic files.
"""

from __future__ import annotations

import hashlib
import json
import re
from collections import defaultdict
from dataclasses import dataclass, asdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOPIC_DIR = ROOT / "topics"

SOURCE_NAMES = [
    "interview_questions.md",
    "interview_questions2.md",
    "interview_questions3.md",
    "interview_questions4.md",
    "interview_s4.md",
    "interviews.md",
    "interviews2.md",
]

TOPICS = {
    "00-source-front-matter-and-question-banks": "Source Front Matter And Question Banks",
    "algorithms-and-data-structures": "Algorithms And Data Structures",
    "containers-and-devops": "Containers And DevOps",
    "concurrency-async-io": "Concurrency Async IO",
    "database-and-storage": "Database And Storage",
    "distributed-systems-and-architecture": "Distributed Systems And Architecture",
    "functional-programming": "Functional Programming",
    "jvm-java-kotlin-runtime": "JVM Java Kotlin Runtime",
    "language-runtimes-go-node-php": "Language Runtimes Go Node PHP",
    "messaging-and-streaming": "Messaging And Streaming",
    "network-and-web-protocols": "Network And Web Protocols",
    "os-kernel-computer-architecture": "OS Kernel Computer Architecture",
    "search-and-nosql": "Search And NoSQL",
    "security-and-cryptography": "Security And Cryptography",
    "spring-and-frameworks": "Spring And Frameworks",
    "testing-and-code-design": "Testing And Code Design",
    "miscellaneous": "Miscellaneous",
}


@dataclass(frozen=True)
class Chunk:
    source_file: str
    start_line: int
    end_line: int
    title: str
    body: str
    sha256: str
    topic: str


def normalize(text: str) -> str:
    return text.casefold()


def has_any(text: str, *keywords: str) -> bool:
    haystack = normalize(text)
    return any(normalize(keyword) in haystack for keyword in keywords)


def classify(title: str, body: str) -> str:
    title_text = normalize(title)
    whole_text = normalize(title + "\n" + body[:3000])

    if title == "[source front matter]":
        return "00-source-front-matter-and-question-banks"

    if title_text.strip() in ("# interview", "# interviews"):
        return "00-source-front-matter-and-question-banks"

    if has_any(title_text, "시나리오", "면접", "지원서", "카페24", "프로젝트에서 사용한", "질문 목록"):
        return "00-source-front-matter-and-question-banks"

    if has_any(title_text, "docker", "container", "컨테이너", "kubernetes", "gitlab", "ci/cd", "배포"):
        return "containers-and-devops"

    if has_any(title_text, "elasticsearch", "elasticsearch", "hot/warm", "샤드", "카우치베이스", "couchbase", "nosql"):
        return "search-and-nosql"

    if has_any(title_text, "oauth", "pkce", "csrf", "tls", "https", "dhe", "ecdhe", "전방 비밀성", "암호", "crypt", "ssl", "pkcs", "token", "토큰", "보안", "프라이버시"):
        return "security-and-cryptography"

    if has_any(title_text, "rabbitmq", "rabbit mq", "kafka", "amqp", "stomp", "mqtt", "메시지", "메세징", "amqplib"):
        return "messaging-and-streaming"

    if has_any(title_text, "resttemplate", "webclient", "webflux", "mono", "flux", "tomcat", "spring", "스프링", "@transactional", "bean", "빈 라이프", "빈 생명", "ioc", "dependency injection", "di", "applicationcontext", "beanfactory", "hikari", "어노테이션 프로세싱", "dispatcher"):
        return "spring-and-frameworks"

    if has_any(title_text, "netty", "epoll", "io_uring", "aio_", "libuv", "i/o", "비동기", "논블로킹", "블로킹", "동시성", "멀티스레딩", "스레드", "쓰레드", "thread", "pthread", "synchronized", "wait", "notify", "모니터", "race condition", "경쟁 조건", "deadlock", "교착", "file descriptor", "소켓이 몇 만", "async"):
        return "concurrency-async-io"

    if has_any(title_text, "jvm", "java", "kotlin", "코루틴", "coroutine", "classloader", "클래스 로더", "gc", "garbage", "가비지", "메타스페이스", "infix", "inline", "부모 위임", "when"):
        return "jvm-java-kotlin-runtime"

    if has_any(title_text, "go", "golang", "goroutine", "node.js", "nodejs", "php", "php-fpm", "pdo", "runtime", "런타임"):
        return "language-runtimes-go-node-php"

    if has_any(title_text, "network", "네트워크", "tcp", "ip 구조", "nat", "http", "keep-alive", "grpc", "protocol buffers", "소켓", "nginx", "proxy", "reverse proxy", "컨제스천", "congestion", "streaming", "스트리밍"):
        return "network-and-web-protocols"

    if has_any(title_text, "db", "database", "데이터베이스", "mysql", "postgresql", "acid", "mvcc", "isolation", "격리", "lock", "락", "replication", "레플리케이션", "b tree", "b+ tree", "b+tree", "b+ 트리", "인덱스", "파티셔닝", "트랜잭션", "2단계", "카디널리티", "cardinality", "테이블 처리", "수억건", "3억건"):
        return "database-and-storage"

    if has_any(title_text, "distributed", "분산", "msa", "마이크로서비스", "eventual consistency", "cap", "가용성", "고가용성", "saga", "사가", "raft", "paxos", "architecture", "아키텍처", "로드 밸런싱", "확장성", "결제", "멱등", "토폴로지"):
        return "distributed-systems-and-architecture"

    if has_any(title_text, "os", "운영체제", "커널", "kernel", "systemd", "부팅", "boot", "fork", "exec", "프로세스", "process", "cpu", "메모리", "memory", "파일 시스템", "background process", "부동소수점", "서버용 컴퓨터", "노트북도 서버"):
        return "os-kernel-computer-architecture"

    if has_any(title_text, "algorithm", "알고리즘", "복잡도", "big o", "퀵소트", "자료 구조", "정규표현식", "재귀", "귀납", "그래프"):
        return "algorithms-and-data-structures"

    if has_any(title_text, "test", "테스트", "mock", "stub", "tdd", "solid", "디자인 패턴", "프록시 패턴", "proxy pattern"):
        return "testing-and-code-design"

    if has_any(title_text, "함수형", "functional", "functor", "monad", "either", "arrow"):
        return "functional-programming"

    if has_any(whole_text, "wait()", "notify()", "notifyall", "wait set", "synchronized", "모니터", "blocked", "waiting"):
        return "concurrency-async-io"

    if has_any(whole_text, "면접관", "지원자"):
        return "00-source-front-matter-and-question-banks"

    return "miscellaneous"


def read_chunks(path: Path) -> list[Chunk]:
    text = path.read_text()
    lines = text.splitlines(keepends=True)
    heading_starts = [
        index for index, line in enumerate(lines) if re.match(r"^#{1,2} ", line)
    ]

    if not heading_starts:
        body = "".join(lines)
        title = lines[0].strip() if lines else path.name
        return [make_chunk(path, 0, len(lines), title, body)]

    first_content = None
    for start in heading_starts:
        if lines[start].startswith("## ") or start != 0:
            first_content = start
            break

    if first_content is None:
        body = "".join(lines)
        title = lines[0].strip() if lines else path.name
        return [make_chunk(path, 0, len(lines), title, body)]

    chunks: list[Chunk] = []
    if first_content > 0:
        body = "".join(lines[:first_content])
        chunks.append(make_chunk(path, 0, first_content, "[source front matter]", body))

    content_starts = [start for start in heading_starts if start >= first_content]
    for index, start in enumerate(content_starts):
        end = content_starts[index + 1] if index + 1 < len(content_starts) else len(lines)
        body = "".join(lines[start:end])
        title = lines[start].strip()
        chunks.append(make_chunk(path, start, end, title, body))

    return chunks


def make_chunk(path: Path, start: int, end: int, title: str, body: str) -> Chunk:
    digest = hashlib.sha256(body.encode()).hexdigest()
    return Chunk(
        source_file=path.name,
        start_line=start + 1,
        end_line=end,
        title=title,
        body=body,
        sha256=digest,
        topic=classify(title, body),
    )


def source_ref(chunk: Chunk) -> str:
    return f"{chunk.source_file}:{chunk.start_line}-{chunk.end_line}"


def write_topic_file(topic: str, chunks: list[Chunk], aliases_by_hash: dict[str, list[Chunk]]) -> None:
    path = TOPIC_DIR / f"{topic}.md"
    lines: list[str] = [
        f"# {TOPICS[topic]}",
        "",
        "> 원문 보존형 이동본입니다. 이 파일의 source chunk 본문은 원본 `intervie*.md`에서 그대로 복사되었고, 기술적 보강과 딥 리라이트는 다음 단계에서 수행합니다.",
        "",
        "## Source Chunks",
        "",
    ]

    for chunk in chunks:
        aliases = aliases_by_hash[chunk.sha256]
        alias_refs = ", ".join(source_ref(alias) for alias in aliases)
        lines.extend(
            [
                f"<!-- source-chunk: sha256={chunk.sha256} topic={topic} sources={alias_refs} -->",
                "",
                f"> Source: `{source_ref(chunk)}`",
            ]
        )
        if len(aliases) > 1:
            lines.append(f"> Duplicate source aliases: `{alias_refs}`")
        lines.extend(["", chunk.body.rstrip("\n"), "", "<!-- /source-chunk -->", ""])

    path.write_text("\n".join(lines).rstrip() + "\n")


def write_index(unique_chunks: list[Chunk], aliases_by_hash: dict[str, list[Chunk]]) -> None:
    counts = defaultdict(int)
    source_refs = 0
    for chunk in unique_chunks:
        counts[chunk.topic] += 1
        source_refs += len(aliases_by_hash[chunk.sha256])

    lines = [
        "# Interview Topic Migration Index",
        "",
        "이 디렉터리는 `intervie*.md` 원재료를 주제별 문서로 옮긴 첫 보존 단계입니다.",
        "현재 문서들은 정식 딥 리라이트 결과가 아니라, 원문 손상을 막기 위한 source-preserving staging area입니다.",
        "",
        "## Coverage Summary",
        "",
        f"- source files: {len(SOURCE_NAMES)}",
        f"- source chunks including duplicates: {source_refs}",
        f"- unique chunks after exact-digest deduplication: {len(unique_chunks)}",
        "",
        "## Topic Files",
        "",
    ]
    for topic in TOPICS:
        if counts[topic] == 0:
            continue
        lines.append(f"- [{TOPICS[topic]}]({topic}.md): {counts[topic]} unique chunks")
    lines.extend(
        [
            "",
            "## Verification",
            "",
            "재생성은 아래 명령으로 수행합니다.",
            "",
            "```sh",
            "python3 interviews/tools/split_interview_sources.py",
            "```",
            "",
            "`_migration_manifest.json`에는 모든 원본 chunk의 source span, SHA-256, topic, 대표 chunk 여부가 기록됩니다.",
        ]
    )
    (TOPIC_DIR / "README.md").write_text("\n".join(lines) + "\n")


def write_manifest(all_chunks: list[Chunk], unique_chunks: list[Chunk], aliases_by_hash: dict[str, list[Chunk]]) -> None:
    representative_by_hash = {chunk.sha256: chunk for chunk in unique_chunks}
    manifest = {
        "source_files": SOURCE_NAMES,
        "source_file_hashes": {
            name: hashlib.sha256((ROOT / name).read_bytes()).hexdigest()
            for name in SOURCE_NAMES
        },
        "source_chunk_count": len(all_chunks),
        "unique_chunk_count": len(unique_chunks),
        "topics": TOPICS,
        "chunks": [
            {
                **{key: value for key, value in asdict(chunk).items() if key != "body"},
                "source_ref": source_ref(chunk),
                "representative_ref": source_ref(representative_by_hash[chunk.sha256]),
                "is_representative": chunk is representative_by_hash[chunk.sha256],
            }
            for chunk in all_chunks
        ],
        "deduplicated_groups": [
            {
                "sha256": digest,
                "topic": representative_by_hash[digest].topic,
                "representative_ref": source_ref(representative_by_hash[digest]),
                "source_refs": [source_ref(chunk) for chunk in aliases],
            }
            for digest, aliases in sorted(aliases_by_hash.items())
            if len(aliases) > 1
        ],
    }
    (TOPIC_DIR / "_migration_manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n"
    )


def clear_generated_outputs() -> None:
    for topic in TOPICS:
        (TOPIC_DIR / f"{topic}.md").unlink(missing_ok=True)
    (TOPIC_DIR / "README.md").unlink(missing_ok=True)
    (TOPIC_DIR / "_migration_manifest.json").unlink(missing_ok=True)


def main() -> None:
    TOPIC_DIR.mkdir(exist_ok=True)
    clear_generated_outputs()

    all_chunks: list[Chunk] = []
    for name in SOURCE_NAMES:
        path = ROOT / name
        if not path.exists():
            raise FileNotFoundError(path)
        all_chunks.extend(read_chunks(path))

    aliases_by_hash: dict[str, list[Chunk]] = defaultdict(list)
    unique_by_hash: dict[str, Chunk] = {}
    for chunk in all_chunks:
        aliases_by_hash[chunk.sha256].append(chunk)
        unique_by_hash.setdefault(chunk.sha256, chunk)

    unique_chunks = list(unique_by_hash.values())
    chunks_by_topic: dict[str, list[Chunk]] = defaultdict(list)
    for chunk in unique_chunks:
        chunks_by_topic[chunk.topic].append(chunk)

    for topic, chunks in chunks_by_topic.items():
        write_topic_file(topic, chunks, aliases_by_hash)

    write_index(unique_chunks, aliases_by_hash)
    write_manifest(all_chunks, unique_chunks, aliases_by_hash)

    print(f"source chunks: {len(all_chunks)}")
    print(f"unique chunks: {len(unique_chunks)}")
    for topic in sorted(chunks_by_topic):
        print(f"{topic}: {len(chunks_by_topic[topic])}")


if __name__ == "__main__":
    main()
