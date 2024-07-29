use std::collections::LinkedList;

pub fn list_pop() {
    let mut list: LinkedList<u32> = LinkedList::new();
    list.pop_back();
}
