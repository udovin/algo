package btree

import (
	"math/rand"
	"testing"
)

func intLess(x, y int) bool {
	return x < y
}

func TestSimpleIntMap(t *testing.T) {
	m := NewMap[int, int](intLess)
	n := 3000
	if v := m.Len(); v != 0 {
		t.Fatalf("Expected len = %d, got %d", 0, v)
	}
	for i := 0; i < n; i++ {
		m.Set(i, i)
		if v := m.Len(); v != i+1 {
			t.Fatalf("Expected len = %d, got %d", i+1, v)
		}
	}
	for i := 0; i < 3; i++ {
		m.Set(n, i)
		v, ok := m.Get(n)
		if !ok {
			t.Fatalf("Key %d does not exist", i)
		}
		if v != i {
			t.Fatalf("Expected value = %d, got %d", i, v)
		}
	}
	for i := 0; i < 3; i++ {
		m.Delete(n)
	}
	for i := 0; i < n; i++ {
		v, ok := m.Get(i)
		if !ok {
			t.Fatalf("Key %d does not exist", i)
		}
		if v != i {
			t.Fatalf("Expected value = %d, got %d", i, v)
		}
	}
	{
		it := m.Iter()
		for i := 0; i < n; i++ {
			if !it.Next() {
				t.Fatal("Unexpected end of iter")
			}
			if v := it.Key(); v != i {
				t.Fatalf("Expected key = %d, got %d", i, v)
			}
			if v := it.Value(); v != i {
				t.Fatalf("Expected value = %d, got %d", i, v)
			}
		}
		if it.Next() {
			t.Fatal("Iter should be ended", it.Value())
		}
	}
	{
		it := m.Iter()
		for i := n - 1; i >= 0; i-- {
			if !it.Prev() {
				t.Fatal("Unexpected end of iter")
			}
			if v := it.Key(); v != i {
				t.Fatalf("Expected key = %d, got %d", i, v)
			}
			if v := it.Value(); v != i {
				t.Fatalf("Expected value = %d, got %d", i, v)
			}
		}
		if it.Prev() {
			t.Fatal("Iter should be ended", it.Value())
		}
	}
	{
		it := m.Iter()
		for i := 0; i < n; i++ {
			if !it.Seek(i) {
				t.Fatal("Unexpected end of iter")
			}
			if v := it.Key(); v != i {
				t.Fatalf("Expected key = %d, got %d", i, v)
			}
			if v := it.Value(); v != i {
				t.Fatalf("Expected value = %d, got %d", i, v)
			}
		}
		if it.Seek(n) {
			t.Fatal("Iter should be ended", it.Value())
		}
	}
	for i := 0; i < n; i++ {
		m.Delete(i)
		if v := m.Len(); v != n-i-1 {
			t.Fatalf("Expected len = %d, got %d", n-i-1, v)
		}
	}
}

func TestRandomIntMap(t *testing.T) {
	m := NewMap[int, int](intLess)
	rnd := rand.New(rand.NewSource(42))
	n := 3000
	{
		p := rnd.Perm(n)
		for i := 0; i < n; i++ {
			m.Set(p[i], i)
			if v := m.Len(); v != i+1 {
				t.Fatalf("Expected len = %d, got %d", i+1, v)
			}
		}
		for i := 0; i < n; i++ {
			v, ok := m.Get(p[i])
			if !ok {
				t.Fatalf("Key %d does not exist", i)
			}
			if v != i {
				t.Fatalf("Expected value = %d, got %d", i, v)
			}
		}
		if _, ok := m.Get(n); ok {
			t.Fatalf("Key %d should not exist", n)
		}
	}
	{
		p := rnd.Perm(n)
		for i := 0; i < n; i++ {
			m.Delete(p[i])
			if v := m.Len(); v != n-i-1 {
				t.Fatalf("Expected len = %d, got %d", n-i-1, v)
			}
		}
		m.Delete(0)
		if _, ok := m.Get(n); ok {
			t.Fatalf("Key %d should not exist", n)
		}
	}
}

func BenchmarkSimpleIntMapSeqSet(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
}

func BenchmarkSimpleIntMapSeqGet(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, ok := m.Get(i)
		if !ok {
			b.Fatalf("Unable to find key = %d", i)
		}
		if v != i {
			b.Fatalf("Expected value = %d, got %d", i, v)
		}
	}
}

func BenchmarkSimpleIntMapSeqDelete(b *testing.B) {
	m := NewMap[int, int](intLess)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(i)
	}
}
