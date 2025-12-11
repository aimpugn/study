질문에 있는 그림을 한 문장으로 요약하면 이겁니다.

“만든 `app.jar`와 JDK 25를 OCI 이미지로 묶고, 이를 실행할 수 있는 `udocker` 런타임까지 같이 동봉해서, 사용자는 Docker 설치 없이 `./run.sh` 한 번만 실행하면 `udocker`가 컨테이너 안에서 JDK 25가 `app.jar`가 애플리케이션이 실행되도록 만들 수 있을까? 이게 native-image+musl처럼 환경 독립적이면서도, JVM 리플렉션·동적 로딩 같은 기능은 그대로 쓰게 해 줄 수 있을까?”

그 대신, 어디까지가 진짜로 “환경 무관”이고 어디부터는 커널·아키텍처 의존성이 남는지, native-image와 비교하면 어떤 트레이드오프가 생기는지까지 같이 봐야 합니다.

---

### udocker가 정확히 뭘 해주는지부터 짚고 가야 합니다

공식 매뉴얼을 보면 udocker는 “Docker가 설치되어 있지 않은 리눅스 시스템에서, 일반 사용자 권한으로 Docker 컨테이너를 실행하게 해주는 유저 공간 도구”라고 정의되어 있습니다. Docker 데몬도 필요 없고, root 권한도 필요 없고, 그냥 유저 홈 디렉토리에 udocker를 풀어 두고 실행해 쓸 수 있다고 되어 있습니다.

내부적으로는 Docker 이미지 포맷(OCI 계열)을 그대로 받아서, 로컬에 저장하고, 여러 가지 “실행 엔진”으로 이미지 루트 파일시스템을 실행합니다. PRoot, Fakechroot, runc rootless, Singularity 등 여러 모드를 제공하는데, 매뉴얼과 슬라이드를 보면 각 모드가 chroot·네임스페이스·ptrace 등을 조합해서 권한 없이 컨테이너 비슷한 환경을 만들어낸다고 설명합니다.

또 중요한 특징 하나가, udocker 배포 자체를 “엔드 유저가 그냥 tarball로 가져와서 자기 홈 디렉토리에 풀고 쓴다”고 상정하고 있다는 점입니다. 공식 자료에서 “필요한 바이너리들은 정적으로 컴파일되어 있고, udocker는 그냥 복사해서 풀고 쓰면 된다, 서로 다른 사이트에 컨테이너를 가져가 실행하기에 이상적이다”라고 강조합니다.

즉, 이미 설계 철학 자체가 “Docker 없는 곳에 컨테이너 런타임과 이미지를 같이 들고 가서 실행하는 도구”입니다.

이걸 그대로 제품 배포 모델로 가져오는 게 지금 질문의 아이디어입니다.

---

### 지금 떠올리신 배포 구조는 기술적으로 그대로 구현 가능합니다

머릿속에 그리신 구조를 조금 더 구체적인 그림으로 바꿔보면 대략 이런 흐름입니다.

개발 시점에는 평범한 JVM 애플리케이션을 만든다고 가정합니다. JDK 25와 Spring Boot든, 순수 Spring이든, 어쨌든 최종 결과물은 `app.jar` 같은 fat-jar 또는 layer-jar입니다.

그 다음 단계에서 이 `app.jar`를 포함하는 컨테이너 이미지를 빌드합니다. 예를 들어 `debian-slim`이나 `ubi`, `distroless/java-debian` 같은 베이스를 쓰고 그 위에 JDK 25, 필요한 glibc/zlib 등을 깔고, `/opt/app/app.jar`를 추가해서 `ENTRYPOINT ["java", "-jar", "/opt/app/app.jar"]`로 만든다고 합시다. 이 이미지는 평범한 Docker/OCI 이미지입니다.

그 다음, 빌드된 이미지를 `docker save` 같은 걸로 tar 파일로 뽑거나, udocker가 직접 registry에서 pull 할 수 있게 레지스트리에 푸시합니다. udocker는 Docker Hub 등의 레지스트리와 연동하는 기능을 갖고 있어서, 사용자가 권한만 있으면 `udocker pull your-registry/yourimage:tag`로 이미지를 받아올 수 있습니다.

