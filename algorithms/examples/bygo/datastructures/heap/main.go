package main

import (
	"container/heap"
	"fmt"
)

// An IntHeap is a min-heap of ints.
type IntHeap []int

func (h *IntHeap) Len() int {
	return len(*h)
}

func (h *IntHeap) Less(left, right int) bool {
	// https://github.com/golang/go/blob/772f024c615ec13c6cd28bf024e9d6be852201b6/src/runtime/slice.go#L15-L19
	// slice 인스턴스가 복사되지만, 슬라이스의 기반이 되는 배열은 포인터 통해 그대로 유지됩니다.
	intHeap := *h

	return intHeap[left] < intHeap[right]
}

func (h *IntHeap) Swap(i, j int) {
	intHeap := *h

	intHeap[i], intHeap[j] = intHeap[j], intHeap[i]
}

func (h *IntHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	intHeap := *h

	intHeap = append(intHeap, x.(int))
}

func (h *IntHeap) Pop() any {
	intHeap := *h

	n := len(intHeap)
	x := intHeap[n-1]
	*h = intHeap[0 : n-1] // 슬라이스 길이를 줄임

	return x
}

func minIntHeapExample() {
	fmt.Println("intMinHeapExample start")
	h := &IntHeap{2, 1, 5}
	heap.Init(h)
	heap.Push(h, 3)
	fmt.Printf("\tminimum: %d\n", (*h)[0])
	for h.Len() > 0 {
		fmt.Printf("\tpop: %d\n", heap.Pop(h))
	}
}

// An Item is something we manage in a priority queue.
type Item struct {
	value    string // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq *PriorityQueue) Len() int {
	return len(*pq)
}

func (pq *PriorityQueue) Less(left, right int) bool {
	pqHeap := *pq
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pqHeap[left].priority > pqHeap[right].priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	pqHeap := *pq

	pqHeap[i], pqHeap[j] = pqHeap[j], pqHeap[i]
	pqHeap[i].index = i
	pqHeap[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq

	n := len(old)
	item := old[n-1]

	// Golang GC는 어떠한 객체(값, 구조체 등)에 대한 포인터 참조가 계속 남아 있으면,
	// 그 객체를 “아직 쓸 것”으로 인식해 수거(reclaim)하지 않습니다.
	//
	// 만약 `old[n-1]` 위치에 계속 `item`이 남아있다면,
	// 그 슬라이스 내부 배열(array) 에 item에 대한 참조가 살아 있는 상태가 됩니다.
	// 그러면 Go GC는
	// - 슬라이스가 그 원소를 아직 참조하고 있다고 판단합니다.
	// - 해당 객체가 아직 도달 가능(reachable)하다고 해석하고 해제하지 못합니다.
	//
	// `Pop` 되는 아이템의 인덱스에 `nil`을 설정하면 더 이상 그 아이템을 참조하지 않는다는 의미가 되고,
	// 해당 객체는 가비지로 판단하여 메모리를 해제(reclaim)해버릴 수 있게 됩니다.
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // 다른 곳에서 사용할 경우 보호 목적으로 "안 쓰는 상태"임을 표시
	*pq = old[0 : n-1]

	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

func maxPriorityQueueExample() {
	fmt.Println("minPriorityQueueExample start")
	// Some items and their priorities.
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, 0, len(items))
	i := 0
	for value, priority := range items {
		pq = append(pq, &Item{
			value:    value,
			priority: priority,
			index:    i,
		})
		i++
	}
	heap.Init(&pq)

	// Insert a new item and then modify its priority.
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&pq, item)
	pq.update(item, item.value, 5)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("\t[%d] %.2d:%s\n", item.index, item.priority, item.value)
	}
}

func main() {
	minIntHeapExample()
	maxPriorityQueueExample()
}
