package linkedlist

import "iter"

// LinkedList is a doubly linked list.
type LinkedList[V any] interface {
	// All iterates over LinkedList.
	All() iter.Seq[V]
	// First element of LinkedList
	First() *Node[V]
	// Last element of LinkedList
	Last() *Node[V]
	// PushBack makes node the last element in the list.
	PushBack(node *Node[V])
	// PushFront makes node the first element in the list.
	PushFront(node *Node[V])
}

// linkedListImpl is a doubly linked list implementation.
type linkedListImpl[V any] struct {
	// head is the first element of LinkedList.
	head *Node[V]
}

// Node is an element of the doubly linked list.
type Node[V any] struct {
	// next points to the next element of doubly linked list.
	Next *Node[V]
	// Prev points to the Previous element of doubly linked list.
	Prev *Node[V]
	// value of doubly linked list element
	Value V
}

func (list *linkedListImpl[V]) All() iter.Seq[V] {
	return func(yield func(V) bool) {
		current := list.head.Next
		for current != list.head {
			if !yield(current.Value) {
				return
			}
			current = current.Next
		}
	}
}

func (list *linkedListImpl[V]) First() *Node[V] {
	return list.head.Next
}

func (list *linkedListImpl[V]) Last() *Node[V] {
	return list.head.Prev
}

// NewNode creates new list node with the given value.
func NewNode[V any](value V) *Node[V] {
	return &Node[V]{
		Value: value,
	}
}

// New creates LinkedList with dummies and a given node.
func New[V any](node *Node[V]) *linkedListImpl[V] {
	// Create dummy node to make operations with the list more
	// convenient.
	dummyHead := &Node[V]{
		Next: node,
		Prev: node,
	}
	node.Next = dummyHead
	node.Prev = dummyHead
	return &linkedListImpl[V]{
		head: dummyHead,
	}
}

func (list *linkedListImpl[V]) PushFront(node *Node[V]) {
	PutNodeBeforeAnotherNode(node, list.head.Next)
}

func (list *linkedListImpl[V]) PushBack(node *Node[V]) {
	PutNodeBeforeAnotherNode(node, list.head)
}

// PutNodeBeforeAnotherNode places given node before another node in doubly
// linked list.
func PutNodeBeforeAnotherNode[V any](node *Node[V], anotherNode *Node[V]) {
	node.Prev = anotherNode.Prev
	node.Next = anotherNode
	anotherNode.Prev.Next = node
	anotherNode.Prev = node
}

// RemoveNode removes the given node from its current position in doubly linked
// list.
func RemoveNode[V any](node *Node[V]) {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
}