마지막으로, 배포 아티팩트를 구성합니다. 이 단계에서 당신이 하고 싶은 건 “사용자 입장에서는 그냥 `./run.sh`만 실행하게 하고 싶다”는 요구입니다. 그러면 tar.gz 하나를 만들 때 다음을 묶어 넣을 수 있습니다.

예시를 코드로 쓰면 대략 이런 식의 구조와 런처가 됩니다.

```bash
#!/usr/bin/env bash
# run.sh: 유저가 실행하는 단일 진입점

set -euo pipefail

# 현재 스크립트가 있는 디렉토리 기준으로 상대 경로를 잡습니다.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1) udocker 바이너리/스크립트 위치 설정
UDOCKER_BIN="${SCRIPT_DIR}/udocker/udocker"   # udocker 배포물을 함께 동봉했다고 가정

# 2) udocker가 내부적으로 쓸 작업 디렉토리를 유저 홈이나 로컬에 하나 잡습니다.
export UDOCKER_DIR="${SCRIPT_DIR}/.udocker"   # 시스템 전역이 아니라, 이 앱 전용 udocker 디렉토리

# 3) udocker가 초기화되지 않았다면 한 번 초기화합니다.
if [ ! -d "${UDOCKER_DIR}" ]; then
  "${UDOCKER_BIN}" install    # udocker 자기 자신 설치 (스크립트가 내부 구조를 셋업)
fi

# 4) 우리의 애플리케이션 이미지를 udocker 로컬 저장소에 등록합니다.
#    이미 등록되어 있다면 건너뜁니다. 태그 이름은 예시로 "myapp:1.0".
if ! "${UDOCKER_BIN}" images | grep -q "myapp:1.0"; then
  # 로컬 파일에서 불러오는 경우 (run.sh 옆에 myapp.tar 가 있다고 가정)
  "${UDOCKER_BIN}" load -i "${SCRIPT_DIR}/images/myapp.tar"
  # 또는 레지스트리에서 당겨오는 경우:
  # "${UDOCKER_BIN}" pull my-registry.example.com/myapp:1.0
fi

# 5) 컨테이너를 실행합니다.
#    이 안에서는 우리가 Dockerfile에서 정의한 ENTRYPOINT, 즉 "java -jar app.jar"가 실행됩니다.
#    호스트의 JDK, glibc와는 무관하게, 이미지 안의 JDK 25, glibc를 사용하게 됩니다.
exec "${UDOCKER_BIN}" run \
  --rm \
  myapp:1.0
```

여기서 핵심은, 사용자는 그냥 `./run.sh`만 실행한다는 점입니다. 내부에서 무슨 일이 벌어지는지는 숨겨져 있습니다.

스크립트가 처음 실행될 때는 udocker를 자기 디렉토리 안에 “셀프 설치”하고, 이미지 tar를 로컬 udocker 리포지토리에 import 하고, 이후에는 `udocker run`으로 컨테이너를 띄웁니다.

컨테이너 안에서는 그냥 평범한 리눅스 유저 공간입니다. JDK 25가 설치되어 있고, `java -jar app.jar`가 실행됩니다. JVM 입장에서는 “내가 리눅스 한 대에서 java를 띄웠구나”일 뿐이고, 리플렉션, ServiceLoader, 동적 클래스 로딩, 에이전트, JMX 등 일반 JVM 기능을 그대로 쓸 수 있습니다.

여기까지는 질문에서 상상한 구조와 완전히 일치하고, udocker의 사용 패턴과도 잘 맞습니다. udocker 공식 슬라이드에서도 “udocker tarball을 배치 작업과 함께 전송해서, 서로 다른 사이트에서 Docker 없이 동일 컨테이너를 실행할 수 있다”는 패턴을 예시로 보여줍니다.

즉, “이미지 + udocker 런타임 동봉 + shell 런처”라는 아이디어 자체는, 이미 검증된 사용 패턴입니다.

---

### 이 방식이 native-image + musl 과 비교했을 때 갖는 호환성의 성격

이제 중요한 질문으로 돌아가야 합니다.

“이렇게 하면 native-image+musl과 비슷하거나, 어떤 면에서는 더 많은 호환성으로 배포할 수 있나?”

여기서 “호환성”을 두 가지 층으로 나누어 보는 게 이해에 도움이 됩니다.

