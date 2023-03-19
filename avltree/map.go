package avltree

// Node represents element of map.
type Node[K, V any] struct {
	key    K
	value  V
	height int8
	left   *Node[K, V]
	right  *Node[K, V]
	parent *Node[K, V]
}

// Key returns node key.
func (n *Node[K, V]) Key() K {
	return n.key
}

// Valure returns node value.
func (n *Node[K, V]) Value() V {
	return n.value
}

// SetValue sets new value to node.
func (n *Node[K, V]) SetValue(value V) {
	n.value = value
}

// Next returns next node.
//
// If current node is the last, Next will return nil.
func (n *Node[K, V]) Next() *Node[K, V] {
	if n.right != nil {
		next := n.right
		for next.left != nil {
			next = next.left
		}
		return next
	}
	next := n
	for next.parent != nil && next.parent.right == next {
		next = next.parent
	}
	return next.parent
}

// Prev returns prevuous node.
//
// If current node is the first, Prev will return nil.
func (n *Node[K, V]) Prev() *Node[K, V] {
	if n.left != nil {
		next := n.left
		for next.right != nil {
			next = next.right
		}
		return next
	}
	next := n
	for next.parent != nil && next.parent.left == next {
		next = next.parent
	}
	return next.parent
}

// Map represents ordered map data structure.
//
// Map is not copyable, but Clone method creates a valid copy of map.
// Map is thread-safe in read-only mode. For parallel read-write use
// sync.RWMutex to avoid any race conditions.
type Map[K, V any] struct {
	root *Node[K, V]
	less func(K, K) bool
	len  int
}

// Get returns value by specified key.
//
// If there is no such node, ok will be false.
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	if it := m.Find(key); it != nil {
		value = it.value
		ok = true
	}
	return
}

