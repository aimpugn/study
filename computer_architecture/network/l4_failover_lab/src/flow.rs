//! 한 연결을 붙들고 한 번씩 왕복하며 결과를 분류합니다.
//!
//! 측정 도구마다 복사돼 있던 로직을 여기 한 곳에 모았습니다.

use std::io::{BufRead, BufReader, ErrorKind, Write};
use std::net::{SocketAddr, TcpStream};
use std::time::{Duration, Instant};

/// 한 번 왕복한 결과.
///
/// 실패를 뭉뚱그리지 않고 갈래로 나눈 이유가 있습니다. RST 와 타임아웃은 원인이
/// 전혀 다릅니다. RST 는 VIP 를 가진 장비가 살아 있는데 이 연결을 모른다는 뜻이고,
/// 타임아웃은 VIP 를 가진 장비가 아예 없다는 뜻입니다. 이 구분이 실험 1 과 실험 4 를
/// 가르는 관측입니다.
pub enum Outcome {
    /// 응답을 받았습니다.
    Reply {
        /// 백엔드가 붙여 준 이름 (`[B3]` 의 `B3`).
        backend: String,
        /// 보내고 받기까지 걸린 시간(초).
        secs: f64,
    },
    /// 상대가 연결을 정상 종료했습니다. 백엔드 프로세스가 죽으면 이렇게 됩니다.
    Fin {
        /// 종료를 확인하기까지 걸린 시간(초).
        secs: f64,
    },
    /// 상대가 RST 를 보냈습니다.
    Reset {
        /// RST 를 받기까지 걸린 시간(초).
        secs: f64,
    },
    /// 응답이 아예 없습니다.
    Timeout {
        /// 기다린 시간(초).
        secs: f64,
    },
    /// 위 어디에도 들어가지 않는 입출력 오류.
    Other {
        /// `std::io::ErrorKind` 를 그대로 찍은 문자열.
        kind: String,
        /// 오류가 날 때까지 걸린 시간(초).
        secs: f64,
    },
}

impl Outcome {
    /// 응답을 받았다면 그 백엔드 이름을 돌려줍니다.
    ///
    /// # Returns
    /// 성공했으면 `Some(백엔드 이름)`, 실패했으면 `None`.
    /// 흐름이 원래 백엔드를 유지했는지 셀 때 이 값을 비교합니다.
    pub fn backend(&self) -> Option<&str> {
        match self {
            Outcome::Reply { backend, .. } => Some(backend),
            _ => None,
        }
    }

    /// 사람이 읽을 한 줄로 만듭니다.
    ///
    /// # Returns
    /// `OK [B3] (0.00s)` 또는 `RST — 상대가 연결을 재설정했습니다 (0.00s)` 같은 문자열.
    pub fn label(&self) -> String {
        match self {
            Outcome::Reply { backend, secs } => format!("OK [{backend}] ({secs:.2}s)"),
            Outcome::Fin { secs } => format!("FIN — 상대가 연결을 닫았습니다 ({secs:.2}s)"),
            Outcome::Reset { secs } => format!("RST — 상대가 연결을 재설정했습니다 ({secs:.2}s)"),
            Outcome::Timeout { secs } => format!("타임아웃 — 응답이 없습니다 ({secs:.2}s)"),
            Outcome::Other { kind, secs } => format!("{kind} ({secs:.2}s)"),
        }
    }

    /// 실패했을 때 그 신호가 무엇을 뜻하는지 설명합니다.
    ///
    /// # Returns
    /// 설명할 것이 있으면 `Some(설명)`, 성공했거나 설명할 것이 없으면 `None`.
    pub fn hint(&self) -> Option<&'static str> {
        match self {
            Outcome::Reset { .. } => Some(
                "VIP를 가진 장비는 살아 있는데 이 연결을 모릅니다. \
                 상태를 못 넘겨받았거나 백엔드를 다르게 골랐다는 뜻입니다.",
            ),
            Outcome::Timeout { .. } => Some(
                "VIP를 가진 장비가 아무도 없어 패킷이 사라지고 있습니다. \
                 소켓은 아직 ESTABLISHED입니다.",
            ),
            Outcome::Fin { .. } => Some("상대 프로세스가 종료되면서 소켓이 닫혔습니다."),
            _ => None,
        }
    }
}

/// 하나의 TCP 연결.
///
/// TCP 는 바이트 스트림이라 메시지 경계를 지켜 주지 않습니다. 보내는 쪽이 한 번에 쓴
/// `[B3] hello\n` 이 받는 쪽에서 `[` 와 `B3] hello\n` 으로 쪼개져 도착할 수 있습니다.
/// 그래서 `read` 를 한 번 부르는 대신, 개행이 나올 때까지 읽는 `BufReader` 를
/// 연결마다 하나씩 붙여 둡니다.
pub struct Conn {
    /// 쓰기용 핸들. 타임아웃 설정도 이쪽에 겁니다.
    writer: TcpStream,
    /// 읽기용 핸들. 남은 바이트를 들고 있어야 하므로 왕복마다 새로 만들지 않습니다.
    reader: BufReader<TcpStream>,
}

