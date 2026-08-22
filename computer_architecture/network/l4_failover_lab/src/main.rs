// 문서 없는 항목이 있으면 컴파일을 실패시킵니다. 비공개 헬퍼까지 포함합니다.
// 규율에 기대지 않고 기계가 막게 하려는 것입니다.
#![deny(missing_docs)]
#![deny(clippy::missing_docs_in_private_items)]

//! L4 실험대의 측정 도구.
//!
//! 망을 세우고 부수는 일은 셸이 맡는다(`net/`, `scenario/`).
//! 이 프로그램은 서버 노릇을 하거나 결과를 재기만 합니다.

mod flow;
mod probe;
mod proxy;
mod serve;
mod talk;

use clap::{Parser, Subcommand};

/// 명령줄 전체.
#[derive(Parser)]
#[command(name = "l4lab", version, about = "L4 장애 전환 실험대의 측정 도구")]
struct Cli {
    /// 실행할 하위 명령.
    #[command(subcommand)]
    command: Command,
}

/// 이 도구가 할 수 있는 일.
///
/// 각 갈래의 문서 주석이 그대로 `--help` 출력이 됩니다.
/// 그래서 문서와 실제 동작이 갈라질 수 없습니다.
#[derive(Subcommand)]
enum Command {
    /// 백엔드 서버. 받은 줄에 자기 이름을 붙여 돌려주고, 접속할 때마다 상대 주소를 찍습니다.
    Serve(serve::Options),
    /// 프록시 모드 시연. 연결 두 개를 각각 맺고 바이트를 중계합니다.
    Proxy(proxy::Options),
    /// 측정. 흐름을 열고 단계마다 왕복시키며, 단계 사이에서 셸의 신호를 기다립니다.
    Probe(probe::Options),
    /// 손으로 두드리는 클라이언트.
    Talk(talk::Options),
}

/// 인자를 파싱해 해당 명령으로 넘깁니다.
///
/// # Returns
/// 명령이 정상 종료하면 `SUCCESS`, 입출력 오류로 끝나면 `FAILURE`.
/// 인자가 잘못된 경우는 clap 이 직접 메시지를 찍고 종료합니다.
fn main() -> std::process::ExitCode {
    let cli = Cli::parse();
    let result = match &cli.command {
        Command::Serve(options) => serve::run(options),
        Command::Proxy(options) => proxy::run(options),
        Command::Probe(options) => probe::run(options),
        Command::Talk(options) => talk::run(options),
    };
    match result {
        Ok(()) => std::process::ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("l4lab: {err}");
            std::process::ExitCode::FAILURE
        }
    }
}
