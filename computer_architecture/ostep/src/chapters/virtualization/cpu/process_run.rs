use std::collections::HashMap;

// struct: tuple 타입과 비슷하다
pub struct Scheduler<'a> {
    pub proc_info: HashMap<i32, ProcInfo>,
    pub process_switch_behavior: String,
    pub io_done_behavior: String,
    pub program: Vec<&'a str>,
    pub io_length: i32,
}

pub struct ProcInfo {
    proc_pc: i32,
    proc_id: i32,
    proc_code: Vec<ProcCode>,
    proc_state: ProcState,
}

enum ProcCode {
    DoCompute,
    DoIo,
    DoIoDone,
}
// enum: https://doc.rust-lang.org/book/ch06-01-defining-an-enum.html
enum ProcState {
    StateRunning,
    StateReady,
    StateDone,
    StateWait,
}

/// pub fn new_process() {}
/// pub fn load_program() {}
/// pub fn load() {}
/// pub fn move_to_ready() {}
/// pub fn move_to_wait() {}
/// pub fn move_to_running() {}
/// pub fn move_to_done() {}
/// pub fn next_proc() {}
/// pub fn get_num_processes() {}
/// pub fn get_num_instructions() {}
/// pub fn get_instruction() {}
/// pub fn get_num_active() {}
/// pub fn get_num_runnable() {}
/// pub fn get_ios_in_flight() {}
/// pub fn check_for_switch() {}
/// pub fn check_forspace_switch() {}
/// pub fn check_if_done() {}
/// pub fn run() {}
impl<'a> Scheduler<'a> {
    pub fn init(&self) {
        if self.program.len() > 0 {
            Self::load_program(self)
        } else {
            Self::load(self)
        }
    }

    /// program looks like this:
    ///
    ///   c7,i,c1,i
    ///
    /// which means
    ///   compute for 7, then i/o, then compute for 1, then i/o
    fn load_program(&self) {
        for p in self.program.iter() {
            for line in p.split(",") {
                let line_string = line.to_string();
                let opcode = line_string.get(0..1).unwrap();
                match opcode {
                    "c" => println!("opcode is c"),
                    "i" => println!("opcode is i"),
                    _ => println!("opcode is nothing"),
                }
            }
        }
    }

    fn load(&self) {}

    /// Create new ProcInfo and add it into proc_info HashMap
    pub fn new_process(&mut self) -> i32 {
        let proc_id = self.proc_info.len() as i32;
        let proc_info = ProcInfo {
            proc_pc: 0,
            proc_id: proc_id,
            proc_code: Vec::new(),
            proc_state: ProcState::StateReady,
        };
        self.proc_info.insert(proc_id, proc_info);
        proc_id
    }
    pub fn move_to_ready() {}
    pub fn move_to_wait() {}
    pub fn move_to_running() {}
    pub fn move_to_done() {}
    pub fn next_proc() {}
    pub fn get_num_processes() {}
    pub fn get_num_instructions() {}
    pub fn get_instruction() {}
    pub fn get_num_active() {}
    pub fn get_num_runnable() {}
    pub fn get_ios_in_flight() {}
    pub fn check_for_switch() {}
    pub fn check_forspace_switch() {}
    pub fn check_if_done() {}
    pub fn run(&self) {}
}