첫째는 “어디서 실행이 되느냐”라는 시스템 호환성입니다.
둘째는 “실행되었을 때 런타임 기능을 얼마나 풍부하게 쓸 수 있느냐”라는 언어/런타임 호환성입니다.

시스템 호환성 측면에서 보면, native-image+musl 정적 바이너리는 “해당 아키텍처용 리눅스 커널”만 있으면 됩니다. glibc도 필요 없고, 시스템에 어떤 패키지가 깔려 있든 상관 없습니다. 다만, musl로 링크된 바이너리가 기대하는 최소 커널 버전은 존재하고(보통 2.6.32 이상 같은 수준), 그 아래의 골동품 커널에서는 동작하지 않을 수 있습니다. 그래도 오늘날 운영되는 리눅스 서버 대부분은 이 최소 조건을 넘기 때문에, 현실적으로는 “같은 아키텍처라면 거의 모든 리눅스 배포판에서 돌아간다”에 가깝습니다.

udocker+컨테이너+JDK25 조합은, 시스템 호환성 관점에서 보면 이렇게 됩니다.

컨테이너는 호스트 커널을 그대로 공유합니다. 이것은 Docker든 udocker든, 컨테이너 기술의 공통 전제입니다. 그래서 컨테이너 이미지 안에 넣은 glibc, JDK, zlib 등은 “최소 커널 버전”을 가정합니다. 예를 들어 Debian 12 기반 glibc는 대략 3.x 이상의 커널을 전제합니다. 이건 native-image+musl과 비슷하게 “최소 커널 버전” 제약을 가집니다.

다만, 컨테이너 안에는 glibc, zlib, JDK가 다 들어 있기 때문에 **호스트의 glibc나 호스트에 설치된 Java 버전에는 거의 완전히 독립**입니다. 호스트가 glibc 2.12든 2.37이든 상관 없이, 컨테이너 안에서 우리가 넣어 둔 glibc가 쓰입니다. 이 지점에서 plain JDK 동봉 방식보다 한 단계 더 분리된 구조가 됩니다.

udocker 자체는 여러 실행 모드를 제공합니다. PRoot 기반 모드는 `ptrace`를 써서 시스템 콜을 가로채고, Fakechroot 기반 모드는 ELF 헤더를 조작하고 LD_LIBRARY_PATH를 바꿔서 chroot 비슷한 효과를 내고, runc/crun 기반 모드는 user namespace를 사용하는 rootless 컨테이너를 만듭니다.

덕분에, 커널에 user namespace가 꺼져 있거나, 일부 기능이 막혀 있어도 어느 정도 대응할 수 있게 되어 있지만, 아주 오래된 커널이나 보안 설정이 과하게 제한적인 서버에서는 특정 모드가 동작하지 않을 수 있습니다. 공식 논문과 매뉴얼에서도 “호스트의 기능에 따라 가능한 실행 모드를 골라 쓴다”고 설명합니다.

정리하면, 시스템 호환성 관점에서 이 조합은 다음과 같은 성질을 갖습니다.

호스트의 glibc, libstdc++, Java 설치 여부와는 거의 무관해집니다.
호스트는 리눅스 커널과 CPU 아키텍처만 제공하면 되고, 나머지 유저랜드는 이미지 안에 모두 들어 있습니다.
udocker가 기대하는 커널 기능들(user namespace, ptrace, seccomp 등)이 막혀 있거나 너무 오래된 커널이면 일부 모드는 동작하지 않지만, 여러 모드 중 하나로 우회할 수 있는 여지가 있습니다.

그래서 “같은 아키텍처·적당히 최신 커널”이라는 전제 아래에서는, plain JDK 동봉 방식보다, native-image+musl에 훨씬 가까운 포터블한 배포 모델을 구현할 수 있습니다.

런타임 호환성 측면에서 보면, native-image는 “closed-world” 제약을 갖습니다. 즉, 리플렉션, 동적 클래스 로딩, 프록시, 서비스 디스커버리 등은 빌드 타임 구성에 따라 제한되거나, 완전히 지원되지 않기도 합니다. Spring Native 시절에 보았던 것처럼, 리플렉션이 조금만 복잡해져도 config가 지옥이 됩니다.

