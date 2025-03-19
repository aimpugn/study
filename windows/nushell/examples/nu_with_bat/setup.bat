:: 명령어를 화면에 출력하지 않도록 설정g
@echo off
setlocal
:: Nushell 실행 파일 경로 찾기
set "NU_PATH=.\programs\nu\nu.exe"
:: install.nu 실행
"%NU_PATH%" install.nu
