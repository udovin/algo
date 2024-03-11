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
	Len() int
	Iter() MapIter[K, V]
}

func NewMap[K, V any](less func(K, K) bool) Map[K, V] {
	return &mapImpl[K, V]{less: less}
}

const (
	mapDegree = 32
	maxLen    = mapDegree*2 - 1
	minLen    = maxLen / 2
)

type mapNode[K, V any] struct {
	keys     [maxLen]K
	values   [maxLen]V
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
			return n.values[i], true
		}
		if n.children == nil {
			return empty, false
		}
		n = n.children[i]
	}
}

func (m *mapImpl[K, V]) Set(key K, value V) {
	m.setRootNode(key, value)
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

// search returns `pos` that `keys[pos] >= key` and flag that `keys[pos] == key`.
func (m *mapImpl[K, V]) search(n *mapNode[K, V], key K) (int, bool) {
	low, high := 0, n.len
	for low < high {
		mid := (low + high) / 2
		if m.less(key, n.keys[mid]) {
			high = mid
		} else {
			low = mid + 1
		}
	}
	if low > 0 && !m.less(n.keys[low-1], key) {
		return low - 1, true
	}
	return low, false
}

func (m *mapImpl[K, V]) setRootNode(key K, value V) {
	if m.root == nil {
		m.root = &mapNode[K, V]{len: 1}
		m.root.keys[0] = key
		m.root.values[0] = value
		m.len = 1
		return
	}
	split := m.setNode(&m.root, key, value)
	if split {
		left := m.root
		k, v, right := m.splitNode(left)
		m.root = &mapNode[K, V]{len: 1}
		m.root.keys[0] = k
		m.root.values[0] = v
		m.root.children = &[maxLen + 1]*mapNode[K, V]{left, right}
		m.setRootNode(key, value)
	}
}

func (m *mapImpl[K, V]) setNode(p **mapNode[K, V], key K, value V) bool {
	n := *p
	i, ok := m.search(*p, key)
	if ok {
		n.keys[i] = key
		n.values[i] = value
		return false
	}
	if n.children == nil {
		if n.len == maxLen {
			return true
		}
		copy(n.keys[i+1:], n.keys[i:n.len])
		copy(n.values[i+1:], n.values[i:n.len])
		n.keys[i] = key
		n.values[i] = value
		n.len++
		m.len++
		return false
	}
	split := m.setNode(&n.children[i], key, value)
	if split {
		if n.len == maxLen {
			return true
		}
		k, v, right := m.splitNode(n.children[i])
		copy(n.keys[i+1:], n.keys[i:n.len])
		copy(n.values[i+1:], n.values[i:n.len])
		n.keys[i] = k
		n.values[i] = v
		copy(n.children[i+1:], n.children[i:n.len+1])
		n.children[i+1] = right
		n.len++
		return m.setNode(p, key, value)
	}
	return false
}

func (m *mapImpl[K, V]) splitNode(n *mapNode[K, V]) (K, V, *mapNode[K, V]) {
	i := maxLen / 2
	key := n.keys[i]
	value := n.values[i]
	right := &mapNode[K, V]{len: n.len - i - 1}
	copy(right.keys[:], n.keys[i+1:])
	copy(right.values[:], n.values[i+1:])
	if n.children != nil {
		right.children = &[maxLen + 1]*mapNode[K, V]{}
		copy(right.children[:], n.children[i+1:])
	}
	var emptyKey K
	var emptyValue V
	for j := i; j < maxLen; j++ {
		n.keys[j] = emptyKey
		n.values[j] = emptyValue
		if n.children != nil {
			n.children[j+1] = nil
		}
	}
	n.len = i
	return key, value, right
}

func (m *mapImpl[K, V]) deleteNode(p **mapNode[K, V], key K) bool {
	n := *p
	i, ok := m.search(n, key)
	if n.children == nil {
		if ok {
			copy(n.keys[i:], n.keys[i+1:n.len])
			copy(n.values[i:], n.values[i+1:n.len])
			n.len--
			var emptyKey K
			var emptyValue V
			n.keys[n.len] = emptyKey
			n.values[n.len] = emptyValue
			m.len--
			return true
		}
		return false
	}
	deleted := false
	if ok {
		n.keys[i], n.values[i] = m.deleteMaxItem(n.children[i])
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

func (m *mapImpl[K, V]) deleteMaxItem(n *mapNode[K, V]) (K, V) {
	for {
		if n.children == nil {
			n.len--
			key := n.keys[n.len]
			value := n.values[n.len]
			var emptyKey K
			var emptyValue V
			n.keys[n.len] = emptyKey
			n.values[n.len] = emptyValue
			return key, value
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
		copy(node.keys[:], left.keys[:left.len])
		copy(node.values[:], left.values[:left.len])
		node.keys[left.len] = n.keys[i]
		node.values[left.len] = n.values[i]
		copy(node.keys[left.len+1:], right.keys[:right.len])
		copy(node.values[left.len+1:], right.values[:right.len])
		if left.children != nil {
			node.children = &[maxLen + 1]*mapNode[K, V]{}
			copy(node.children[:], left.children[:left.len+1])
			copy(node.children[left.len+1:], right.children[:right.len+1])
		}
		copy(n.keys[i:], n.keys[i+1:n.len])
		copy(n.values[i:], n.values[i+1:n.len])
		copy(n.children[i+1:], n.children[i+2:n.len+1])
		n.children[i] = node
		n.children[n.len] = nil
		n.len--
		var emptyKey K
		var emptyValue V
		n.keys[n.len] = emptyKey
		n.values[n.len] = emptyValue
	} else if left.len > right.len {
		copy(right.keys[1:], right.keys[:right.len])
		copy(right.values[1:], right.values[:right.len])
		right.keys[0] = n.keys[i]
		right.values[0] = n.values[i]
		right.len++
		n.keys[i] = left.keys[left.len-1]
		n.values[i] = left.values[left.len-1]
		if left.children != nil {
			copy(right.children[1:], right.children[:right.len])
			right.children[0] = left.children[left.len]
			left.children[left.len] = nil
		}
		left.len--
		var emptyKey K
		var emptyValue V
		left.keys[left.len] = emptyKey
		left.values[left.len] = emptyValue
	} else {
		left.keys[left.len] = n.keys[i]
		left.values[left.len] = n.values[i]
		left.len++
		n.keys[i] = right.keys[0]
		n.values[i] = right.values[0]
		copy(right.keys[:], right.keys[1:right.len])
		copy(right.values[:], right.values[1:right.len])
		if left.children != nil {
			left.children[left.len] = right.children[0]
			copy(right.children[:], right.children[1:right.len+1])
			right.children[right.len] = nil
		}
		right.len--
		var emptyKey K
		var emptyValue V
		right.keys[right.len] = emptyKey
		right.values[right.len] = emptyValue
	}
}

type mapIter[K, V any] struct {
	m      *mapImpl[K, V]
	stack  []mapIterPos[K, V]
	key    K
	value  *V
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
	m.key = n.keys[0]
	m.value = &n.values[0]
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
	m.key = n.keys[n.len-1]
	m.value = &n.values[n.len-1]
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
					var emptyKey K
					m.key = emptyKey
					m.value = nil
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
	m.key = s.n.keys[s.i]
	m.value = &s.n.values[s.i]
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
					var emptyKey K
					m.key = emptyKey
					m.value = nil
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
	m.key = s.n.keys[s.i]
	m.value = &s.n.values[s.i]
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
			m.key = n.keys[i]
			m.value = &n.values[i]
			return true
		}
		if n.children == nil {
			m.stack[len(m.stack)-1].i--
			return m.Next()
		}
		n = n.children[i]
	}
}

func (m *mapIter[K, V]) SeekPrev(key K) bool {
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
			m.key = n.keys[i]
			m.value = &n.values[i]
			return true
		}
		if n.children == nil {
			return m.Prev()
		}
		n = n.children[i]
	}
}

func (m *mapIter[K, V]) Key() K {
	return m.key
}

func (m *mapIter[K, V]) Value() V {
	return *m.value
}
