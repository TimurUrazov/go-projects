package lfu

import (
	"errors"
	"iter"
	"lfucache/internal/linkedlist"
)

// ErrKeyNotFound is an error that indicates that a requested key does not
// exist in the cache. It is used for operations that attempt to retrieve a
// value in the cache when the specified key is not found.
var ErrKeyNotFound = errors.New("key not found")

const DefaultCapacity = 5

// CacheItem is the item stored in the cache.
type CacheItem[K comparable, V any] struct {
	// value of cache item
	value V
	// key of cache item
	key K
	// frequency of usage of cache item
	frequency int
}

// Frequency is cache item usage frequency.
type Frequency struct {
	counter int
}

// FrequencyGroup represents group of cache items that share the same frequency.
type FrequencyGroup[V any] struct {
	// frequency of all elements in list
	frequency int
	// elementsList contains elements with the same frequency.
	elementsList linkedlist.LinkedList[V]
	// size of list
	size int
}

// Cache
// O(capacity) memory
type Cache[K comparable, V any] interface {
	// Get returns the value of the key if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1)
	Get(key K) (V, error)

	// Put updates the value of the key if present, or inserts the key if not already present.
	//
	// When the cache reaches its capacity, it should invalidate and remove the least frequently used key
	// before inserting a new item. For this problem, when there is a tie
	// (i.e., two or more keys with the same frequency), the least recently used key would be invalidated.
	//
	// O(1)
	Put(key K, value V)

	// All returns the iterator in descending order of frequency.
	// If two or more keys have the same frequency, the most recently used key will be listed first.
	//
	// O(capacity)
	All() iter.Seq2[K, V]

	// Size returns the cache size.
	//
	// O(1)
	Size() int

	// Capacity returns the cache capacity.
	//
	// O(1)
	Capacity() int

	// GetKeyFrequency returns the element's frequency if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1)
	GetKeyFrequency(key K) (int, error)
}

// cacheImpl represents LFU cache implementation
type cacheImpl[K comparable, V any] struct {
	// freqToFreqGroupNode maps each frequency to corresponding frequency
	// group.
	freqToFreqGroupNode map[int]*linkedlist.Node[FrequencyGroup[CacheItem[K, V]]]
	// freqGroupsList is a doubly linked list of frequency groups.
	freqGroupsList linkedlist.LinkedList[FrequencyGroup[CacheItem[K, V]]]
	// keyToCacheItem maps each key to its cache item with corresponding value.
	keyToCacheItem map[K]*linkedlist.Node[CacheItem[K, V]]
	// capacity serves the cache capacity.
	capacity int
	// size serves the cache size.
	size int
	// freeNodesOfFreqGroups serves unused nodes of frequency groups.
	freeNodesOfFreqGroups []*linkedlist.Node[FrequencyGroup[CacheItem[K, V]]]
}

// New initializes the cache with the given capacity.
// If no capacity is provided, the cache will use DefaultCapacity.
func New[K comparable, V any](capacity ...int) *cacheImpl[K, V] {
	var cacheCapacity int
	length := len(capacity)
	if length == 0 {
		cacheCapacity = DefaultCapacity
	} else if length > 2 {
		panic("Invalid capacity")
	} else {
		cacheCapacity = capacity[0]
		// Capacity cannot be negative.
		if cacheCapacity < 0 {
			panic("Invalid capacity")
		}
	}
	// Since the maximum size of the cache is known, memory for its elements
	// can be allocated in advance.
	return &cacheImpl[K, V]{
		capacity:              cacheCapacity,
		freqToFreqGroupNode:   make(map[int]*linkedlist.Node[FrequencyGroup[CacheItem[K, V]]], cacheCapacity),
		keyToCacheItem:        make(map[K]*linkedlist.Node[CacheItem[K, V]], cacheCapacity),
		freeNodesOfFreqGroups: make([]*linkedlist.Node[FrequencyGroup[CacheItem[K, V]]], 0, cacheCapacity),
	}
}

func (l *cacheImpl[K, V]) Get(key K) (V, error) {
	var value V

	// If the cache item exists, find it in the keyToCacheItem mapping;
	// otherwise, return an error.
	if cacheItem, ok := l.keyToCacheItem[key]; ok {
		value = cacheItem.Value.value
		// If it exists, its frequency will be updated.
		l.updateFreqAndMoveCacheItemNode(cacheItem)
		return value, nil
	}

	return value, ErrKeyNotFound
}

