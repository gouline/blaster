package scache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHit(t *testing.T) {
	cache := New(5*time.Minute, 10*time.Minute)
	key := "test1"
	value := "value1"

	counter := 0
	retriever := func(key string) (interface{}, error) {
		counter++
		return value, nil
	}
	response := <-cache.ResponseChan(key, retriever)
	assert.Equal(t, value, response.Value)
	assert.Equal(t, 1, counter)

	response = <-cache.ResponseChan(key, retriever)
	assert.Equal(t, value, response.Value)
	assert.Equal(t, 1, counter)
}

func TestExpire(t *testing.T) {
	cache := New(50*time.Millisecond, 100*time.Millisecond)
	key := "test2"
	value := "value2"

	counter := 0
	retriever := func(key string) (interface{}, error) {
		counter++
		return value, nil
	}
	response := <-cache.ResponseChan(key, retriever)
	assert.Equal(t, value, response.Value)
	assert.Equal(t, 1, counter)

	time.Sleep(100 * time.Millisecond)

	response = <-cache.ResponseChan(key, retriever)
	assert.Equal(t, value, response.Value)
	assert.Equal(t, 2, counter)
}
