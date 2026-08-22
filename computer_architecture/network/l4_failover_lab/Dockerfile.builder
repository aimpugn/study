# 측정 도구를 컴파일하는 이미지.
# clippy 를 미리 넣어 두어야 문서 강제 lint(missing_docs_in_private_items)를 돌릴 수 있습니다.
# Docker 가 이미지를 캐시하므로 이 설치는 처음 한 번만 걸립니다.
FROM rust:alpine
RUN rustup component add clippy