func (l *cacheImpl[K, V]) Put(key K, value V) {
	// Before placing the cache item, it should be checked whether such an item
	// exists.
	if cacheItem, ok := l.keyToCacheItem[key]; ok {
		// If it exists, its frequency should be updated.
		l.updateFreqAndMoveCacheItemNode(cacheItem)
		cacheItem.Value.value = value
	} else {
		// If it does not exist, it should be checked whether the capacity has
		// been exceeded.
		var cacheItemNode *linkedlist.Node[CacheItem[K, V]]
		if l.size == l.capacity {
			// Retrieve the element with the lowest usage frequency and its
			// group.
			minFrequencyGroup := l.freqGroupsList.Last()
			cacheItemNode = minFrequencyGroup.Value.elementsList.Last()
			// Update the value of the last item and remove the old item from
			// keyToCacheItem.
			delete(l.keyToCacheItem, cacheItemNode.Value.key)
			cacheItemNode.Value.key = key
			cacheItemNode.Value.value = value
			// If the minimum frequency group is not equal to 1, a new group
			// needs to be created. Otherwise, make the cache item the most
			// recently used if it is not the only one in the group.
			if minFrequencyGroup.Value.frequency != 1 {
				// If the cache item is the only one in the group, updating the
				// group's frequency to 1 will suffice. Otherwise, remove the
				// item from the old group and place it into the group with
				// frequency 1.
				if minFrequencyGroup.Value.size == 1 {
					minFrequencyGroup.Value.frequency = 1
					l.freqToFreqGroupNode[1] = minFrequencyGroup
				} else {
					minFrequencyGroup.Value.size--
					linkedlist.RemoveNode(cacheItemNode)
					l.freqToFreqGroupNode[1] = l.getNewFrequencyGroupNode(
						cacheItemNode, 1,
					)
					l.freqGroupsList.PushBack(l.freqToFreqGroupNode[1])
				}
			} else if minFrequencyGroup.Value.size != 1 {
				linkedlist.RemoveNode(cacheItemNode)
				minFrequencyGroup.Value.elementsList.PushFront(cacheItemNode)
				cacheItemNode.Value.frequency =
					minFrequencyGroup.Value.frequency
			}
		} else {
			var unitFrequencyGroupNode *linkedlist.Node[FrequencyGroup[CacheItem[K, V]]]
			// Create a cache item node to insert it into either the newly
			// created list or an existing one.
			cacheItemNode = linkedlist.NewNode(CacheItem[K, V]{
				key:   key,
				value: value,
			})
			// If the list is empty, it needs to be created.
			if l.size == 0 {
				unitFrequencyGroupNode = createFrequencyGroupNode(
					cacheItemNode, 1,
				)
				l.freqGroupsList = linkedlist.New(
					unitFrequencyGroupNode,
				)
			} else {
				// If the list has already been created, locate the group with
				// frequency 1 and place the element there. If such a group
				// does not exist, create it.
				if l.freqGroupsList.Last().Value.frequency == 1 {
					lastListElement := l.freqGroupsList.Last()
					unitFrequencyGroupNode = lastListElement
					cacheItemNode.Value.frequency =
						unitFrequencyGroupNode.Value.frequency
					unitFrequencyGroupNode.Value.elementsList.PushFront(cacheItemNode)
					unitFrequencyGroupNode.Value.size++
				} else {
					unitFrequencyGroupNode = l.getNewFrequencyGroupNode(
						cacheItemNode, 1,
					)
					l.freqGroupsList.PushBack(unitFrequencyGroupNode)
				}
			}
			l.freqToFreqGroupNode[1] = unitFrequencyGroupNode
			// Increase the size of the cache.
			l.size++
		}
		// Also, create a mapping from key to cacheItemNode.
		l.keyToCacheItem[key] = cacheItemNode
	}
}

// createFrequencyGroupNode creates node with group of given frequency which
// includes given cache item.
func createFrequencyGroupNode[K comparable, V any](
	cacheItemNode *linkedlist.Node[CacheItem[K, V]],
	frequency int,
) *linkedlist.Node[FrequencyGroup[CacheItem[K, V]]] {
	frequencyGroupNode := linkedlist.NewNode(
		FrequencyGroup[CacheItem[K, V]]{
			elementsList: linkedlist.New(cacheItemNode),
			size:         1,
			frequency:    frequency,
		},
	)
	cacheItemNode.Value.frequency = frequency
	return frequencyGroupNode
}

