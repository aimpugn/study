# systemctl

## `reload PATTERN...`

Asks all units listed on the command line to reload their configuration. Note that this will reload the service-specific configuration, not the unit configuration file of `systemd`. If you want `systemd` to reload the configuration file of a unit, use the daemon-reload command. In other words: for the example case of Apache, this will reload Apache's httpd.conf in the web server, not the apache.service `systemd` unit file.

This command should not be confused with the daemon-reload command.

명령줄에 나열된 모든 장치에 구성을 다시 로드하도록 요청합니다.
이 경우 `systemd`의 단위 구성 파일이 아니라 *서비스별 구성을 다시 로드*한다는 점에 유의하세요.
`systemd`가 단위의 구성 파일을 다시 로드하도록 하려면 `daemon-reload` 명령을 사용합니다.

즉, 예를 들어 Apache의 경우, 이 명령은 Apache의 httpd.conf를 다시 로드하는 것이고,
apache.service systemd 유닛 파일을 다시 로즈하지 않습니다.

이 명령을 daemon-reload 명령과 혼동해서는 안 됩니다.

## Manager State Command

### `daemon-reload`

Reload the systemd manager configuration. This will rerun all generators (see systemd.generator(7)), reload all unit files, and recreate the entire dependency tree. While the daemon is being reloaded, all sockets systemd listens on behalf of user configuration will stay accessible.

This command should not be confused with the reload command.

systemd manager의 구성을 다시 로드합니다.
- 모든 generator가 다시 실행되고(`systemd.generator(7)` 참조)
- 모든 단위 파일이 다시 로드되며
- 전체 종속성 트리가 다시 생성

데몬이 다시 로드되는 동안 `systemd`가 사용자 구성을 대신하여 수신 대기(listen)하는 모든 소켓은 계속 액세스할 수 있습니다.

이 명령을 `reload` 명령과 혼동해서는 안 됩니다.

즉, 파일 시스템에서 변경된 구성을 가져와 종속성 트리를 다시 생성하는 soft reload입니다.

- [What does "systemctl daemon-reload" do?](https://unix.stackexchange.com/a/364787)
