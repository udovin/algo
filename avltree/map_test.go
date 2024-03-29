package avltree

import (
	"math/rand"
	"sync"
	"testing"
)

func intLess(x, y int) bool {
	return x < y
}

func TestSimpleIntMap(t *testing.T) {
	m := NewMap[int, int](intLess)
	n := 1000
	if v := m.Len(); v != 0 {
		t.Fatalf("Expected len = %d, got %d", 0, v)
	}
	for i := 0; i < n; i++ {
		m.Insert(i, i)
		if v := m.Len(); v != i+1 {
			t.Fatalf("Expected len = %d, got %d", i+1, v)
		}
	}
	for i := 0; i < n; i++ {
		it := m.Find(i)
		if it == nil {
			t.Fatalf("Value %d does not exist", i)
		}
		if v := it.Key(); v != i {
			t.Fatalf("Expected key = %d, got %d", i, v)
		}
		if v := it.Value(); v != i {
			t.Fatalf("Expected value = %d, got %d", i, v)
		}
	}
	for i := 0; i < n; i++ {
		if v, ok := m.Get(i); !ok || v != i {
			t.Fatalf("Invalid value: %d", v)
		}
	}
	{
		c := m.Clone()
		for i := 0; i < n; i++ {
			c.Set(i, i*i)
			if v, ok := c.Get(i); !ok || v != i*i {
				t.Fatalf("Invalid value: %d", v)
			}
		}
		for i := 0; i < n; i++ {
			c.Unset(i)
		}
		if v := c.Len(); v != 0 {
			t.Fatalf("Expected len = %d, got %d", 0, v)
		}
	}
	if it := m.Front(); it == nil {
		t.Fatalf("Expected non empty value")
	} else {
		if v := it.Key(); v != 0 {
			t.Fatalf("Expected key = %d, got %d", 0, v)
		}
	}
	if it := m.Back(); it == nil {
		t.Fatalf("Expected non empty value")
	} else {
		if v := it.Key(); v != n-1 {
			t.Fatalf("Expected key = %d, got %d", n-1, v)
		}
	}
	{
		prev := -1
		cnt := 0
		for it := m.Front(); it != nil; it = it.Next() {
			cnt++
			if v := it.Key(); v != prev+1 {
				t.Fatalf("Key out of order: %d != %d", v, prev+1)
			}
			prev = it.Key()
		}
		if cnt != n {
			t.Fatalf("Invalid len: %d != %d", cnt, 16)
		}
	}
	{
		prev := n
		cnt := 0
		for it := m.Back(); it != nil; it = it.Prev() {
			cnt++
			if v := it.Key(); v != prev-1 {
				t.Fatalf("Key out of order: %d != %d", v, prev-1)
			}
			prev = it.Key()
		}
		if cnt != n {
			t.Fatalf("Invalid len: %d != %d", cnt, n)
		}
	}
	{
		c := m.Clone()
		for it := c.Front(); it != nil; {
			jt := it.Next()
			c.Erase(it)
			it = jt
		}
		if v := c.Len(); v != 0 {
			t.Fatalf("Expected len = %d, got %d", 0, v)
		}
	}
	{
		c := m.Clone()
		for it := c.Front(); it != nil; {
			jt := it.Next()
			c.Erase(it)
			it = jt
		}
		if v := c.Len(); v != 0 {
			t.Fatalf("Expected len = %d, got %d", 0, v)
		}
	}
}

func testCheckBalance(tb testing.TB, n *Node[int, int]) {
	if n == nil {
		return
	}
	h := n.height
	var lh, rh int8 = 0, 0
	if n.left != nil {
		lh = n.left.height
	}
	if n.right != nil {
		rh = n.right.height
	}
	if lh < rh {
		if rh-lh > 1 {
			tb.Fatal("Tree is not balanced")
		}
		if rh+1 != h {
			tb.Fatal("Tree is not balanced")
		}
	} else {
		if lh-rh > 1 {
			tb.Fatal("Tree is not balanced")
		}
		if lh+1 != h {
			tb.Fatal("Tree is not balanced")
		}
	}
}

func TestRandomIntMap(t *testing.T) {
	m := NewMap[int, int](intLess)
	rnd := rand.New(rand.NewSource(42))
	n := 1000
	{
		p := rnd.Perm(n)
		for i := 0; i < n; i++ {
			it := m.Insert(p[i], i)
			if it == nil {
				t.Fatalf("Invalid insert (%d, %d)", p[i], i)
			}
			testCheckBalance(t, m.root)
		}
	}
	{
		p := rnd.Perm(n)
		for i := 0; i < n; i++ {
			it := m.Find(p[i])
			if it == nil {
				t.Fatalf("Unable to find key %d", p[i])
			}
			testCheckBalance(t, m.root)
			m.Erase(it)
		}
	}
}

func TestInvalidErase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic")
		}
	}()
	m1 := NewMap[int, int](intLess)
	m2 := NewMap[int, int](intLess)
	m2.Erase(m1.Insert(1, 2))
}

func BenchmarkAvltreeSimpleIntMapSeqSet(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
}

func BenchmarkAvltreeSimpleIntMapSeqGet(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
	b.ResetTimer()
	mu := sync.RWMutex{}
	for i := 0; i < b.N; i++ {
		mu.RLock()
		v, ok := m.Get(i)
		mu.RUnlock()
		if !ok {
			b.Fatalf("Unable to find key = %d", i)
		}
		if v != i {
			b.Fatalf("Expected value = %d, got %d", i, v)
		}
	}
}

func BenchmarkAvltreeSimpleIntMapSeqDelete(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Unset(i)
	}
}

func BenchmarkAvltreeSimpleIntMapSeqInsert(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Insert(i, i)
	}
}

func BenchmarkAvltreeSimpleIntMapSeqFind(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Insert(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if it := m.Find(i); it == nil {
			b.Fatalf("Unable to find key = %d", i)
		}
	}
}

func BenchmarkAvltreeSimpleIntMapSeqErase(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Insert(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := m.Find(i)
		if it == nil {
			b.Fatalf("Unable to find key = %d", i)
		}
		m.Erase(it)
	}
}

func BenchmarkAvltreeSimpleIntMapInsert(b *testing.B) {
	rnd := rand.New(rand.NewSource(42))
	b.ResetTimer()
	m := NewMap[int, int](intLess)
	p := rnd.Perm(b.N)
	for i := 0; i < b.N; i++ {
		m.Insert(p[i], i)
	}
}

func BenchmarkAvltreeSimpleIntMapFind(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Insert(i, i)
	}
	rnd := rand.New(rand.NewSource(42))
	b.ResetTimer()
	p := rnd.Perm(b.N)
	for i := 0; i < b.N; i++ {
		if it := m.Find(p[i]); it == nil {
			b.Fatalf("Unable to find key = %d", p[i])
		}
	}
}

func BenchmarkAvltreeSimpleIntMapErase(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Insert(i, i)
	}
	rnd := rand.New(rand.NewSource(42))
	b.ResetTimer()
	p := rnd.Perm(b.N)
	for i := 0; i < b.N; i++ {
		it := m.Find(p[i])
		if it == nil {
			b.Fatalf("Unable to find key = %d", p[i])
		}
		m.Erase(it)
	}
}