// updateFreqAndMoveCacheItemNode increases the cache item's usage frequency
// and moves it to the corresponding frequency group
func (l *cacheImpl[K, V]) updateFreqAndMoveCacheItemNode(
	cacheItemNode *linkedlist.Node[CacheItem[K, V]],
) {
	// Retrieve frequency group of the cacheItemNode.
	currentFrequency := cacheItemNode.Value.frequency
	currentFrequencyGroupNode := l.freqToFreqGroupNode[currentFrequency]

	// Increase the cache item's frequency by 1.
	newFrequency := currentFrequency + 1
	// Reduce the size of the frequency group before removing the element.
	currentFrequencyGroupNode.Value.size--

	// Check that the cache item being moved is not the last item in the list.
	if currentFrequencyGroupNode.Value.size == 0 {
		// Otherwise, remove the frequency group from freqToFreqGroupNode.
		delete(l.freqToFreqGroupNode, currentFrequency)
	}

	// Get the group with a frequency higher than the current frequency.
	greaterFrequencyGroup := currentFrequencyGroupNode.Prev.Value
	// Compare this frequency with newFrequency.
	if greaterFrequencyGroup.frequency == newFrequency {
		// If there is a group with a frequency equal to newFrequency, set the
		// current cache item as the most recently used item in that group.
		linkedlist.RemoveNode(cacheItemNode)
		greaterFrequencyGroup.elementsList.PushFront(cacheItemNode)
		currentFrequencyGroupNode.Prev.Value.size++
		// Change the pointer to the frequency of the new group.
		cacheItemNode.Value.frequency = greaterFrequencyGroup.frequency
		// If the element was the last one in the old group, remember to place
		// the node with the frequency group in the list of unused nodes.
		if currentFrequencyGroupNode.Value.size == 0 {
			linkedlist.RemoveNode(currentFrequencyGroupNode)
			l.freeNodesOfFreqGroups = append(l.freeNodesOfFreqGroups, currentFrequencyGroupNode)
		}
	} else {
		// If there is no group with a frequency equal to newFrequency, create
		// this group and place the given cache item into it.
		if currentFrequencyGroupNode.Value.size == 0 {
			// If the element is the only one in the current group, simply
			// update the frequency counter of the current group to the new
			// value, and create a mapping from the new frequency to this
			// group.
			currentFrequencyGroupNode.Value.frequency = newFrequency
			l.freqToFreqGroupNode[newFrequency] = currentFrequencyGroupNode
			currentFrequencyGroupNode.Value.size++
			cacheItemNode.Value.frequency = newFrequency
		} else {
			// If there are other elements remaining in the current group, the
			// current element should be removed from it and placed in the new
			// group.
			linkedlist.RemoveNode(cacheItemNode)
			l.freqToFreqGroupNode[newFrequency] = l.getNewFrequencyGroupNode(
				cacheItemNode, newFrequency,
			)
			linkedlist.PutNodeBeforeAnotherNode(
				l.freqToFreqGroupNode[newFrequency],
				currentFrequencyGroupNode,
			)
		}
	}
}

// getNewFrequencyGroupNode retrieves a new group with the specified frequency
// and inserts cacheItemNode in corresponding list.
func (l *cacheImpl[K, V]) getNewFrequencyGroupNode(
	cacheItemNode *linkedlist.Node[CacheItem[K, V]],
	newFrequency int,
) *linkedlist.Node[FrequencyGroup[CacheItem[K, V]]] {
	var newFrequencyGroupNode *linkedlist.Node[FrequencyGroup[CacheItem[K, V]]]
	// It is necessary to check if there is an unused frequency group.
	freeNodesOfFreqGroups := l.freeNodesOfFreqGroups
	freeNodesLength := len(freeNodesOfFreqGroups)
	if freeNodesLength == 0 {
		// If it doesn't exist, create it and place the cacheItemNode inside
		// it.
		newFrequencyGroupNode = createFrequencyGroupNode(
			cacheItemNode, newFrequency,
		)
	} else {
		// If an unused frequency group exists, use it and place the
		// cacheItemNode inside, making this cache item the most recently used
		// in that group.
		newFrequencyGroupNode = freeNodesOfFreqGroups[freeNodesLength-1]
		l.freeNodesOfFreqGroups = freeNodesOfFreqGroups[:freeNodesLength-1]
		newFrequencyGroupNode.Value.elementsList.PushFront(cacheItemNode)
		// Update the pointer in the cache item to the new frequency and
		// refresh the frequency of the group.
		newFrequencyGroupNode.Value.size = 1
		newFrequencyGroupNode.Value.frequency = newFrequency
		cacheItemNode.Value.frequency = newFrequency
	}
	return newFrequencyGroupNode
}

func (l *cacheImpl[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// If nothing has been placed in the cache, then the freqGroupsList
		// has not been created.
		if l.size == 0 {
			return
		}
		// If there is at least one element in the cache, then freqGroupsList
		// can be iterated over (as can elementsList).
		l.freqGroupsList.All()(func(freqGroup FrequencyGroup[CacheItem[K, V]]) bool {
			yieldResult := true
			freqGroup.elementsList.All()(func(cacheItem CacheItem[K, V]) bool {
				yieldResult = yield(cacheItem.key, cacheItem.value)
				return yieldResult
			})
			return yieldResult
		})
	}
}

func (l *cacheImpl[K, V]) Size() int {
	return l.size
}

func (l *cacheImpl[K, V]) Capacity() int {
	return l.capacity
}

func (l *cacheImpl[K, V]) GetKeyFrequency(key K) (int, error) {
	// If the element exists, it will be found in the keyToCacheItem mapping,
	// or an error will be returned otherwise.
	// At the same time, there is no need to increase the frequency since the
	// cache item itself is not being retrieving.
	if element, ok := l.keyToCacheItem[key]; !ok {
		return 0, ErrKeyNotFound
	} else {
		return element.Value.frequency, nil
	}
}
