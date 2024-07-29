# commands

- [commands](#commands)
    - [brew commands?](#brew-commands)
    - [commands list](#commands-list)
    - [버전 관리](#버전-관리)

## brew commands?

Show lists of built-in and external commands.

- `-q`, `--quiet`

    List only the names of commands without category headers.

- `--include-aliases`

    Include aliases of internal commands.

## commands list

1. **analytics**: Homebrew의 사용 데이터를 Google Analytics에 보내는 것을 관리합니다.
2. **autoremove**: 더 이상 필요하지 않은 종속 패키지를 자동으로 제거합니다.
3. **casks**: 설치 가능하거나 관리 중인 cask 목록을 보여줍니다.
4. **cleanup**: 오래된 버전의 패키지와 캐시를 정리합니다.
5. **commands**: 사용 가능한 모든 Homebrew 명령어를 보여줍니다.
6. **completions**: Homebrew의 쉘 완성 스크립트를 관리합니다.
7. **config**: Homebrew와 시스템의 구성 정보를 보여줍니다.
8. **deps**: 패키지의 종속성을 보여줍니다.
9. **desc**: 패키지의 설명을 보여줍니다.
10. **developer**: Homebrew 개발자 모드를 관리합니다.
11. **docs**: Homebrew 문서를 보여줍니다.
12. **doctor**: 시스템의 문제를 진단하고 일반적인 문제에 대한 해결책을 제공합니다.
13. **fetch**: 패키지를 다운로드하지만 설치하지는 않습니다.
14. **formulae**: Homebrew에서 관리하는 formulae 목록을 보여줍니다.
15. **gist-logs**: 문제가 있는 패키지의 로그를 GitHub Gist에 업로드합니다.
16. **help**: 특정 명령어의 사용법을 보여줍니다.
17. **home**: 패키지의 홈페이지를 기본 웹 브라우저에서 열어줍니다.
18. **info**: 패키지의 상세 정보를 보여줍니다.
19. **install**: 패키지를 설치합니다.
20. **leaves**: 다른 패키지에 의해 필요로 하지 않는, 즉 최상위 종속성 패키지들을 보여줍니다.
21. **link**: 설치된 패키지를 사용자의 시스템에 연결합니다.
22. **list**: 설치된 패키지 목록을 보여줍니다.
23. **log**: Homebrew의 이벤트 로그를 보여줍니다.
24. **migrate**: 변경된 formula 이름에 따라 설치된 패키지를 업데이트합니다.
25. **missing**: 설치된 패키지에서 누락된 종속성을 보여줍니다.
26. **nodenv-sync, pyenv-sync, rbenv-sync**: 각각 Node, Python, Ruby 버전 관리 도구와의 동기화를 돕습니다.
27. **options**: 특정 패키지의 설치 옵션을 보여줍니다.
28. **outdated**: 업그레이드 가능한 설치된 패키지를 보여줍니다.
29. **pin**: 패키지의 현재 버전을 고정하여 자동 업그레이드되지 않도록 합니다.
30. **postinstall**: 패키지 설치 후 실행되는 스크립트를 수동으로 실행합니다.
31. **postgresql-upgrade-database**: PostgreSQL 데이터베이스를 업그레이드합니다.
32. **reinstall**: 패키지를 재설치합니다.
33. **search**: 패키지를 검색합니다.
34. **setup-ruby**: Homebrew가 사용하는 내부 Ruby 환경을 설정합니다.
35. **shellenv**: Homebrew의 환경 변수를 초기화합니다.
36. **tap**: 추가 패키지 저장소를 추가하거나 관리합니다.
37. **tap-info**: 특정 tap에 대한 정보를 보여줍니다.
38. **uninstall**: 패키지를 제거합니다.
39. **unlink**: 패키지의 링크를 제거합니다.
40. **unpin**: 고정된 패키지의 고정을 해제합니다.
41. **untap**: 추가한 tap을 제거합니다.
42. **update**: Homebrew를 최신 상태로 업데이트합니다.
43. **update-report**: 마지막 `update` 명령의 결과를 보여줍니다.
44. **update-reset**: Homebrew 저장소를 기본 상태로 리셋합니다.
45. **upgrade**: 설치된 패키지를 최신 버전으로 업그레이드합니다.
46. **uses**: 특정 패키지를 종속성으로 사용하는 다른 패키지를 보여줍니다.
47. **vendor-install**: Homebrew가 필요로 하는 외부 소프트웨어를 설치합니다.

## 버전 관리

특정 패키지의 여러 버전을 관리하려면 일반적으로 해당 패키지에 대해 `@` 기호와 함께 버전 번호를 명시해야 합니다 (예: `node@14`). 일부 패키지의 경우, 여러 버전이 동시에 공존할 수 있도록 설계되어 있습니다. 그러나 Homebrew 자체에서는 다른 버전 관리 도구처럼 명령어로 버전을 전환하는 기능을 제공하지 않습니다. 대신, 특정 버전을 설치하고 `link` 및 `unlink` 명령어를 사용하여 활성화하거나 비활성화할 수 있습니다.
