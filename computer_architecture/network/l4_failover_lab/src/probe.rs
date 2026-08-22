//! 측정 도구. 흐름을 N개 열고, 단계마다 한 번씩 왕복시켜 결과를 보고합니다.
//!
//! 장애를 언제 어떻게 낼지는 이 모듈이 정하지 않습니다. 단계 사이에서 멈춰 셸이
//! 신호를 줄 때까지 기다리기만 합니다. 그래서 장애 종류가 늘어도 이 코드는 그대로입니다.

use crate::flow::{Conn, Outcome};
use clap::Args;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::thread::sleep;
use std::time::Duration;

/// `probe` 명령의 인자.
#[derive(Args)]
pub struct Options {
    /// 접속할 주소. `10.0.1.100:8080` 형식.
    #[arg(long)]
    pub to: String,
    /// 동시에 열 흐름 수.
    #[arg(long, default_value_t = 1)]
    pub flows: usize,
    /// 왕복을 몇 단계에 걸쳐 할지. 단계 사이마다 셸의 신호를 기다립니다.
    #[arg(long, default_value_t = 2)]
    pub stages: usize,
    /// 요청 순서를 섞는 난수 씨앗. 같은 값이면 같은 순서가 나옵니다.
    #[arg(long, default_value_t = 1)]
    pub seed: u64,
    /// 응답을 기다릴 시간(초).
    #[arg(long, default_value_t = 8)]
    pub timeout: u64,
    /// 단계 사이 신호를 주고받을 디렉터리.
    #[arg(long, default_value = "/work/.gates")]
    pub gates: PathBuf,
    /// 이 단계에서는 기존 연결을 재사용하지 않고 새로 맺습니다.
    #[arg(long)]
    pub new_at: Option<usize>,
}

/// 측정을 처음부터 끝까지 진행합니다.
///
/// # Arguments
/// * `options` — 접속 대상, 흐름 수, 단계 수, 씨앗, 타임아웃, 신호 디렉터리.
///
/// # Returns
/// 모든 단계를 마치면 `Ok(())`. 왕복이 실패해도 그것은 측정 결과이지 오류가 아니다.
///
/// # Errors
/// 신호 디렉터리를 만들지 못하거나, 접속에 실패하거나, 기준선 파일 쓰기가
/// 실패하면 그 오류를 돌려줍니다.
pub fn run(options: &Options) -> std::io::Result<()> {
    // 셸과 신호를 주고받을 자리를 먼저 확보합니다. 이 디렉터리가 없으면
    // 단계 사이에서 서로를 기다리지 못합니다.
    let timeout = Duration::from_secs(options.timeout);
    fs::create_dir_all(&options.gates)?;

    // 측정할 흐름을 한꺼번에 열고, 클라이언트가 무엇을 상대로 인식하는지 남깁니다.
    // 로드밸런서를 거쳐도 여기 찍히는 상대는 백엔드가 아니라 VIP 입니다.
    let mut streams = open_flows(&options.to, options.flows)?;
    say(&format!(
        "연결: local={}  peer(소켓이 기억하는 상대)={}",
        streams[0].local()?,
        streams[0].peer()?
    ));

    let mut origin: Vec<String> = Vec::new();
    for stage in 1..=options.stages {
        // 2단계부터는 셸이 장애를 낼 시간을 주기 위해 멈춰 기다립니다.
        if stage > 1 {
            wait_gate(&options.gates, stage - 1);
        }

        // 지정된 단계에서는 기존 연결 대신 새 연결로 잽니다. 죽은 백엔드를 걷어낸 뒤
        // 새 연결은 정상 백엔드로 가는지를 같은 시점에 확인하기 위해서입니다.
        let fresh = options.new_at == Some(stage);
        if fresh {
            streams = open_flows(&options.to, options.flows)?;
        }

        // 순서를 섞어 왕복시킵니다. 실제 트래픽은 연결을 만든 순서대로 도착하지 않는데,
        // 섞지 않으면 라운드로빈이 생성 순서 그대로 재배분해 우연히 원래 백엔드를 맞춥니다.
        let order = shuffled(options.flows, options.seed.wrapping_add(stage as u64));
        let mut slots: Vec<Option<Outcome>> = (0..options.flows).map(|_| None).collect();
        for &index in &order {
            slots[index] = Some(streams[index].roundtrip(&format!("s{stage}"), timeout));
        }
        let results: Vec<Outcome> = slots.into_iter().flatten().collect();

        // 1단계 결과를 기준선으로 남깁니다. 이후 단계는 이것과 비교해 유지 여부를 세고,
        // 셸도 이 파일을 읽어 어느 백엔드를 죽일지 정합니다.
        if stage == 1 {
            origin = results
                .iter()
                .map(|result| result.backend().unwrap_or("?").to_string())
                .collect();
            fs::write(options.gates.join("backends"), origin.join(" "))?;
        }

        report(stage, options.flows, &origin, &results, fresh);
    }
    Ok(())
}