impl Conn {
    /// 주소에 접속해 연결 하나를 만듭니다.
    ///
    /// # Arguments
    /// * `addr` — `10.0.1.100:8080` 형식의 접속 대상.
    ///
    /// # Returns
    /// 읽기·쓰기 핸들이 준비된 [`Conn`].
    ///
    /// # Errors
    /// 접속에 실패하거나 소켓 핸들 복제에 실패하면 그 오류를 돌려줍니다.
    pub fn connect(addr: &str) -> std::io::Result<Conn> {
        let writer = TcpStream::connect(addr)?;
        // try_clone 은 새 연결을 만들지 않습니다. 내부적으로 dup 이라 같은 파일
        // 디스크립션을 가리키고, 그래서 타임아웃 설정이 양쪽에 함께 적용됩니다.
        let reader = BufReader::new(writer.try_clone()?);
        Ok(Conn { writer, reader })
    }

    /// 이 연결의 내 쪽 주소.
    ///
    /// # Errors
    /// 소켓에서 주소를 읽지 못하면 그 오류를 돌려줍니다.
    pub fn local(&self) -> std::io::Result<SocketAddr> {
        self.writer.local_addr()
    }

    /// 이 연결의 상대 주소.
    ///
    /// 로드밸런서를 거쳐 접속했어도 여기 찍히는 값은 백엔드의 실제 주소가 아니라
    /// VIP 다. TCP 연결은 네 값으로 식별되고 그 값은 끝까지 바뀌지 않기 때문입니다.
    ///
    /// # Errors
    /// 소켓에서 주소를 읽지 못하면 그 오류를 돌려줍니다.
    pub fn peer(&self) -> std::io::Result<SocketAddr> {
        self.writer.peer_addr()
    }

    /// 한 줄을 보내고 한 줄을 받습니다.
    ///
    /// # Arguments
    /// * `message` — 보낼 내용. 개행은 이 함수가 붙입니다.
    /// * `timeout` — 응답을 기다릴 시간.
    ///
    /// # Returns
    /// 성공·FIN·RST·타임아웃·기타 중 하나. 이 함수는 실패해도 `Err` 를 돌려주지 않습니다.
    /// 실패 자체가 측정 대상이라 [`Outcome`] 안에 담아 돌려줍니다.
    pub fn roundtrip(&mut self, message: &str, timeout: Duration) -> Outcome {
        // 응답이 영영 안 올 수 있으므로 기다릴 상한을 먼저 겁니다.
        // 이 상한이 없으면 VIP 를 아무도 안 들고 있을 때 측정이 멈춰 버립니다.
        self.writer.set_read_timeout(Some(timeout)).ok();
        let start = Instant::now();

        // 보내기부터 실패하면 그 자체가 결과이므로 분류해서 돌려줍니다.
        if let Err(err) = writeln!(self.writer, "{message}") {
            return from_error(&err, start.elapsed().as_secs_f64());
        }

        // 개행이 나올 때까지 읽습니다. 한 번의 read 로는 응답이 잘려 올 수 있습니다.
        let mut line = String::new();
        match self.reader.read_line(&mut line) {
            Ok(0) => Outcome::Fin { secs: start.elapsed().as_secs_f64() },
            Ok(_) => Outcome::Reply {
                backend: parse_backend(&line),
                secs: start.elapsed().as_secs_f64(),
            },
            Err(err) => from_error(&err, start.elapsed().as_secs_f64()),
        }
    }
}

/// 입출력 오류를 [`Outcome`] 갈래로 옮깁니다.
///
/// # Arguments
/// * `err` — 읽기 또는 쓰기에서 나온 오류.
/// * `secs` — 그 시점까지 걸린 시간(초).
///
/// # Returns
/// 오류 종류에 맞는 [`Outcome`].
fn from_error(err: &std::io::Error, secs: f64) -> Outcome {
    match err.kind() {
        // 읽기 타임아웃은 리눅스에서 EAGAIN 으로 오므로 WouldBlock 으로 보입니다.
        ErrorKind::WouldBlock | ErrorKind::TimedOut => Outcome::Timeout { secs },
        ErrorKind::ConnectionReset => Outcome::Reset { secs },
        ErrorKind::BrokenPipe => Outcome::Fin { secs },
        other => Outcome::Other { kind: format!("{other:?}"), secs },
    }
}

/// 응답에서 백엔드 이름을 꺼냅니다.
///
/// # Arguments
/// * `reply` — 백엔드가 보낸 한 줄. `[B3] hello` 형태를 기대합니다.
///
/// # Returns
/// 대괄호 안의 이름. 형식이 맞지 않으면 `"?"`.
fn parse_backend(reply: &str) -> String {
    reply
        .split_once('[')
        .and_then(|(_, rest)| rest.split_once(']'))
        .map(|(name, _)| name.to_string())
        .unwrap_or_else(|| "?".to_string())
}