반대로, udocker+JDK 컨테이너 방식에서 JVM은 그냥 “진짜 HotSpot”입니다. JIT, GC, JVMTI, Attach, 리플렉션, 동적 클래스로더, 모든 게 그대로입니다. 시스템 콜 몇 개를 intercept 하거나 네임스페이스로 격리하는 정도가 추가될 뿐, JVM 내부는 네이티브와 동일하게 동작합니다.

이 의미에서 “호환성”을 런타임 기능 관점까지 확장해서 보면, udocker+JDK 방식은 native-image보다 훨씬 풍부한 호환성을 제공합니다. “원래 JVM에서 되던 건 다 된다”라고 말할 수 있기 때문입니다.

---

### 이 구조를 쓸 때 현실적으로 고려해야 하는 제약과 비용

가능하냐, 호환성은 어떠냐를 넘어서, “이걸 실제 제품 배포 전략으로 쓸 때 무엇을 감수해야 하느냐”를 짚어야 합니다.

우선, 아티팩트 크기입니다. 이미지 안에 베이스 OS, glibc, JDK 25, application jar까지 들어갑니다. 여기에 udocker까지 붙이면, x86_64 기준으로 수백 MB는 금방 넘어갑니다. native-image+musl의 수십 MB와 비교하면 덩치가 꽤 큽니다.

둘째, 성능 오버헤드입니다. udocker의 PRoot 기반 모드(P1, P2)는 ptrace로 시스템 콜을 감시하며 파일 경로를 수정하기 때문에, 시스템 콜이 많은 워크로드에서는 눈에 띄는 오버헤드가 발생할 수 있습니다. Fakechroot 기반 모드는 ELF 헤더 변조와 LD_LIBRARY_PATH를 이용하지만, 이 역시 완전히 네이티브 컨테이너보다는 약간의 마찰이 있습니다. runc/crun 기반 rootless 모드는 거의 Docker에 가까운 성능이지만, user namespace 활성화 등이 필요하고, 모든 환경에서 가능하지는 않습니다.

셋째, udocker 자체의 의존성입니다. udocker는 기본적으로 Python으로 작성되어 있고, 일부 배포에서는 Python 2에 맞춰져 있었다는 이슈도 있습니다. 공식 배포본이나 최신 포크들은 Python 3를 지원하지만, 어쨌든 호스트에 최소한의 Python 런타임이 있어야 합니다. 앞서 본 슬라이드에서는 “필요한 바이너리를 정적으로 포함해 tarball 하나로 배포할 수 있다”고 되어 있으니, 실제로는 udocker와 그 내부 바이너리를 우리 패키지 안에 같이 넣어두고, Python까지 자체적으로 포함시키는 전략도 가능합니다. 다만 이 경우는 “JDK 동봉”에 더해 “Python 런타임 + udocker 스택까지 동봉”하는 셈이므로, 덩치와 복잡도가 또 증가합니다.

넷째, 보안 격리입니다. udocker는 “보안 샌드박스”를 표방하지 않습니다. 문서에서도 “루트 권한 없이 컨테이너를 실행하는 유틸리티”이지, 강한 격리나 멀티테넌트 보안을 제공하는 도구가 아니라는 점을 강조합니다. 결국 컨테이너 안에서의 프로세스는 호스트에서 동일한 UID로 실행되고, 커널을 공유합니다. 당신의 요구는 “소비자가 신경 쓰지 않고 실행만 하도록 만들고 싶다”이지, 강한 샌드박스는 아니기 때문에, 이 제약은 크게 문제 되지 않을 수도 있습니다. 하지만 누군가 “컨테이너니까 안전하겠지?”라고 오해하면 곤란합니다.

이 네 가지를 합쳐보면, 이 구조는 사실상 이렇게 요약됩니다.

“native-image+musl 수준의 glibc/라이브러리 독립성에, JVM 수준의 런타임 호환성을 더 얹는 대신, 아티팩트 크기와 복잡도를 크게 치르는 전략”입니다.

대신 Docker 설치를 강제하지 않아도 되고, 사용자는 여전히 “그냥 실행”만 하면 됩니다.

---

### 결론적으로, 이 전략은 “가능하고, 꽤 그럴듯한 절충안”입니다

