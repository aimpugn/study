# Interview Topic Migration Index

이 디렉터리는 `intervie*.md` 원재료를 주제별 문서로 옮긴 첫 보존 단계입니다.
현재 문서들은 정식 딥 리라이트 결과가 아니라, 원문 손상을 막기 위한 source-preserving staging area입니다.

## Coverage Summary

- source files: 7
- source chunks including duplicates: 241
- unique chunks after exact-digest deduplication: 133

## Topic Files

- [Source Front Matter And Question Banks](00-source-front-matter-and-question-banks.md): 9 unique chunks
- [Algorithms And Data Structures](algorithms-and-data-structures.md): 1 unique chunks
- [Containers And DevOps](containers-and-devops.md): 1 unique chunks
- [Concurrency Async IO](concurrency-async-io.md): 34 unique chunks
- [Database And Storage](database-and-storage.md): 8 unique chunks
- [Distributed Systems And Architecture](distributed-systems-and-architecture.md): 4 unique chunks
- [Functional Programming](functional-programming.md): 1 unique chunks
- [JVM Java Kotlin Runtime](jvm-java-kotlin-runtime.md): 11 unique chunks
- [Language Runtimes Go Node PHP](language-runtimes-go-node-php.md): 3 unique chunks
- [Messaging And Streaming](messaging-and-streaming.md): 4 unique chunks
- [Network And Web Protocols](network-and-web-protocols.md): 8 unique chunks
- [OS Kernel Computer Architecture](os-kernel-computer-architecture.md): 9 unique chunks
- [Search And NoSQL](search-and-nosql.md): 6 unique chunks
- [Security And Cryptography](security-and-cryptography.md): 8 unique chunks
- [Spring And Frameworks](spring-and-frameworks.md): 24 unique chunks
- [Testing And Code Design](testing-and-code-design.md): 2 unique chunks

## Verification

재생성은 아래 명령으로 수행합니다.

```sh
python3 interviews/tools/split_interview_sources.py
```

`_migration_manifest.json`에는 모든 원본 chunk의 source span, SHA-256, topic, 대표 chunk 여부가 기록됩니다.
