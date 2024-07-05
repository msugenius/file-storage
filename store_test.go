package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "testing-directory"
	pathKey := CASPathTransformFunc(key)
	expectedPathName := "f1ac6\\54172\\49f50\\6bfc5\\92d93\\e9490\\b0e13\\ede59"
	expectedFilename := "f1ac65417249f506bfc592d93e9490b0e13ede59"
	if pathKey.Pathname != expectedPathName {
		t.Errorf("have %s want %s", pathKey.Pathname, expectedPathName)
	}
	if pathKey.Filename != expectedFilename {
		t.Errorf("have %s want %s", pathKey.Filename, expectedPathName)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := range 20 {
		key := fmt.Sprintf("some-key-%d", i)
		data := []byte("some bytes")

		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			t.Error(err)
		}

		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("expected to not have key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
		Root:              "test-root",
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