컨테이너 이미지 안에 JDK 25와 앱을 넣고, 그 이미지를 udocker와 함께 동봉한 뒤, 쉘 스크립트 하나로 `udocker run`을 감추는 방식은 기술적으로 충분히 구현 가능하고, udocker의 설계 철학과도 잘 맞습니다.

이렇게 하면 호스트에 Docker를 설치할 필요가 없고, 호스트에 Java가 설치되어 있을 필요도 없으며, 호스트 glibc 버전과도 상당 부분 독립적인 배포 모델을 만들 수 있습니다. 사용자는 `./run.sh`만 실행하면 JDK 25 위에서 Spring 애플리케이션이 정상적인 JVM 모드로 돌아가고, 리플렉션과 동적 클래스 로딩, 플러그인 구조를 그대로 쓸 수 있습니다.

환경 독립성의 성격은 native-image+musl과 비슷한 수준으로 올라가지만, 런타임 기능 면에서는 오히려 native-image보다 훨씬 풍부합니다. 그 대가로 아티팩트 크기와 udocker 스택의 복잡도, 그리고 일부 커널 기능 의존성이 따라붙습니다.

리눅스 x86_64만 타깃으로 하고, “사용자에게 Docker 설치를 강제하지 않으면서도, 최대한 환경 독립적인 JVM 런타임을 통째로 들고 다니고 싶다”라는 요구라면,

“OCI 이미지 + udocker 번들 + run.sh 런처” 모델은, native-image와 plain JDK 동봉 사이의 꽤 설득력 있는 절충안입니다.

다만 이 접근을 채택한다면, 반드시 초기에 몇 가지를 명확히 해야 합니다.
- 어떤 최소 커널 버전과 아키텍처를 공식 지원 범위로 잡을지
- udocker 실행 모드 중 무엇을 기본값으로 할지
- Python 의존성을 어떻게 처리할지
- 그리고 성능 오버헤드가 허용 가능한지입니다.

---

“OCI 이미지 + udocker 조합으로 JVM을 띄워 놓았을 때, 이미 돌아가고 있는 그 컨테이너 안으로 JAR 파일을 꽂아 넣어서, 진짜 ‘plug & play’처럼 기능을 바로 추가할 수 있나? 컨테이너 레벨에서 그런 게 가능한가?”

여기서 관점을 먼저 정리해야 합니다.

컨테이너는 “프로세스를 어떻게 띄우고 격리하느냐”를 책임지고,
JVM은 “프로세스 안에서 클래스와 JAR을 어떻게 로딩하느냐”를 책임집니다.

지금 질문은 사실 둘을 섞어서 생각하고 있는 상태입니다.
그래서 둘을 분해해서 보면 답이 명확해집니다.

---

먼저 결론부터 말하면 이렇습니다.

“OCI 이미지 + udocker”라는 조합만으로는,
이미 돌아가고 있는 JVM 프로세스의 classpath를 바깥에서 바꾸거나,
JAR을 꽂았다고 자동으로 새 기능이 붙는 식의 plug & play는 불가능합니다.

하지만,

컨테이너 안에서 돌아가는 애플리케이션을
“플러그인 JAR을 동적으로 로딩하도록 설계”해 두면,
컨테이너가 udocker든 Docker든 상관없이,
실행 중에 JAR을 복사하고 애플리케이션이 그걸 읽어서 붙이는 구조는 충분히 만들 수 있습니다.

차이는 이겁니다.

컨테이너 레벨만으로는 “코어가 전혀 그런 의식이 없는 JVM”에 JAR을 꽂아 넣을 수 없고,
애플리케이션 레벨에서 명시적으로 플러그인 시스템을 설계해야 한다는 것.

이걸 단계별로 풀어서 보겠습니다.

---

컨테이너 입장에서 보면 “이미지”는 항상 정적입니다.

OCI 이미지든 Docker 이미지든, “이미지”라는 것은 레이어들의 스냅샷입니다.
컨테이너를 한 번 띄우면, 그 시점의 이미지 레이어에 쓰기 가능한 레이어가 하나 덧붙어서 rootfs가 됩니다.

중요한 포인트는, 이 rootfs 위에서 돌아가는 건 그냥 “리눅스 프로세스”라는 것입니다.
udocker가 컨테이너를 실행한다는 건, 결국

