package freecache

import (
	"github.com/coocood/freecache"
	"github.com/slava-911/test-task-0723/pkg/cache"
)

type iterator struct {
	iter *freecache.Iterator
}

func (i *iterator) Next() *cache.Entry {
	entry := i.iter.Next()
	if entry == nil {
		return nil
	}

	return &cache.Entry{
		Key:   entry.Key,
		Value: entry.Value,
	}
}
