package maps

type Node[K, V any] struct {
	m *Map[K, V]
	n *splayNode[K, V]
}

func (n *Node[K, V]) Key() K {
	return n.n.key
}

func (n *Node[K, V]) Value() V {
	return n.n.value
}

func (n *Node[K, V]) SetValue(value V) {
	n.n.value = value
}

func (n *Node[K, V]) Next() *Node[K, V] {
	if it := n.n.next(); it != nil {
		return &Node[K, V]{m: n.m, n: it}
	}
	return nil
}

func (n *Node[K, V]) Prev() *Node[K, V] {
	if it := n.n.prev(); it != nil {
		return &Node[K, V]{m: n.m, n: it}
	}
	return nil
}

type Map[K, V any] struct {
	m *splayTree[K, V]
}

func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	if it := m.Find(key); it != nil {
		value = it.Value()
		ok = true
	}
	return
}

func (m *Map[K, V]) Set(key K, value V) {
	if it := m.Find(key); it != nil {
		it.SetValue(value)
		return
	}
	m.Insert(key, value)
}

func (m *Map[K, V]) Unset(key K) {
	if it := m.Find(key); it != nil {
		m.Erase(it)
	}
}

func (m *Map[K, V]) Insert(key K, value V) *Node[K, V] {
	return &Node[K, V]{m: m, n: m.m.insert(key, value)}
}

func (m *Map[K, V]) wrapNode(n *splayNode[K, V]) *Node[K, V] {
	if n == nil {
		return nil
	}
	return &Node[K, V]{m: m, n: n}
}

func (m *Map[K, V]) Erase(it *Node[K, V]) {
	if it == nil {
		return
	}
	if it.m != m {
		panic("unable to remove invalid node")
	}
	it.m = nil
	m.m.erase(it.n)
}

func (m *Map[K, V]) Find(key K) *Node[K, V] {
	return m.wrapNode(m.m.find(key))
}

func (m *Map[K, V]) Front() *Node[K, V] {
	return m.wrapNode(m.m.front())
}

func (m *Map[K, V]) Back() *Node[K, V] {
	return m.wrapNode(m.m.back())
}

func (m *Map[K, V]) LowerBound(key K) *Node[K, V] {
	if it := m.m.lowerBound(key); it != nil {
		return &Node[K, V]{m: m, n: it}
	}
	return nil
}

func (m *Map[K, V]) Len() int {
	return m.m.len
}

func (m *Map[K, V]) Clone() *Map[K, V] {
	return &Map[K, V]{m: m.m.clone()}
}

func NewMap[K, V any](less func(K, K) bool) *Map[K, V] {
	return &Map[K, V]{m: &splayTree[K, V]{less: less}}
}
