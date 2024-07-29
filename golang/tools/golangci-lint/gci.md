# gci

- [gci](#gci)
    - [gci?](#gci-1)
    - [goimports vs gci](#goimports-vs-gci)
        - [주요 차이점](#주요-차이점)
    - [결론](#결론)

## [gci](https://github.com/daixiang0/gci)?

Go Control Import

> GCI, a tool that controls Go package import order and makes it always deterministic.
>
> The desired output format is highly configurable and allows for more custom formatting than goimport does.

원하는 출력 형식을 고도로 구성할 수 있으며, `goimport`보다 더 많은 사용자 지정 형식을 허용합니다.

## goimports vs gci

- goimports

    **기본 기능**:
    - 필요한 패키지를 자동으로 추가하고 사용하지 않는 패키지를 제거합니다.
    - 기본적인 임포트 정렬을 수행합니다.

    **동작 방식**:
    - 코드를 분석하여 필요한 패키지를 결정합니다.
    - 표준 라이브러리, 외부 패키지, 로컬 패키지 순으로 정렬합니다.

    **장점**:
    - Go 표준 도구의 일부로, 널리 사용되고 신뢰할 수 있습니다.
    - 대부분의 기본적인 정렬 요구사항을 충족합니다.

    **제한사항**:
    - 세부적인 정렬 규칙을 커스터마이즈하기 어렵습니다.

- gci (Go Code Import)

    **기본 기능**:
    - 임포트 문을 정렬하고 그룹화합니다.
    - 더 세밀한 정렬 규칙을 제공합니다.

    **동작 방식**:
    - 사용자 정의 규칙에 따라 임포트를 그룹화하고 정렬합니다.
    - 여러 섹션으로 임포트를 나눌 수 있습니다.

    **장점**:
    - 높은 수준의 커스터마이제이션이 가능합니다.
    - 프로젝트 특정 요구사항에 맞는 임포트 스타일을 적용할 수 있습니다.

    **특징**:
    - 표준 라이브러리, 외부 패키지, 내부 패키지 등을 세부적으로 구분할 수 있습니다.
    - 별칭(alias)이 있는 임포트도 잘 처리합니다.

### 주요 차이점

1. 기능 범위:

    `goimports`:
    - 기본적인 정렬: 표준 라이브러리, 외부 패키지, 내부 패키지 순으로 정렬합니다.
    - 필요한 패키지 관리: 사용되는 패키지를 자동으로 추가하고, 사용되지 않는 패키지를 제거합니다.

    예시:

    ```go
    // 정렬 전
    import (
        "myproject/internal"
        "fmt"
        "github.com/external/pkg"
    )

    // goimports 적용 후
    import (
        "fmt"

        "github.com/external/pkg"

        "myproject/internal"
    )
    ```

    `gci`:
    - 세밀한 정렬 규칙: 사용자 정의 그룹을 만들고, 각 그룹 내에서 추가적인 정렬 규칙을 적용할 수 있습니다.
    - 복잡한 그룹화: 여러 계층의 그룹을 정의할 수 있습니다.

    예시:

    ```go
    // gci 적용 (사용자 정의 규칙)
    import (
        // 표준 라이브러리
        "fmt"
        "strings"

        // 외부 패키지
        "github.com/external/pkg1"
        "github.com/external/pkg2"

        // 내부 공통 패키지
        "mycompany/common/logger"
        "mycompany/common/utils"

        // 프로젝트 내부 패키지
        "myproject/internal/config"
        "myproject/internal/models"
    )
    ```

2. 커스터마이제이션:

    `goimports`:
    - 제한적인 옵션: 주로 `-local` 플래그를 사용하여 로컬 패키지의 범위를 지정할 수 있습니다.

    예시:

    ```sh
    goimports -local mycompany/myproject
    ```

    `gci`:
    - 높은 수준의 커스터마이제이션: 정렬 섹션, 그룹 순서, 프리픽스 매칭 등을 상세히 설정할 수 있습니다.

    예시 (gci 설정 파일):

    ```yaml
    sections:
        - standard
        - default
        - prefix(github.com/external)
        - prefix(mycompany)
        - prefix(myproject/internal)
    ```

3. 사용 사례:

    `goimports`:
    - 일반적인 Go 프로젝트: 대부분의 소규모~중규모 프로젝트에서 충분한 기능을 제공합니다.

    예시:

    ```go
    // 작은 프로젝트의 파일
    import (
        "fmt"
        "net/http"

        "github.com/gorilla/mux"

        "myproject/handlers"
    )
    ```

    `gci`:
    - 대규모 프로젝트: 복잡한 패키지 구조를 가진 대규모 프로젝트에서 유용합니다.

    예시:

    ```go
    // 대규모 프로젝트의 파일
    import (
        // 핵심 시스템
        "mycompany/core/auth"
        "mycompany/core/database"

        // 특정 서비스
        "mycompany/services/user"
        "mycompany/services/billing"

        // 외부 의존성
        "github.com/aws/aws-sdk-go/aws"
        "github.com/stripe/stripe-go"

        // 내부 유틸리티
        "mycompany/internal/logger"
        "mycompany/internal/metrics"
    )
    ```

4. 통합:

    `goimports`:
    - Go 도구 체인과 통합: 대부분의 IDE와 에디터에 기본적으로 통합되어 있습니다.

    예시 (VSCode 설정):

    ```json
    {
        "go.formatTool": "goimports",
        "editor.formatOnSave": true
    }
    ```

    `gci`:
    - 별도 설치 및 설정: 추가적인 설정이 필요하며, CI/CD 파이프라인에 통합하기 위해서는 추가 작업이 필요할 수 있습니다.

    예시 (CI/CD 파이프라인 스크립트):

    ```yaml
    - name: Install gci
        run: go install github.com/daixiang0/gci@latest

    - name: Run gci
        run: gci write -s standard -s default -s "prefix(github.com/)" -s "prefix(mycompany/)" .
    ```

## 결론

- `goimports`는 대부분의 기본적인 요구사항을 충족시키는 표준 도구입니다.
- `gci`는 더 세밀한 제어와 커스터마이제이션이 필요한 경우에 유용합니다.
- 프로젝트의 복잡성과 팀의 선호도에 따라 선택할 수 있습니다.
- 일부 프로젝트에서는 두 도구를 함께 사용하기도 합니다. 예를 들어, `goimports`로 기본 정리를 하고 `gci`로 추가적인 정렬을 적용할 수 있습니다.