/// 같은 대상으로 흐름을 여러 개 엽니다.
///
/// # Arguments
/// * `target` — 접속할 주소.
/// * `count` — 열 흐름 수.
///
/// # Returns
/// 접속이 끝난 연결 목록.
///
/// # Errors
/// 하나라도 접속에 실패하면 그 오류를 돌려줍니다.
fn open_flows(target: &str, count: usize) -> std::io::Result<Vec<Conn>> {
    let mut streams = Vec::with_capacity(count);
    for _ in 0..count {
        streams.push(Conn::connect(target)?);
    }
    Ok(streams)
}

/// 한 단계의 결과를 사람이 읽을 형태로 찍습니다.
///
/// 흐름이 하나면 결과 한 줄과 그 신호의 뜻을 찍습니다.
/// 여럿이면 1단계는 백엔드 배분을, 이후 단계는 원래 백엔드를 유지한 비율과
/// 흐름별 내역을 찍습니다.
///
/// # Arguments
/// * `stage` — 지금 몇 단계인지(1부터).
/// * `flows` — 흐름 수.
/// * `origin` — 1단계에서 각 흐름이 붙었던 백엔드 이름.
/// * `results` — 이번 단계의 왕복 결과.
/// * `fresh` — 이번 단계에서 연결을 새로 맺었는지. 그렇다면 유지 여부를 따지지 않습니다.
fn report(stage: usize, flows: usize, origin: &[String], results: &[Outcome], fresh: bool) {
    if flows == 1 {
        let tag = if fresh { " (새 연결)" } else { "" };
        say(&format!("{stage}단계{tag}: {}", results[0].label()));
        if let Some(hint) = results[0].hint() {
            say(&format!("        {hint}"));
        }
        return;
    }
    if stage == 1 {
        say(&format!("{stage}단계 배분: {}", origin.join(" ")));
        return;
    }
    if fresh {
        let ok = results.iter().filter(|r| r.backend().is_some()).count();
        say(&format!("{stage}단계 (새 연결): 성공 {ok}/{flows}"));
        return;
    }
    let kept = results
        .iter()
        .zip(origin)
        .filter(|(result, was)| result.backend() == Some(was.as_str()))
        .count();
    say(&format!("{stage}단계: 기존 백엔드 유지 {kept}/{flows}"));
    for (result, was) in results.iter().zip(origin) {
        let mark = match result.backend() {
            Some(now) if now == was => "유지".to_string(),
            Some(now) => format!("→ {now} (다른 백엔드)"),
            None => String::new(),
        };
        say(&format!("    {was}: {} {mark}", result.label()));
    }
}

/// 셸이 다음 단계를 허락할 때까지 기다립니다.
///
/// `ready.<index>` 파일을 만들어 "여기까지 왔다"고 알리고,
/// `go.<index>` 파일이 생길 때까지 기다립니다.
///
/// # Arguments
/// * `dir` — 신호 파일을 주고받을 디렉터리.
/// * `index` — 몇 번째 관문인지.
fn wait_gate(dir: &Path, index: usize) {
    fs::File::create(dir.join(format!("ready.{index}"))).ok();
    let go = dir.join(format!("go.{index}"));
    while !go.exists() {
        sleep(Duration::from_millis(50));
    }
}

/// 0부터 `len-1` 까지를 섞은 순서를 만듭니다.
///
/// 같은 씨앗이면 같은 순서가 나옵니다. 실제 트래픽은 연결을 만든 순서대로 도착하지
/// 않으므로, 섞지 않으면 라운드로빈이 생성 순서 그대로 다시 배분하면서 우연히
/// 원래 백엔드를 맞춰 버립니다.
///
/// # Arguments
/// * `len` — 섞을 원소 수.
/// * `seed` — 난수 씨앗. 0 이면 고정된 다른 값으로 바꿉니다. xorshift 는 0 에서
///   벗어나지 못하기 때문입니다.
///
/// # Returns
/// `0..len` 을 섞은 색인 목록.
fn shuffled(len: usize, seed: u64) -> Vec<usize> {
    let mut state = if seed == 0 { 0x9E37_79B9_7F4A_7C15 } else { seed };
    let mut next = move || {
        state ^= state << 13;
        state ^= state >> 7;
        state ^= state << 17;
        state
    };
    let mut order: Vec<usize> = (0..len).collect();
    for i in (1..len).rev() {
        order.swap(i, (next() % (i as u64 + 1)) as usize);
    }
    order
}

/// 한 줄을 찍고 바로 흘려보냅니다.
///
/// 출력이 파일로 리다이렉트되면 버퍼링되므로, 셸이 중간 결과를 제때 보려면
/// 단계마다 flush 해야 합니다.
///
/// # Arguments
/// * `line` — 찍을 내용.
fn say(line: &str) {
    println!("{line}");
    std::io::stdout().flush().ok();
}