* 어떤 디렉터리를 root처럼 보이게 만들고
* 그 안에서 `/usr/bin/java -jar /opt/app/app.jar` 같은 프로세스를 하나 띄우는 일

을 해주는 것뿐입니다.

여기서 JVM은 평범한 프로세스입니다.
컨테이너 런타임(udocker, Docker 등)은 이 프로세스의 내부 상태,
예를 들어 classpath에 어떤 JAR이 올라가 있는지, 어떤 클래스가 로딩되어 있는지에 개입할 수 없습니다.

그래서 “이미 실행 중인 컨테이너에 JAR을 꽂는다”는 말을 좀 더 엄밀하게 바꾸면 이렇게 됩니다.

이미 실행 중인 JVM 프로세스의 메모리 안으로
새 JAR의 바이트코드를 주입해서,
classloader가 그걸 보고 새로운 클래스를 로딩하게 만들고 싶다.

이건 컨테이너의 역할이 아니라, JVM 내부의 classloader와 애플리케이션 설계가 책임져야 하는 영역입니다.

---

그럼 컨테이너에서는 아무 것도 못하냐 하면, 그건 아닙니다.

컨테이너는 “파일 시스템”을 제공합니다.
구체적으로는, 컨테이너를 띄울 때

* 이미지 레이어에서 rootfs를 만들고
* 필요하면 호스트 디렉터리를 특정 경로에 마운트해 줄 수 있습니다.

udocker도 Docker처럼 `-v hostdir:containerdir` 형태로 호스트 디렉터리를 컨테이너 안에 보이게 할 수 있습니다.

이 말은, “플러그인 JAR들이 들어 있는 디렉터리”를
컨테이너 안의 `/opt/app/plugins` 같은 경로에 마운트해 둘 수 있다는 뜻입니다.

그 다음부터는 컨테이너는 아무 것도 하지 않습니다.
단지 “플러그인 디렉터리”를 JVM이 볼 수 있게 해줄 뿐입니다.

실제로 plug & play가 되느냐 안 되느냐는,
JVM 안에서 돌아가는 애플리케이션이 “그 디렉터리를 어떻게 쓰도록 설계되었느냐”에 달려 있습니다.

---

여기서 시나리오를 두 개로 나눠볼 수 있습니다.

첫 번째 시나리오는, “애플리케이션이 플러그인 시스템을 갖고 있지 않은 경우”입니다.

단순한 Spring Boot 앱을 생각해 봅시다.
실행 시점에 `java -cp app.jar SomeMain`으로 떠 있습니다.
이 앱은 시작할 때 classpath를 기반으로 모든 빈과 클래스들을 로딩해 놓고, 그 뒤에는 새로운 JAR을 전혀 고려하지 않습니다.

이 상황에서 컨테이너 밖에서 아무리 JAR을 컨테이너 rootfs에 복사해도,
이미 떠 있는 JVM은 그 JAR이 존재한다는 사실 자체를 모릅니다.

애플리케이션 코드가 “새 JAR이 생겼는지 확인하고, 거기서 클래스를 로딩하는” 기능을 전혀 갖고 있지 않기 때문에,
컨테이너 레벨에서 할 수 있는 일은 없습니다.

이 경우 JAR을 플러그인처럼 쓰려면,
결국 컨테이너를 재시작해서 JVM을 다시 띄우고,
새 JAR을 classpath에 포함시키는 수밖에 없습니다.

컨테이너 상식으로 말하면 “이미지를 새로 빌드해서 새 컨테이너를 띄운다”는 패턴이 바로 이것입니다.

두 번째 시나리오는, “애플리케이션 안에 명시적으로 플러그인 시스템을 설계한 경우”입니다.

만약 애플리케이션이 다음과 같은 구조를 갖고 있다면 상황이 달라집니다.

* `/opt/app/plugins` 디렉터리 안의 JAR을 플러그인으로 취급한다.
* 시작할 때 이 디렉터리를 스캔해서 JAR 목록을 얻는다.
* 각 JAR에 대해 URLClassLoader 같은 별도의 classloader를 만들어 필요한 인터페이스 구현을 찾는다.
* 필요하다면 주기적으로 디렉터리를 다시 스캔하거나, 파일 시스템 watcher를 걸어 새로운 플러그인이 생기면 로딩한다.

