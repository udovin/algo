package btree

import (
	"testing"
)

func intLess(x, y int) bool {
	return x < y
}

func TestSimpleIntMap(t *testing.T) {
	m := NewMap[int, int](intLess)
	if v := m.Len(); v != 0 {
		t.Fatalf("Expected len = %d, got %d", 0, v)
	}
	for i := 0; i < 128; i++ {
		m.Set(i, i)
		if v := m.Len(); v != i+1 {
			t.Fatalf("Expected len = %d, got %d", i+1, v)
		}
	}
	for i := 0; i < 128; i++ {
		v, ok := m.Get(i)
		if !ok {
			t.Fatalf("Value %d does not exist", i)
		}
		if v != i {
			t.Fatalf("Expected value = %d, got %d", i, v)
		}
	}
	{
		it := m.Iter()
		for i := 0; i < 128; i++ {
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
		for i := 127; i >= 0; i-- {
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
	for i := 0; i < 128; i++ {
		m.Delete(i)
		if v := m.Len(); v != 127-i {
			t.Fatalf("Expected len = %d, got %d", 127-i, v)
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
