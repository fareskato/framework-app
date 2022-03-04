package cache

import (
	"bytes"
	"encoding/gob"
)

type Cache interface {
	Has(string) (bool, error)
	Get(string) (interface{}, error)
	Set(string, interface{}, ...int) error
	Forget(string) error
	EmptyByMatch(string) error
	Empty() error
}

type EntryCache map[string]interface{}

func encodeCache(item EntryCache) ([]byte, error) {
	buff := bytes.Buffer{}
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(&item)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func decodeCache(s string) (EntryCache, error) {
	ec := EntryCache{}
	buff := bytes.Buffer{}
	buff.Write([]byte(s))
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&ec)
	if err != nil {
		return nil, err
	}
	return ec, nil
}