이렇게 설계되어 있다면, 컨테이너가 udocker든 Docker든, 아니면 bare-metal JVM이든 상관없이,
“JAR 파일을 복사한 즉시” 혹은 “주기적 스캔이 도는 시점”에 플러그인이 활성화될 수 있습니다.

여기서 컨테이너의 역할은 단 하나입니다.
플러그인 디렉터리를 애플리케이션이 볼 수 있게 마운트해 주는 것.
그 이후의 dynamic plug & play는 순전히 애플리케이션의 플러그인 시스템이 처리합니다.

즉, “컨테이너 + udocker → dynamic plug & play”가 아니라,
“플러그인 시스템을 가진 JVM 애플리케이션 → 컨테이너든 udocker든 어디서나 plug & play 가능”입니다.

---

조금 더 구체적인 감을 위해, 아주 단순한 플러그인 로더를 코드로 한 번 써보겠습니다.

진짜 실무용은 훨씬 복잡해지겠지만,
“어떤 원리로 돌아가는지”를 보는 데에는 이 정도로 충분합니다.

```java
// Plugin 인터페이스: 플러그인이 구현해야 하는 계약입니다.
// 이 인터페이스는 코어 애플리케이션의 classpath에 포함됩니다.
public interface Plugin {
    // 플러그인이 제공하는 기능의 엔트리 포인트라고 가정합니다.
    String name();          // 플러그인 이름
    void execute();         // 플러그인 실행 로직
}
```

```java
import java.io.File;
import java.net.URL;
import java.net.URLClassLoader;
import java.util.ArrayList;
import java.util.List;

// 아주 단순한 플러그인 로더 예시입니다.
// - /opt/app/plugins 디렉터리 안의 모든 .jar 를 찾아서
// - 각 JAR마다 별도의 ClassLoader를 만들고
// - 특정 FQCN (예: "com.example.PluginImpl") 을 로딩해서 Plugin 인터페이스로 캐스팅합니다.
public class PluginLoader {

    // 플러그인 JAR 이 놓일 디렉터리 경로입니다.
    // 컨테이너를 띄울 때 이 경로에 호스트 디렉터리를 -v 로 마운트해 두면,
    // 호스트에서 JAR 파일만 복사해도 컨테이너 안 JVM 에서 볼 수 있습니다.
    private final File pluginDir;

    public PluginLoader(String pluginDirPath) {
        this.pluginDir = new File(pluginDirPath);
    }

    public List<Plugin> loadPlugins() throws Exception {
        List<Plugin> plugins = new ArrayList<>();

        if (!pluginDir.isDirectory()) {
            System.out.println("플러그인 디렉터리가 존재하지 않습니다: " + pluginDir.getAbsolutePath());
            return plugins;
        }

        // 디렉터리 안의 모든 .jar 파일을 순회합니다.
        File[] jars = pluginDir.listFiles((dir, name) -> name.endsWith(".jar"));
        if (jars == null) {
            return plugins;
        }

        for (File jar : jars) {
            System.out.println("플러그인 JAR 발견: " + jar.getName());

            // JAR 하나당 URLClassLoader 하나를 만듭니다.
            // parent 로는 코어 애플리케이션의 ClassLoader 를 넘겨서 Plugin 인터페이스를 공유합니다.
            URL jarUrl = jar.toURI().toURL();
            URLClassLoader cl = new URLClassLoader(
                    new URL[]{jarUrl},
                    Plugin.class.getClassLoader() // 코어 인터페이스가 로딩된 클래스 로더
            );

            // 아주 단순화를 위해, 각 JAR 안에 "com.example.PluginImpl" 이라는 구현체가 있다고 가정합니다.
            // 실제로는 META-INF/services, 스캔, 어노테이션 기반 등으로 확장합니다.
            Class<?> implClass = cl.loadClass("com.example.PluginImpl");

            // Plugin 인터페이스로 캐스팅해서 인스턴스를 생성합니다.
            Object instance = implClass.getDeclaredConstructor().newInstance();
            if (!(instance instanceof Plugin plugin)) {
                throw new IllegalStateException("플러그인 구현체가 Plugin 인터페이스를 구현하지 않습니다: " + implClass);
            }

            plugins.add(plugin);
        }

        return plugins;
    }
}
```

