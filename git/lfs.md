# LFS

Git mirror 이전에서는 Git ref와 LFS 실제 파일을 별도로 전송해야 합니다. 전체 절차와 검증 방법은 [`git clone --mirror` 결과를 새 원격 저장소에 그대로 반영하기](git_clone_mirror.md#6-lfs-객체를-원본에서-모두-받는다)를 참고합니다.

## LFS 파일 확인

```shell
git lfs ls-files
```

## LFS 파일 삭제
