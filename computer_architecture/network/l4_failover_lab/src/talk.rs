//! 손으로 두드리는 클라이언트.

use crate::flow::{Conn, Outcome};
use clap::Args;
use std::io::{BufRead, Write};
use std::time::Duration;

/// `talk` 명령의 인자.
#[derive(Args)]
pub struct Options {
    /// 접속할 주소. `10.0.1.100:8080` 형식.
    #[arg(long)]
    pub to: String,
    /// 응답을 기다릴 시간(초).
    #[arg(long, default_value_t = 10)]
    pub timeout: u64,
}

/// 연결을 하나 붙들고, 사용자가 친 줄을 그대로 보내며 결과를 보고합니다.
///
/// 순서대로 이렇게 돕니다.
/// 1. 접속하고 내 주소와 상대 주소를 찍습니다. 상대는 VIP 로 찍힙니다.
/// 2. 한 줄을 읽어 보내고 결과를 찍습니다. 실패면 그 신호가 무엇을 뜻하는지도 찍습니다.
/// 3. 타임아웃은 연결이 아직 살아 있을 수 있으므로 계속 두드려 볼 수 있게 둡니다.
///    RST 나 FIN 은 이 소켓으로 더 할 수 있는 일이 없으므로 끝냅니다.
///
/// # Arguments
/// * `options` — 접속 주소와 응답 대기 시간.
///
/// # Returns
/// 사용자가 입력을 끝냈거나(Ctrl-D) 연결이 회복 불가능해지면 `Ok(())`.
///
/// # Errors
/// 접속에 실패하거나 표준 입출력이 실패하면 그 오류를 돌려줍니다.
pub fn run(options: &Options) -> std::io::Result<()> {
    let timeout = Duration::from_secs(options.timeout);
    let mut conn = Conn::connect(&options.to)?;
    println!(
        "연결됨.  내 주소={}   소켓이 기억하는 상대={}",
        conn.local()?,
        conn.peer()?
    );
    println!("아무 글자나 치고 엔터를 누르세요. 종료는 Ctrl-C 또는 Ctrl-D.\n");

    let stdin = std::io::stdin();
    loop {
        print!("> ");
        std::io::stdout().flush()?;
        let mut line = String::new();
        if stdin.lock().read_line(&mut line)? == 0 {
            return Ok(());
        }
        let line = line.trim();
        if line.is_empty() {
            continue;
        }
        let outcome = conn.roundtrip(line, timeout);
        println!("  {}", outcome.label());
        if let Some(hint) = outcome.hint() {
            println!("     {hint}");
        }
        if !matches!(outcome, Outcome::Reply { .. } | Outcome::Timeout { .. }) {
            return Ok(());
        }
    }
}
