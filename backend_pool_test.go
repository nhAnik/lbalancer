package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundRobin(t *testing.T) {
	t.Run("3 backends", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 1)
		b2, _ := newBackend("def.com", 1)
		b3, _ := newBackend("xyz.com", 1)
		rrp := newRoundRobinPool([]*Backend{b1, b2, b3})

		n1 := rrp.getNext()
		n2 := rrp.getNext()
		n3 := rrp.getNext()
		assert.ElementsMatch(t, []*Backend{n1, n2, n3}, []*Backend{b1, b2, b3})
	})

	t.Run("single backend", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 1)
		rrp := newRoundRobinPool([]*Backend{b1})
		assert.Equal(t, b1, rrp.getNext())
	})
}

func TestWeightedRoundRobin(t *testing.T) {
	t.Run("weighted 5-1-1", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 5)
		b2, _ := newBackend("def.com", 1)
		b3, _ := newBackend("xyz.com", 1)
		wrp := newWeightedRoundRobinPool([]*Backend{b1, b2, b3})

		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b2, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
	})

	t.Run("weighted 2-2-2", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 2)
		b2, _ := newBackend("def.com", 2)
		b3, _ := newBackend("xyz.com", 2)
		wrp := newWeightedRoundRobinPool([]*Backend{b1, b2, b3})

		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b2, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b2, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
	})

	t.Run("weighted 2-1-3-2", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 2)
		b2, _ := newBackend("def.com", 1)
		b3, _ := newBackend("pqr.com", 3)
		b4, _ := newBackend("xyz.com", 2)
		wrp := newWeightedRoundRobinPool([]*Backend{b1, b2, b3, b4})

		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b4, wrp.getNext())
		assert.Equal(t, b2, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b4, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
	})

	t.Run("single backend", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 2)
		wrp := newWeightedRoundRobinPool([]*Backend{b1})

		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
	})

	t.Run("weighted 2-1-3", func(t *testing.T) {
		b1, _ := newBackend("abc.com", 2)
		b2, _ := newBackend("def.com", 1)
		b3, _ := newBackend("pqr.com", 3)
		wrp := newWeightedRoundRobinPool([]*Backend{b1, b2, b3})

		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b2, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())

		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b2, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
		assert.Equal(t, b1, wrp.getNext())
		assert.Equal(t, b3, wrp.getNext())
	})
}