// Set updates value by specified key.
func (m *Map[K, V]) Set(key K, value V) {
	if it := m.Find(key); it != nil {
		it.value = value
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
	n := Node[K, V]{key: key, value: value, height: 1}
	if m.root == nil {
		m.root = &n
		m.len = 1
		return &n
	}
	p := m.root
	for {
		if m.less(key, p.key) {
			if p.left == nil {
				p.left = &n
				n.parent = p
				break
			}
			p = p.left
		} else {
			if p.right == nil {
				p.right = &n
				n.parent = p
				break
			}
			p = p.right
		}
	}
	m.len++
	m.rebalance(p)
	return &n
}

func (m *Map[K, V]) Erase(n *Node[K, V]) {
	if r := getRoot(n); r != m.root {
		panic("attempt to erase node from wrong map")
	}
	if n.left == nil && n.right == nil {
		if n.parent == nil {
			m.root = nil
			m.len = 0
			return
		}
		if n.parent.left == n {
			n.parent.left = nil
		} else {
			n.parent.right = nil
		}
		m.len--
		m.rebalance(n.parent)
		n.parent = nil
		return
	}
	if n.left == nil {
		if n.parent == nil {
			n.right.parent = nil
			m.root = n.right
			m.len--
			n.right = nil
			return
		}
		if n.parent.left == n {
			n.parent.left = n.right
		} else {
			n.parent.right = n.right
		}
		n.right.parent = n.parent
		m.len--
		m.rebalance(n.parent)
		n.right = nil
		n.parent = nil
		return
	}
	if n.right == nil {
		if n.parent == nil {
			n.left.parent = nil
			m.root = n.left
			m.len--
			n.left = nil
			return
		}
		if n.parent.left == n {
			n.parent.left = n.left
		} else {
			n.parent.right = n.left
		}
		n.left.parent = n.parent
		m.len--
		m.rebalance(n.parent)
		n.left = nil
		n.parent = nil
		return
	}
	r := n.right
	if r.left == nil {
		r.left = n.left
		r.parent = n.parent
		if r.parent != nil {
			if r.parent.left == n {
				r.parent.left = r
			} else {
				r.parent.right = r
			}
		}
		if r.left != nil {
			r.left.parent = r
		}
		m.len--
		m.rebalance(r)
		n.left = nil
		n.right = nil
		n.parent = nil
		return
	}
	for r.left != nil {
		r = r.left
	}
	p := r.parent
	p.left = r.right
	if p.left != nil {
		p.left.parent = p
	}
	r.left = n.left
	r.right = n.right
	r.parent = n.parent
	if r.parent != nil {
		if r.parent.left == n {
			r.parent.left = r
		} else {
			r.parent.right = r
		}
	}
	if r.left != nil {
		r.left.parent = r
	}
	if r.right != nil {
		r.right.parent = r
	}
	m.len--
	m.rebalance(p)
	n.left = nil
	n.right = nil
	n.parent = nil
}

// Find finds node with specified key.
//
// It is safe to access nodes concurently.
func (m *Map[K, V]) Find(key K) *Node[K, V] {
	n := m.LowerBound(key)
	if n == nil || m.less(key, n.key) {
		return nil
	}
	return n
}

// Front returns first element of map.
//
// If there is no nodes, Front will return nil.
func (m *Map[K, V]) Front() *Node[K, V] {
	if m.root == nil {
		return nil
	}
	it := m.root
	for it.left != nil {
		it = it.left
	}
	return it
}

// Back returns last element of map.
//
// If there is no nodes, Back will return nil.
func (m *Map[K, V]) Back() *Node[K, V] {
	if m.root == nil {
		return nil
	}
	it := m.root
	for it.right != nil {
		it = it.right
	}
	return it
}

// LowerBound returns the smallest node with node.key >= key.
//
// If there is no such nodes, LowerBound will return nil.
func (m *Map[K, V]) LowerBound(key K) (n *Node[K, V]) {
	for it := m.root; it != nil; {
		if m.less(it.key, key) {
			it = it.right
		} else {
			n = it
			it = it.left
		}
	}
	return
}

// Len returns amount of elements in map.
func (m *Map[K, V]) Len() int {
	return m.len
}

// Clone creates copy of map.
func (m *Map[K, V]) Clone() *Map[K, V] {
	c := Map[K, V]{
		less: m.less,
		len:  m.len,
	}
	if m.root != nil {
		c.root = m.root.clone()
	}
	return &c
}

// NewMap creates new instance of ordered map.
func NewMap[K, V any](less func(K, K) bool) *Map[K, V] {
	return &Map[K, V]{less: less}
}

func (m *Map[K, V]) rebalance(n *Node[K, V]) {
	for {
		if b := n.balance(); b > 1 {
			if n.left.balance() < 0 {
				n.left = n.left.leftRotate()
			}
			n = n.rightRotate()
		} else if b < -1 {
			if n.right.balance() > 0 {
				n.right = n.right.rightRotate()
			}
			n = n.leftRotate()
		} else {
			n.recalcHeight()
		}
		if n.parent == nil {
			break
		}
		n = n.parent
	}
	m.root = n
}

func (n *Node[K, V]) clone() *Node[K, V] {
	c := Node[K, V]{
		key:    n.key,
		value:  n.value,
		height: n.height,
	}
	if n.left != nil {
		c.left = n.left.clone()
		c.left.parent = &c
	}
	if n.right != nil {
		c.right = n.right.clone()
		c.right.parent = &c
	}
	return &c
}

func (n *Node[K, V]) balance() int8 {
	var b int8
	if n.left != nil {
		b += n.left.height
	}
	if n.right != nil {
		b -= n.right.height
	}
	return b
}

func (n *Node[K, V]) recalcHeight() {
	if n.left != nil && n.right != nil {
		if n.left.height >= n.right.height {
			n.height = n.left.height + 1
		} else {
			n.height = n.right.height + 1
		}
	} else if n.left != nil {
		n.height = n.left.height + 1
	} else if n.right != nil {
		n.height = n.right.height + 1
	} else {
		n.height = 1
	}
}

// n        x
//
//	\      /
//	 x => n
//	/      \
//
// y        y
func (n *Node[K, V]) leftRotate() *Node[K, V] {
	x := n.right
	y := x.left
	x.left = n
	n.right = y
	if n.parent != nil {
		if n.parent.left == n {
			n.parent.left = x
		} else {
			n.parent.right = x
		}
	}
	x.parent = n.parent
	n.parent = x
	if y != nil {
		y.parent = n
	}
	n.recalcHeight()
	x.recalcHeight()
	return x
}

//	 n    x
//	/      \
//
// x   =>   n
//
//	\      /
//	 y    y
func (n *Node[K, V]) rightRotate() *Node[K, V] {
	x := n.left
	y := x.right
	if n.parent != nil {
		if n.parent.left == n {
			n.parent.left = x
		} else {
			n.parent.right = x
		}
	}
	x.right = n
	n.left = y
	x.parent = n.parent
	n.parent = x
	if y != nil {
		y.parent = n
	}
	n.recalcHeight()
	x.recalcHeight()
	return x
}

func getRoot[K, V any](n *Node[K, V]) *Node[K, V] {
	for n.parent != nil {
		n = n.parent
	}
	return n
}
