package btree

type MapIter[K, V any] interface {
	Next() bool
	Prev() bool
	First() bool
	Last() bool
	Seek(K) bool
	Key() K
	Value() V
}

// Map represents map implementation using B-Tree.
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
	maxLen    = mapDegree*2 - 1
	minLen    = maxLen / 2
)

type mapItem[K, V any] struct {
	value V
	key   K
}

type mapNode[K, V any] struct {
	items    [maxLen]mapItem[K, V]
	children *[maxLen + 1]*mapNode[K, V]
	len      int
}

type mapImpl[K, V any] struct {
	root *mapNode[K, V]
	less func(K, K) bool
	len  int
}

func (m *mapImpl[K, V]) Get(key K) (V, bool) {
	n := m.root
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
	item := mapItem[K, V]{key: key, value: value}
	m.setRootNode(item)
}

func (m *mapImpl[K, V]) Delete(key K) {
	if m.root == nil {
		return
	}
	m.deleteNode(&m.root, key)
	if m.root.len == 0 && m.root.children != nil {
		m.root = m.root.children[0]
	}
	if m.len == 0 {
		m.root = nil
	}
}

func (m *mapImpl[K, V]) Len() int {
	return m.len
}

func (m *mapImpl[K, V]) Iter() MapIter[K, V] {
	return &mapIter[K, V]{m: m}
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

func (m *mapImpl[K, V]) setRootNode(item mapItem[K, V]) {
	if m.root == nil {
		m.root = &mapNode[K, V]{len: 1}
		m.root.items[0] = item
		m.len = 1
		return
	}
	split := m.setNode(&m.root, item)
	if split {
		left := m.root
		mid, right := m.splitNode(left)
		m.root = &mapNode[K, V]{len: 1}
		m.root.items[0] = mid
		m.root.children = &[maxLen + 1]*mapNode[K, V]{left, right}
		m.setRootNode(item)
	}
}

func (m *mapImpl[K, V]) setNode(p **mapNode[K, V], item mapItem[K, V]) bool {
	n := *p
	i, ok := m.search(*p, item.key)
	if ok {
		n.items[i] = item
		return false
	}
	if n.children == nil {
		if n.len == maxLen {
			return true
		}
		copy(n.items[i+1:], n.items[i:n.len])
		n.items[i] = item
		n.len++
		m.len++
		return false
	}
	split := m.setNode(&n.children[i], item)
	if split {
		if n.len == maxLen {
			return true
		}
		mid, right := m.splitNode(n.children[i])
		copy(n.items[i+1:], n.items[i:n.len])
		n.items[i] = mid
		copy(n.children[i+1:], n.children[i:n.len+1])
		n.children[i+1] = right
		n.len++
		return m.setNode(p, item)
	}
	return false
}

func (m *mapImpl[K, V]) splitNode(n *mapNode[K, V]) (mapItem[K, V], *mapNode[K, V]) {
	i := maxLen / 2
	mid := n.items[i]
	right := &mapNode[K, V]{len: n.len - i - 1}
	copy(right.items[:], n.items[i+1:])
	if n.children != nil {
		right.children = &[maxLen + 1]*mapNode[K, V]{}
		copy(right.children[:], n.children[i+1:])
	}
	for j := i; j < maxLen; j++ {
		n.items[j] = mapItem[K, V]{}
		if n.children != nil {
			n.children[j+1] = nil
		}
	}
	n.len = i
	return mid, right
}

func (m *mapImpl[K, V]) deleteNode(p **mapNode[K, V], key K) bool {
	n := *p
	i, ok := m.search(n, key)
	if n.children == nil {
		if ok {
			copy(n.items[i:], n.items[i+1:n.len])
			n.len--
			n.items[n.len] = mapItem[K, V]{}
			m.len--
			return true
		}
		return false
	}
	deleted := false
	if ok {
		item := m.deleteMaxItem(n.children[i])
		n.items[i] = item
		m.len--
		deleted = true
	} else {
		deleted = m.deleteNode(&n.children[i], key)
	}
	if !deleted {
		return false
	}
	if n.children[i].len < minLen {
		m.rebalanceNode(n, i)
	}
	return true
}

func (m *mapImpl[K, V]) deleteMaxItem(n *mapNode[K, V]) mapItem[K, V] {
	for {
		if n.children == nil {
			n.len--
			item := n.items[n.len]
			n.items[n.len] = mapItem[K, V]{}
			return item
		}
		n = n.children[n.len]
	}
}

func (m *mapImpl[K, V]) rebalanceNode(n *mapNode[K, V], i int) {
	if i == n.len {
		i--
	}
	left := n.children[i]
	right := n.children[i+1]
	if left.len+right.len < maxLen {
		node := &mapNode[K, V]{len: left.len + right.len + 1}
		copy(node.items[:], left.items[:left.len])
		node.items[left.len] = n.items[i]
		copy(node.items[left.len+1:], right.items[:right.len])
		if left.children != nil {
			node.children = &[maxLen + 1]*mapNode[K, V]{}
			copy(node.children[:], left.children[:left.len+1])
			copy(node.children[left.len+1:], right.children[:right.len+1])
		}
		copy(n.items[i:], n.items[i+1:n.len])
		copy(n.children[i+1:], n.children[i+2:n.len+1])
		n.children[i] = node
		n.children[n.len] = nil
		n.len--
		n.items[n.len] = mapItem[K, V]{}
	} else if left.len > right.len {
		copy(right.items[1:], right.items[:right.len])
		right.items[0] = n.items[i]
		right.len++
		n.items[i] = left.items[left.len-1]
		if left.children != nil {
			copy(right.children[1:], right.children[:right.len])
			right.children[0] = left.children[left.len]
			left.children[left.len] = nil
		}
		left.len--
		left.items[left.len] = mapItem[K, V]{}
	} else {
		left.items[left.len] = n.items[i]
		left.len++
		n.items[i] = right.items[0]
		copy(right.items[:], right.items[1:right.len])
		if left.children != nil {
			left.children[left.len] = right.children[0]
			copy(right.children[:], right.children[1:right.len+1])
			right.children[right.len] = nil
		}
		right.len--
		right.items[right.len] = mapItem[K, V]{}
	}
}

type mapIter[K, V any] struct {
	m      *mapImpl[K, V]
	stack  []mapIterPos[K, V]
	item   mapItem[K, V]
	seeked bool
}

type mapIterPos[K, V any] struct {
	n *mapNode[K, V]
	i int
}

func (m *mapIter[K, V]) First() bool {
	if m.m.root == nil {
		return false
	}
	m.seeked = true
	m.stack = m.stack[:0]
	n := m.m.root
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
	if m.m.root == nil {
		return false
	}
	m.seeked = true
	m.stack = m.stack[:0]
	n := m.m.root
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

func (m *mapIter[K, V]) Seek(key K) bool {
	if m.m.root == nil {
		return false
	}
	m.seeked = true
	m.stack = m.stack[:0]
	n := m.m.root
	for {
		i, ok := m.m.search(n, key)
		m.stack = append(m.stack, mapIterPos[K, V]{n, i})
		if ok {
			m.item = n.items[i]
			return true
		}
		if n.children == nil {
			m.stack[len(m.stack)-1].i--
			return m.Next()
		}
		n = n.children[i]
	}
}

func (m *mapIter[K, V]) Key() K {
	return m.item.key
}

func (m *mapIter[K, V]) Value() V {
	return m.item.value
}
