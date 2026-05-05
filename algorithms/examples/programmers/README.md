# Programmers Practice

이 디렉터리는 프로그래머스 문제 풀이를 반복해서 훈련하는 공간입니다.

풀이 코드는 지금처럼 언어별 소스 루트 아래에 둡니다.

```text
examples/programmers/jvm/src/main/java/p<problemId>/Main.java
examples/programmers/jvm/src/main/kotlin/p<problemId>/Main.kt
```

반복 훈련 상태는 코드 패키지와 분리해서 [`PRACTICE_QUEUE.md`](PRACTICE_QUEUE.md)에 기록합니다.
코드는 "현재 가장 다시 제출하기 좋은 풀이"를 담고, 반복 큐는 "언제 다시 풀지, 어떤 생각을 복원해야 하는지"를 담습니다.

## 왜 분리하는가

같은 문제를 여러 번 풀 때마다 `p42895_v2`, `p42895_retry` 같은 패키지를 늘리면 나중에 어떤 코드가 현재 기준인지 헷갈립니다.
또 Java/Kotlin에서는 같은 패키지 안에 제출용 `Solution` 클래스를 여러 개 두기도 불편합니다.
그래서 기본 규칙은 단순하게 둡니다.

한 문제와 한 언어에는 현재 풀이 파일 하나만 둡니다.
다시 풀 때는 같은 파일을 다시 작성하고, 막혔던 이유나 복원해야 할 생각은 큐 문서나 문제별 `PROCESS.md`에 남깁니다.

## 반복 루프

문제 하나를 완전히 내 것으로 만들려면 한 번 통과한 것으로는 부족합니다.
이 저장소에서는 아래 네 번을 한 사이클로 봅니다.

1. R1: 처음 풀이합니다. 힌트나 첨삭을 받아도 됩니다.
2. R2: 2~3일 뒤 빈 파일에서 다시 구현합니다.
3. R3: 1주 뒤 코드 없이 풀이 흐름을 5분 안에 설명합니다.
4. R4: 2~3주 뒤 시간 제한을 두고 다시 풉니다.

풀었다는 것은 채점 통과이고, 공부했다는 것은 다음에 같은 냄새가 났을 때 첫 생각이 떠오르는 것입니다.
내 것이 됐다는 것은 2주 뒤에도 빈 파일에서 다시 구현할 수 있다는 뜻입니다.

## 새 문제 준비

새 문제를 시작할 때는 `programmers-jvm-setup` 스킬로 언어와 문제 URL을 넘겨 scaffold를 만듭니다.

```text
$programmers-jvm-setup java https://school.programmers.co.kr/learn/courses/30/lessons/<problemId> 준비
```

스킬은 문제 제목, 공식 난이도 표시, 자체 체감 난이도, 예상 소요 시간, 첫 풀이 방향을 함께 확인합니다.
