package btree

import (
	"sync"
	"sync/atomic"
)

type MapIter[K, V any] interface {
	Next() bool
	Prev() bool
	First() bool
	Last() bool
	Key() K
	Value() V
}

type Map[K, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Iter() MapIter[K, V]
	Len() int
}

func NewMap[K, V any](less func(K, K) bool) Map[K, V] {
	return &mapImpl[K, V]{less: less}
}

const (
	mapDegree = 32
	mapMax    = mapDegree*2 - 1
	mapMin    = mapMax / 2
)

type mapItem[K, V any] struct {
	value V
	key   K
}

type mapNode[K, V any] struct {
	items    [mapMax]mapItem[K, V]
	children *[mapMax + 1]*mapNode[K, V]
	len      int
}

func (n *mapNode[K, V]) clone() *mapNode[K, V] {
	clone := mapNode[K, V]{
		items: n.items,
		len:   n.len,
	}
	if n.children != nil {
		childrenClone := *n.children
		clone.children = &childrenClone
	}
	return &clone
}

type mapImpl[K, V any] struct {
	mutex sync.Mutex
	root  atomic.Pointer[mapNode[K, V]]
	less  func(K, K) bool
	len   int64
}

func (m *mapImpl[K, V]) search(n *mapNode[K, V], key K) (int, bool) {
	low, high := 0, n.len
	for low < high {
		mid := (low + high) / 2
		if m.less(key, n.items[mid].key) {
			high = mid
		} else {
			low = mid + 1
		}
	}
	if low > 0 && !m.less(n.items[low-1].key, key) {
		return low - 1, true
	}
	return low, false
}

func (m *mapImpl[K, V]) Get(key K) (V, bool) {
	n := m.root.Load()
	var empty V
	if n == nil {
		return empty, false
	}
	for {
		i, ok := m.search(n, key)
		if ok {
			return n.items[i].value, true
		}
		if n.children == nil {
			return empty, false
		}
		n = n.children[i]
	}
}

func (m *mapImpl[K, V]) Set(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	item := mapItem[K, V]{key: key, value: value}
	m.setRoot(m.root.Load(), item)
}

func (m *mapImpl[K, V]) setRoot(root *mapNode[K, V], item mapItem[K, V]) {
	if root == nil {
		root = &mapNode[K, V]{len: 1}
		root.items[0] = item
		atomic.StoreInt64(&m.len, 1)
		m.root.Store(root)
		return
	}
	split := m.setNode(&root, item)
	if split {
		left := root.clone()
		mid, right := m.splitNode(left)
		root = &mapNode[K, V]{len: 1}
		root.items[0] = mid
		root.children = &[mapMax + 1]*mapNode[K, V]{left, right}
		m.setRoot(root, item)
		return
	}
	m.root.Store(root)
}

func (m *mapImpl[K, V]) setNode(p **mapNode[K, V], item mapItem[K, V]) bool {
	n := *p
	i, ok := m.search(*p, item.key)
	if ok {
		n = n.clone()
		n.items[i] = item
		*p = n
		return false
	}
	if n.children == nil {
		if n.len == mapMax {
			return true
		}
		n = n.clone()
		copy(n.items[i+1:], n.items[i:n.len])
		n.items[i] = item
		n.len++
		*p = n
		atomic.AddInt64(&m.len, 1)
		return false
	}
	child := n.children[i]
	split := m.setNode(&child, item)
	if split {
		if n.len == mapMax {
			return true
		}
		child = child.clone()
		mid, right := m.splitNode(child)
		n = n.clone()
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = mid
		copy(n.children[i+1:], n.children[i:])
		n.children[i] = child
		n.children[i+1] = right
		n.len++
		*p = n
		return m.setNode(p, item)
	}
	n = n.clone()
	n.children[i] = child
	*p = n
	return false
}

func (m *mapImpl[K, V]) splitNode(n *mapNode[K, V]) (mapItem[K, V], *mapNode[K, V]) {
	i := mapMax / 2
	mid := n.items[i]
	right := &mapNode[K, V]{len: n.len - i - 1}
	copy(right.items[:], n.items[i+1:])
	if n.children != nil {
		right.children = &[mapMax + 1]*mapNode[K, V]{}
		copy(right.children[:], n.children[i+1:])
	}
	for j := i; j < mapMax; j++ {
		n.items[j] = mapItem[K, V]{}
		if n.children != nil {
			n.children[j+1] = nil
		}
	}
	n.len = i
	return mid, right
}

func (m *mapImpl[K, V]) Delete(key K) {
}

func (m *mapImpl[K, V]) Len() int {
	return int(atomic.LoadInt64(&m.len))
}

func (m *mapImpl[K, V]) Iter() MapIter[K, V] {
	return &mapIter[K, V]{root: m.root.Load()}
}

type mapIter[K, V any] struct {
	root   *mapNode[K, V]
	stack  []mapIterPos[K, V]
	item   mapItem[K, V]
	seeked bool
}

type mapIterPos[K, V any] struct {
	n *mapNode[K, V]
	i int
}

func (m *mapIter[K, V]) First() bool {
	if m.root == nil {
		return false
	}
	m.seeked = true
	m.stack = m.stack[:0]
	n := m.root
	for {
		m.stack = append(m.stack, mapIterPos[K, V]{n, 0})
		if n.children == nil {
			break
		}
		n = n.children[0]
	}
	m.item = n.items[0]
	return true
}

func (m *mapIter[K, V]) Last() bool {
	if m.root == nil {
		return false
	}
	m.seeked = true
	m.stack = m.stack[:0]
	n := m.root
	for {
		m.stack = append(m.stack, mapIterPos[K, V]{n, n.len})
		if n.children == nil {
			m.stack[len(m.stack)-1].i--
			break
		}
		n = n.children[n.len]
	}
	m.item = n.items[n.len-1]
	return true
}

func (m *mapIter[K, V]) Next() bool {
	if !m.seeked {
		return m.First()
	}
	s := &m.stack[len(m.stack)-1]
	s.i++
	if s.n.children == nil {
		if s.i == s.n.len {
			for {
				m.stack = m.stack[:len(m.stack)-1]
				if len(m.stack) == 0 {
					m.seeked = false
					return false
				}
				s = &m.stack[len(m.stack)-1]
				if s.i < s.n.len {
					break
				}
			}
		}
	} else {
		n := s.n.children[s.i]
		for {
			m.stack = append(m.stack, mapIterPos[K, V]{n, 0})
			if n.children == nil {
				break
			}
			n = n.children[0]
		}
		s = &m.stack[len(m.stack)-1]
	}
	m.item = s.n.items[s.i]
	return true
}

func (m *mapIter[K, V]) Prev() bool {
	if !m.seeked {
		return m.Last()
	}
	s := &m.stack[len(m.stack)-1]
	if s.n.children == nil {
		s.i--
		if s.i < 0 {
			for {
				m.stack = m.stack[:len(m.stack)-1]
				if len(m.stack) == 0 {
					m.seeked = false
					return false
				}
				s = &m.stack[len(m.stack)-1]
				s.i--
				if s.i >= 0 {
					break
				}
			}
		}
	} else {
		n := s.n.children[s.i]
		for {
			m.stack = append(m.stack, mapIterPos[K, V]{n, n.len})
			if n.children == nil {
				m.stack[len(m.stack)-1].i--
				break
			}
			n = n.children[n.len]
		}
		s = &m.stack[len(m.stack)-1]
	}
	m.item = s.n.items[s.i]
	return true
}

func (m *mapIter[K, V]) Key() K {
	return m.item.key
}

func (m *mapIter[K, V]) Value() V {
	return m.item.value
}
