package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_SetAndGet(t *testing.T) {
	s := NewStorage()

	s.Set("foo", "bar")
	value, ok := s.Get("foo")

	assert.True(t, ok)
	assert.Equal(t, "bar", value.Data)
}

func TestStorage_GetNonExistentKey(t *testing.T) {
	s := NewStorage()

	value, ok := s.Get("missing")

	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestStorage_Delete(t *testing.T) {
	s := NewStorage()
	s.Set("foo", "bar")

	deleted := s.Delete("foo")
	assert.True(t, deleted)

	_, ok := s.Get("foo")
	assert.False(t, ok)
}

func TestStorage_DeleteNonExistent(t *testing.T) {
	s := NewStorage()

	deleted := s.Delete("missing")
	assert.False(t, deleted)
}

func TestStorage_Overwrite(t *testing.T) {
	s := NewStorage()
	s.Set("foo", "bar")
	s.Set("foo", "value")

	value, ok := s.Get("foo")

	assert.True(t, ok)
	assert.Equal(t, "value", value.Data)
}
