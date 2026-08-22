//! 백엔드 서버.

use clap::Args;
use std::io::{BufRead, BufReader, Write};
use std::net::{TcpListener, TcpStream};
use std::thread;

/// `serve` 명령의 인자.
#[derive(Args)]
pub struct Options {
    /// 이 백엔드의 이름. 응답 앞에 `[B1]` 처럼 붙습니다.
    #[arg(long, default_value = "B?")]
    pub name: String,
    /// 접속을 받을 포트.
    #[arg(long, default_value_t = 8080)]
    pub port: u16,
}

/// 백엔드 서버를 띄우고 접속을 계속 받습니다.
///
/// 하는 일은 셋입니다.
/// 1. 지정한 포트에 바인딩합니다.
/// 2. 접속이 올 때마다 "내가 보는 상대"를 표준 출력에 찍습니다.
///    이 한 줄이 패킷 모드와 프록시 모드를 가릅니다. 패킷 모드면 클라이언트의 실제
///    주소가, 프록시 모드면 로드밸런서의 주소가 찍힙니다.
/// 3. 연결마다 스레드를 하나 붙여 [`echo`] 에 넘깁니다. 한 연결이 멈춰도 나머지는
///    계속 받기 위해서입니다.
///
/// # Arguments
/// * `options` — 백엔드 이름과 수신 포트.
///
/// # Returns
/// 정상적으로 끝나는 경로가 없습니다. 반환했다면 접속 수락이 실패한 것입니다.
///
/// # Errors
/// 포트 바인딩에 실패하거나, 접속 수락 또는 상대 주소 조회가 실패하면 그 오류를 돌려줍니다.
pub fn run(options: &Options) -> std::io::Result<()> {
    let listener = TcpListener::bind(("0.0.0.0", options.port))?;
    for incoming in listener.incoming() {
        let stream = incoming?;
        println!("백엔드가 본 상대: {}", stream.peer_addr()?);
        std::io::stdout().flush()?;
        let name = options.name.clone();
        // move 로 소켓 소유권을 스레드에 넘깁니다. 넘기지 않으면 이 반복문이
        // 다음 연결로 넘어갈 때 소켓이 사라집니다.
        thread::spawn(move || echo(&name, stream));
    }
    Ok(())
}

/// 한 연결을 맡아 줄 단위로 에코합니다.
///
/// 받은 줄 앞에 이름을 붙여 돌려줍니다. 상대가 연결을 닫거나 쓰기가 실패하면 끝냅니다.
///
/// # Arguments
/// * `name` — 응답 앞에 붙일 이름.
/// * `stream` — 이 연결의 소켓. 소유권을 가져갑니다.
fn echo(name: &str, stream: TcpStream) {
    // 같은 소켓을 읽기용과 쓰기용으로 나눠 씁니다. try_clone 은 새 연결을 만들지 않고
    // 같은 연결을 가리키는 핸들을 하나 더 만들 뿐입니다.
    let Ok(mut out) = stream.try_clone() else { return };
    for line in BufReader::new(stream).lines() {
        let Ok(line) = line else { return };
        // writeln! 은 포맷 조각마다 write 를 부릅니다. 그러면 `[B3] hello` 가 여러
        // TCP 세그먼트로 쪼개져 나가고, 받는 쪽이 read 를 한 번만 부르면 `[` 만
        // 받는 일이 생깁니다. 한 줄을 한 번에 보내도록 미리 만들어 둡니다.
        if out.write_all(format!("[{name}] {line}\n").as_bytes()).is_err() {
            return;
        }
    }
}
