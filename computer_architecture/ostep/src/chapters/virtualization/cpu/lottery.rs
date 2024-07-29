use nix::unistd::Whence;
// https://docs.rs/rand/latest/rand/
// https://github.com/rust-random/rand
use rand::Rng;
use std::borrow::Borrow;
use std::cell::{Ref, RefCell, RefMut};
use std::ops::Deref;
use std::rc::Rc;

pub fn simple_lottery_scheduling() {
    let total_tickets = 400;
    let mut rng = rand::thread_rng();
    let winner: i32 = rng.gen_range(0..=total_tickets);
    // instead of `let tickets = vec![100, 50, 250];`
    // sorted from highest to lowest
    // -> 일반적으로 가장 적은 수의 반복이 이뤄지도록 한다
    let tickets_desc = vec![250, 100, 50];

    let head = Some(Rc::new(RefCell::new(Process::new(tickets_desc[0]))));

    // When clone creates a new `Rc<RefCell<Process>>` that **points to the same** `RefCell<Process>`
    // But NOT create a new `RefCell<Process>` or `Process`
    let mut tail = head.clone();

    for &ticket in &tickets_desc[1..] {
        let next = Some(Rc::new(RefCell::new(Process::new(ticket))));
        tail // Option<Rc<RefCell<Process>>>
            .as_mut() // Option<&mut Rc<RefCell<Process>>>
            .unwrap() // &mut Rc<RefCell<Process>>
            .borrow_mut() // RefMut<Process>
            .next = next.clone();
        tail = next;
    }

    let mut counter = 0;
    let mut current = head;
    while let Some(node) = current {
        // create temporary reference
        let node_ref = node.as_ref(); // === `let node_ref = &*node;`
        println!("current tickets is {}", node_ref.borrow().tickets);
        counter += node_ref.borrow().tickets;
        if counter > winner {
            println!("counter is: {}, winner: {}", counter, winner);
            // found the winner
            break;
        }
        current = node.as_ref().borrow().next.clone();
    }
    println!("Done");
}

struct Process {
    pub tickets: i32,
    // size of below code [can become infinite](https://stackoverflow.com/a/25296420/8562273)
    // pub next: Option<Process>,
    pub next: Option<Rc<RefCell<Process>>>,
}

impl Process {
    pub fn new(ticket: i32) -> Self {
        Process {
            tickets: ticket,
            next: None,
        }
    }
}

fn process_pop() -> i32 {
    let mut _head = Process::new(0);
    let value = _head.tickets;
    match Rc::try_unwrap(_head.next.unwrap()) {
        Ok(refcell) => {
            _head = refcell.into_inner();
        }
        Err(_err) => {
            println!("error when RC::try_unwrap");
        }
    }

    return value;
}
