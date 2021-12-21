package maps

// splayNode prepresents splay tree splayNode.
type splayNode[K, V any] struct {
	key                 K
	value               V
	parent, left, right *splayNode[K, V]
}

// Next returns next splayNode.
//
// If current splayNode is the last, Next will return nil.
func (n *splayNode[K, V]) next() *splayNode[K, V] {
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

// Prev returns prevuous splayNode.
//
// If current splayNode is the first, Prev will return nil.
func (n *splayNode[K, V]) prev() *splayNode[K, V] {
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

func (n *splayNode[K, V]) rotate() {
	p := n.parent
	if p.left == n {
		p.left = n.right
		if p.left != nil {
			p.left.parent = p
		}
		n.right = p
	} else {
		p.right = n.left
		if p.right != nil {
			p.right.parent = p
		}
		n.left = p
	}
	g := p.parent
	if g != nil {
		if g.left == p {
			g.left = n
		} else {
			g.right = n
		}
	}
	p.parent = n
	n.parent = g
}

func (n *splayNode[K, V]) splay() {
	for n.parent != nil {
		p := n.parent
		g := p.parent
		if g != nil {
			if (g.left == p) == (p.left == n) {
				p.rotate()
			} else {
				n.rotate()
			}
		}
		n.rotate()
	}
}

type splayTree[K, V any] struct {
	less  func(K, K) bool
	root *splayNode[K, V]
	len  int
}

func NewsplayTree[K, V any](less func(K, K) bool) *splayTree[K, V] {
	return &splayTree[K, V]{less: less}
}

// split splits splay tree into two parts by key.
//
// Current tree will have values < key.
// New part will have values >= key.
func (t *splayTree[K, V]) split(key K) *splayNode[K, V] {
	n := t.lowerBound(key)
	if n == nil {
		return nil
	}
	t.root = n.left
	if n.left != nil {
		n.left.parent = nil
		n.left = nil
	}
	return n
}

func (t *splayTree[K, V]) merge(n *splayNode[K, V]) {
	if n == nil {
		return
	}
	for n.left != nil {
		n = n.left
	}
	n.splay()
	n.left = t.root
	if n.left != nil {
		n.left.parent = n
	}
	t.root = n
}

func (t *splayTree[K, V]) insert(key K, value V) *splayNode[K, V] {
	n := &splayNode[K, V]{key: key, value: value}
	r := t.split(key)
	n.left, n.right = t.root, r
	if n.left != nil {
		n.left.parent = n
	}
	if n.right != nil {
		n.right.parent = n
	}
	t.root = n
	t.len++
	return n
}

func (t *splayTree[K, V]) erase(n *splayNode[K, V]) {
	n.splay()
	t.root = n.left
	if t.root != nil {
		t.root.parent = nil
	}
	r := n.right
	if r != nil {
		r.parent = nil
	}
	t.merge(r)
	t.len--
}

func (t *splayTree[K, V]) find(key K) *splayNode[K, V] {
	n := t.lowerBound(key)
	if n == nil || t.less(key, n.key) {
		return nil
	}
	return n
}

func (t *splayTree[K, V]) front() *splayNode[K, V] {
	if t.root == nil {
		return nil
	}
	it := t.root
	for it.left != nil {
		it = it.left
	}
	return it
}

func (t *splayTree[K, V]) back() *splayNode[K, V] {
	if t.root == nil {
		return nil
	}
	it := t.root
	for it.right != nil {
		it = it.right
	}
	return it
}

// LowerBound returns smallest splayNode with key >= K.
func (t *splayTree[K, V]) lowerBound(key K) (n *splayNode[K, V]) {
	for it := t.root; it != nil; {
		if t.less(it.key, key) {
			it = it.right
		} else {
			n = it
			it = it.left
		}
	}
	if n != nil {
		n.splay()
		t.root = n
	}
	return
}

// Clone clones splayTree.
func (t *splayTree[K, V]) clone() *splayTree[K, V] {
	return &splayTree[K,V]{
		less: t.less,
		root: cloneSplayNode(t.root),
		len:  t.len,
	}
}

func cloneSplayNode[K, V any](n *splayNode[K, V]) *splayNode[K, V] {
	if n == nil {
		return nil
	}
	c := &splayNode[K, V]{key: n.key, value: n.value}
	c.left = cloneSplayNode(n.left)
	if c.left != nil {
		c.left.parent = c
	}
	c.right = cloneSplayNode(n.right)
	if c.right != nil {
		c.right.parent = c
	}
	return c
}
