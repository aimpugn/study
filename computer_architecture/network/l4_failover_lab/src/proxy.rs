//! 프록시 모드 최소 구현.
//!
//!   소켓 ①  클라이언트 ↔ 프록시        소켓 ②  프록시 ↔ 백엔드
//!
//! 이 프로그램이 양쪽 TCP 연결의 종단입니다. 그래서 백엔드는 클라이언트가 아니라
//! 이 프로그램의 주소를 봅니다. 패킷 모드와의 차이가 여기서 갈립니다.

use clap::Args;
use std::io;
use std::net::{Shutdown, TcpListener, TcpStream};
use std::thread;

/// `proxy` 명령의 인자.
#[derive(Args)]
pub struct Options {
    /// 접속을 받을 주소. `10.0.1.100:9091` 형식.
    #[arg(long)]
    pub listen: String,
    /// 중계해 보낼 백엔드 주소. `10.0.2.11:9099` 형식.
    #[arg(long)]
    pub to: String,
}

/// 프록시를 띄우고 접속을 계속 받습니다.
///
/// 접속이 올 때마다 스레드를 하나 붙여 [`relay`] 에 넘깁니다.
///
/// # Arguments
/// * `options` — 수신 주소와 백엔드 주소.
///
/// # Returns
/// 정상적으로 끝나는 경로가 없습니다. 반환했다면 접속 수락이 실패한 것입니다.
///
/// # Errors
/// 바인딩이나 접속 수락에 실패하면 그 오류를 돌려줍니다.
pub fn run(options: &Options) -> io::Result<()> {
    let listener = TcpListener::bind(&options.listen)?;
    for incoming in listener.incoming() {
        let client = incoming?;
        let backend = options.to.clone();
        // move 로 소켓 소유권을 스레드에 넘깁니다. 넘기지 않으면 이 반복문이
        // 다음 연결로 넘어갈 때 소켓이 사라집니다.
        thread::spawn(move || relay(client, &backend));
    }
    Ok(())
}

/// 클라이언트 연결 하나를 백엔드와 이어 줍니다.
///
/// # Arguments
/// * `client` — 이미 수락된 클라이언트 소켓. 소유권을 가져갑니다.
/// * `backend` — 접속할 백엔드 주소.
///
/// # Errors
/// 백엔드 접속이나 소켓 핸들 복제에 실패하면 그 오류를 돌려줍니다.
fn relay(client: TcpStream, backend: &str) -> io::Result<()> {
    // 백엔드로 별도의 TCP 연결을 맺습니다. 이 한 줄이 패킷 모드와의 차이를 만듭니다.
    // 이 시점부터 연결이 두 개이고, 백엔드가 보는 상대는 클라이언트가 아니라 이 프로그램입니다.
    let server = TcpStream::connect(backend)?;
    let (client_read, server_write) = (client.try_clone()?, server.try_clone()?);

    // 방향마다 스레드를 하나씩 씁니다. pump 안의 io::copy 가 상대가 끊을 때까지 막히므로,
    // 한 스레드로 양방향을 처리하면 한쪽이 조용할 때 반대쪽이 굶습니다.
    let upstream = thread::spawn(move || pump(client_read, server_write));
    pump(server, client);

    // 올라가는 방향이 끝날 때까지 기다렸다가 이 연결을 정리합니다.
    upstream.join().ok();
    Ok(())
}

/// 한 방향으로 바이트를 나릅니다.
///
/// # Arguments
/// * `from` — 읽을 소켓.
/// * `to` — 쓸 소켓.
///
/// 다 나른 뒤 `to` 의 쓰기 방향을 닫습니다. 이걸 빠뜨리면 상대가 EOF 를 받지 못해
/// 보낼 것이 없는데도 계속 기다립니다. TCP 의 반쪽 닫기(half-close)를 다루는 자리입니다.
fn pump(mut from: TcpStream, mut to: TcpStream) {
    io::copy(&mut from, &mut to).ok();
    to.shutdown(Shutdown::Write).ok();
}
