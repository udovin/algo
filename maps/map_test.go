package maps

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
	for i := 0; i < 16; i++ {
		m.Insert(i, i)
		if v := m.Len(); v != i + 1 {
			t.Fatalf("Expected len = %d, got %d", i + 1, v)
		}
	}
	for i := 0; i < 16; i++ {
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
	for i := 0; i < 16; i++ {
		if v, ok := m.Get(i); !ok || v != i {
			t.Fatalf("Invalid value: %d", v)
		}
	}
	{
		c := m.Clone()
		for i := 0; i < 16; i++ {
			c.Set(i, i * i)
			if v, ok := c.Get(i); !ok || v != i * i {
				t.Fatalf("Invalid value: %d", v)
			}
		}
		for i := 0; i < 16; i++ {
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
		if v := it.Key(); v != 15 {
			t.Fatalf("Expected key = %d, got %d", 15, v)
		}
	}
	{
		prev := -1
		cnt := 0
		for it := m.Front(); it != nil; it = it.Next() {
			cnt++
			if v := it.Key(); v != prev + 1 {
				t.Fatalf("Key out of order: %d != %d", v, prev + 1)
			}
			prev = it.Key()
		}
		if cnt != 16 {
			t.Fatalf("Invalid len: %d != %d", cnt, 16)
		}
	}
	{
		prev := 16
		cnt := 0
		for it := m.Back(); it != nil; it = it.Prev() {
			cnt++
			if v := it.Key(); v != prev - 1 {
				t.Fatalf("Key out of order: %d != %d", v, prev - 1)
			}
			prev = it.Key()
		}
		if cnt != 16 {
			t.Fatalf("Invalid len: %d != %d", cnt, 16)
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
