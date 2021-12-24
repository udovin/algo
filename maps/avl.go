package maps

type avlNode[K, V any] struct {
	key    K
	value  V
	height int
	left   *avlNode[K, V]
	right  *avlNode[K, V]
	parent *avlNode[K, V]
}

func (n *avlNode[K, V]) next() *avlNode[K, V] {
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

func (n *avlNode[K, V]) prev() *avlNode[K, V] {
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

func (n *avlNode[K, V]) clone() *avlNode[K, V] {
	c := avlNode[K, V]{
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

func (n *avlNode[K, V]) balance() int {
	b := 0
	if n.left != nil {
		b += n.left.height
	}
	if n.right != nil {
		b -= n.right.height
	}
	return b
}

func (n *avlNode[K, V]) recalcHeight() {
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
//  \      /
//   x => n
//  /      \
// y        y
func (n *avlNode[K, V]) leftRotate() *avlNode[K, V] {
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

//   n    x
//  /      \
// x   =>   n
//  \      /
//   y    y
func (n *avlNode[K, V]) rightRotate() *avlNode[K, V] {
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

type avlTree[K, V any] struct {
	root *avlNode[K, V]
	less func(K, K) bool
	len  int
}

func (t *avlTree[K, V]) insert(key K, value V) *avlNode[K, V] {
	n := avlNode[K, V]{key: key, value: value, height: 1}
	if t.root == nil {
		t.root = &n
		t.len = 1
		return &n
	}
	p := t.root
	for {
		if t.less(key, p.key) {
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
	t.len++
	t.rebalance(p)
	return &n
}

func (t *avlTree[K, V]) rebalance(n *avlNode[K, V]) {
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
	t.root = n
}

func (t *avlTree[K, V]) erase(n *avlNode[K, V])  {
	if n.left == nil && n.right == nil {
		if n.parent == nil {
			t.root = nil
			t.len = 0
			return
		}
		if n.parent.left == n {
			n.parent.left = nil
		} else {
			n.parent.right = nil
		}
		t.len--
		t.rebalance(n.parent)
		n.parent = nil
		return
	}
	if n.left == nil {
		if n.parent == nil {
			n.right.parent = nil
			t.root = n.right
			t.len--
			n.right = nil
			return
		}
		if n.parent.left == n {
			n.parent.left = n.right
		} else {
			n.parent.right = n.right
		}
		n.right.parent = n.parent
		t.len--
		t.rebalance(n.parent)
		n.right = nil
		n.parent = nil
		return
	}
	if n.right == nil {
		if n.parent == nil {
			n.left.parent = nil
			t.root = n.left
			t.len--
			n.left = nil
			return
		}
		if n.parent.left == n {
			n.parent.left = n.left
		} else {
			n.parent.right = n.left
		}
		n.left.parent = n.parent
		t.len--
		t.rebalance(n.parent)
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
		t.len--
		t.rebalance(r)
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
	t.len--
	t.rebalance(p)
	n.left = nil
	n.right = nil
	n.parent = nil
}

func (t *avlTree[K, V]) find(key K) *avlNode[K, V] {
	n := t.lowerBound(key)
	if n == nil || t.less(key, n.key) {
		return nil
	}
	return n
}

func (t *avlTree[K, V]) lowerBound(key K) (n *avlNode[K, V]) {
	for it := t.root; it != nil; {
		if t.less(it.key, key) {
			it = it.right
		} else {
			n = it
			it = it.left
		}
	}
	return
}

func (t *avlTree[K, V]) front() *avlNode[K, V] {
	if t.root == nil {
		return nil
	}
	it := t.root
	for it.left != nil {
		it = it.left
	}
	return it
}

func (t *avlTree[K, V]) back() *avlNode[K, V] {
	if t.root == nil {
		return nil
	}
	it := t.root
	for it.right != nil {
		it = it.right
	}
	return it
}

func (t *avlTree[K, V]) clone() *avlTree[K, V] {
	c := avlTree[K, V]{
		less: t.less,
		len:  t.len,
	}
	if t.root != nil {
		c.root = t.root.clone()
	}
	return &c
}
