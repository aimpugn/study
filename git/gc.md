# gc

## git gc?

Git의 가비지 컬렉션(`git gc`) 프로세스는 불필요한 파일을 정리하고 로컬 저장소를 최적화합니다.

- 활동이 많은 저장소에서는 정기적으로 `git gc`를 실행하여 최적의 성능을 유지합니다.
- 주요 브랜치 변경이나 정리 작업 후에는 도달할 수 없는 객체들을 주기적으로 확인하고 제거합니다.
- `.gitignore` 파일을 적절하게 설정하여 불필요한 파일이 추적되지 않도록 합니다.

이러한 단계를 따르면 경고 메시지를 방지하고 Git 저장소를 더 깨끗하고 효율적으로 유지할 수 있습니다.

## There are too many unreachable loose objects

```bash
From https://github.com/my-org/arepo
 + 65c4a7b8...d3d106e8 subservice1/subservice1 -> origin/subservice1/subservice1  (forced update)
Auto packing the repository in background for optimum performance.
See "git help gc" for manual housekeeping.
warning: The last gc run reported the following. Please correct the root cause
and remove .git/gc.log
Automatic cleanup will not be performed until the file is removed.

warning: There are too many unreachable loose objects; run 'git prune' to remove them.
```

1. **도달할 수 없는 느슨한 객체**(unreachable loose objects):

    Git은 저장소의 모든 변경 내역을 관리합니다.
    브랜치 삭제나 rebase와 같은 작업으로 커밋이 도달할 수 없게 되면, 해당 커밋에 연결된 객체들이 "느슨한 객체"가 됩니다.
    이 객체들은 즉시 삭제되지 않고 복구를 위해 잠시 남겨집니다.

    "There are too many unreachable loose objects" 메시지는 저장소에 이러한 객체가 많이 쌓였음을 의미하며,
    이를 정리할 필요가 있음을 나타냅니다.

2. **가비지 컬렉션 (gc)**:

    Git의 가비지 컬렉션(`git gc`) 프로세스는 불필요한 파일을 정리하고 로컬 저장소를 최적화합니다.
    자동 가비지 컬렉션은 특정 조건에서 자동으로 실행되지만, 중단되거나 특정 조건이 충족되면 수동으로 실행해야 할 수 있습니다.

    "Auto packing the repository in the background" 메시지는 Git이 저장소를 최적화하는 중임을 나타냅니다.

이러한 메시지를 방지하고 해결하기 위해 다음 단계를 따릅니다.

1. **수동으로 가비지 컬렉션 실행**:

    가비지 컬렉션을 수동으로 트리거하면 느슨한 객체가 정리되고 저장소가 최적화됩니다.

    ```bash
    git gc
    ```

2. **도달할 수 없는 객체 제거**:

    더 이상 브랜치나 태그에서 참조되지 않는 객체들을 제거합니다.
    참고로 `git gc`는 기본적으로 `git prune`도 실행하므로, 이를 별도로 실행하면 모든 도달할 수 없는 객체가 제거됩니다.

    ```bash
    git prune
    ```

3. **gc.log 파일 제거**:

    `.git/gc.log` 파일에 대한 경고는 가비지 컬렉션 프로세스가 중단되었거나 완료되지 않았음을 나타냅니다.
    이 파일을 수동으로 제거하여 이후 자동 가비지 컬렉션이 정상적으로 작동하도록 하세요.

    ```bash
    rm .git/gc.log
    ```
