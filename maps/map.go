package maps

// Node represents element of map.
type Node[K, V any] struct {
	m *Map[K, V]
	n *avlNode[K, V]
}

// Key returns node key.
func (n *Node[K, V]) Key() K {
	return n.n.key
}

// Valure returns node value.
func (n *Node[K, V]) Value() V {
	return n.n.value
}

// SetValue sets new value to node.
func (n *Node[K, V]) SetValue(value V) {
	n.n.value = value
}

// Next returns next node.
//
// If current node is the last, Next will return nil.
func (n *Node[K, V]) Next() *Node[K, V] {
	return n.m.wrapNode(n.n.next())
}

// Prev returns prevuous node.
//
// If current node is the first, Prev will return nil.
func (n *Node[K, V]) Prev() *Node[K, V] {
	return n.m.wrapNode(n.n.prev())
}

// Map represents ordered map data structure.
//
// Map is not copyable, but Clone method creates a valid copy of map.
// Map is thread-safe in read-only mode. For parallel read-write use
// sync.RWMutex to avoid any race conditions.
type Map[K, V any] struct {
	m *avlTree[K, V]
}

// Get returns value by specified key.
//
// If there is no such node, ok will be false.
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	if it := m.Find(key); it != nil {
		value = it.Value()
		ok = true
	}
	return
}

// Set updates value by specified key.
func (m *Map[K, V]) Set(key K, value V) {
	if it := m.Find(key); it != nil {
		it.SetValue(value)
		return
	}
	m.Insert(key, value)
}

// Unset removes specified key.
func (m *Map[K, V]) Unset(key K) {
	if it := m.Find(key); it != nil {
		m.Erase(it)
	}
}

func (m *Map[K, V]) Insert(key K, value V) *Node[K, V] {
	return m.wrapNode(m.m.insert(key, value))
}

func (m *Map[K, V]) Erase(it *Node[K, V]) {
	if it.m != m {
		panic("unable to remove invalid node")
	}
	it.m = nil
	m.m.erase(it.n)
}

// Find finds node with specified key.
//
// It is safe to access nodes concurently.
func (m *Map[K, V]) Find(key K) *Node[K, V] {
	return m.wrapNode(m.m.find(key))
}

// Front returns first element of map.
//
// If there is no nodes, Front will return nil.
func (m *Map[K, V]) Front() *Node[K, V] {
	return m.wrapNode(m.m.front())
}

// Back returns last element of map.
//
// If there is no nodes, Back will return nil.
func (m *Map[K, V]) Back() *Node[K, V] {
	return m.wrapNode(m.m.back())
}

// LowerBound returns the smallest node with node.key >= key.
//
// If there is no such nodes, LowerBound will return nil.
func (m *Map[K, V]) LowerBound(key K) *Node[K, V] {
	return m.wrapNode(m.m.lowerBound(key))
}

// Len returns amount of elements in map.
func (m *Map[K, V]) Len() int {
	return m.m.len
}

// Clone creates copy of map.
func (m *Map[K, V]) Clone() *Map[K, V] {
	return &Map[K, V]{m: m.m.clone()}
}

func (m *Map[K, V]) wrapNode(n *avlNode[K, V]) *Node[K, V] {
	if n == nil {
		return nil
	}
	return &Node[K, V]{m: m, n: n}
}

// NewMap creates new instance of ordered map.
func NewMap[K, V any](less func(K, K) bool) *Map[K, V] {
	return &Map[K, V]{m: &avlTree[K, V]{less: less}}
}