```java
public class App {

    public static void main(String[] args) throws Exception {
        // 컨테이너 안에서의 플러그인 디렉터리 경로
        // 예: udocker run 에서 -v /host/plugins:/opt/app/plugins 로 마운트했다고 가정합니다.
        String pluginDir = "/opt/app/plugins";

        PluginLoader loader = new PluginLoader(pluginDir);

        // 1. 시작 시점에 한 번 플러그인들을 모두 로딩합니다.
        //    이때 디렉터리 안에 아무 JAR도 없다면, 그냥 빈 목록이 됩니다.
        var plugins = loader.loadPlugins();
        for (Plugin plugin : plugins) {
            System.out.println("플러그인 로드됨: " + plugin.name());
        }

        // 2. 매우 단순한 예로, 10초마다 한 번씩 플러그인을 다시 로드해서 실행해보겠습니다.
        //    실제로는 더 정교한 변경 감지/캐시/언로드 설계가 필요합니다.
        while (true) {
            System.out.println("=== 플러그인 재로딩 시도 ===");
            plugins = loader.loadPlugins();
            for (Plugin plugin : plugins) {
                System.out.println("플러그인 실행: " + plugin.name());
                plugin.execute();
            }
            Thread.sleep(10_000L);
        }
    }
}
```

이 구조를 컨테이너에서 돌린다고 생각해 봅시다.

udocker로 컨테이너를 띄울 때, 대략 이런 식의 명령이 들어갈 것입니다.
(실제 옵션 이름은 예시입니다. 개념만 보시면 됩니다.)

```bash
udocker run \
  -v /host/plugins:/opt/app/plugins \
  myapp-image:1.0
```

컨테이너 안 JVM 입장에서는 `/opt/app/plugins` 아래의 JAR이 언제 생기는지 모릅니다.
그냥 10초마다 디렉터리를 스캔할 뿐입니다.

이제 호스트에서 `/host/plugins`에 새로운 `my-plugin.jar`를 복사하면,
컨테이너 안에서는 `/opt/app/plugins/my-plugin.jar`로 보이게 되고,
10초 뒤에 다시 스캔할 때 그 JAR을 발견해서 로딩합니다.

이게 진짜 의미의 “이미 돌아가고 있는 컨테이너에 JAR를 꽂아서 plug & play”입니다.

컨테이너는

* 프로세스를 죽이지 않았고
* 컨테이너를 재생성하지도 않았으며
* 이미지를 재빌드하지도 않았습니다.

변한 건

* 플러그인 디렉터리에 JAR이 추가되었다는 사실과
* 애플리케이션이 그렇게 추가된 JAR을 반영하도록 설계되어 있다는 것뿐입니다.

---

여기서 핵심 요약을 다시 명확히 하자면 이렇습니다.

첫째, “컨테이너 런타임(udocker 포함)”만으로는 plug & play는 안 됩니다.
컨테이너는 프로세스 외부에 있고, JVM 내부의 classloader를 몰래 바꿀 수 없습니다.

둘째, “애플리케이션이 플러그인 시스템을 갖고 있다”면,
컨테이너든 udocker든 상관없이, 실행 중에 JAR을 추가하는 plug & play는 충분히 구현할 수 있습니다.
이때 컨테이너의 역할은 플러그인 디렉터리를 어떻게 마운트하고,
어디서 JAR을 공급받을지만 정해주는 수준입니다.

셋째, native-image + musl은 이런 동적 JAR 로딩에 매우 약합니다.
closed-world 제약 때문에, 이런 URLClassLoader 기반 플러그인 시스템이 사실상 막히거나,
엄청 복잡한 설정이 필요합니다.
반대로 JVM 위에 컨테이너(udocker든 Docker든)를 얹는 구조는 이런 플러그인 구조를 매우 자연스럽게 지원합니다.

그래서 결국 “이미 띄워진 컨테이너에 JAR을 꽂을 수 있느냐?”라는 질문의 핵심은,
“컨테이너냐 udocker냐”가 아니라
“애플리케이션이 플러그인 구조를 갖고 있느냐”입니다.

그 구조만 갖추면,

* OCI 이미지 + udocker
* 혹은 plain java -jar
* 혹은 Docker / Podman / Kubernetes

어떤 형태로 배포해도, 같은 plug & play 스토리를 유지할 수 있습니다.
