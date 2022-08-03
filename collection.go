package conditions

import (
	"fmt"
	"reflect"
)

type Collection interface {
	String() string
	Count() int
}

type NumberCollection interface {
	Has(number float64) bool
	String() string
}

type StringCollection interface {
	Has(val string) bool
	String() string
}

type MapNumberCollection struct {
	items map[float64]bool
}

func NewMapNumberCollectionFromMap(items map[float64]bool) Collection {
	return &MapNumberCollection{items: items}
}

func NewMapNumberCollection(items []interface{}) Collection {
	itemMap := make(map[float64]bool)

	for _, item := range items {
		switch i := item.(type) {
		case float64:
			itemMap[i] = true
		case float32:
			itemMap[float64(i)] = true
		case int64:
			itemMap[float64(i)] = true
		case int32:
			itemMap[float64(i)] = true
		case int:
			itemMap[float64(i)] = true
		}
	}

	return NewMapNumberCollectionFromMap(itemMap)
}

func (c *MapNumberCollection) Has(number float64) bool {
	_, exists := c.items[number]

	return exists
}

func (c *MapNumberCollection) String() string {
	str := "["

	for item := range c.items {
		str += fmt.Sprintf("%f,", item)
	}

	str += "]"

	return str
}

func (c *MapNumberCollection) Count() int {
	return len(c.items)
}

type MapStringCollection struct {
	items map[string]bool
}

func NewMapStringCollection(items []interface{}) Collection {
	strMap := make(map[string]bool)

	for _, item := range items {
		strMap[item.(string)] = true
	}

	return &MapStringCollection{
		items: strMap,
	}
}

func (c *MapStringCollection) Has(val string) bool {
	_, exists := c.items[val]

	return exists
}

func (c *MapStringCollection) String() string {
	str := "["

	for item := range c.items {
		str += fmt.Sprintf("%s,", item)
	}

	str += "]"

	return str
}

func (c *MapStringCollection) Count() int {
	return len(c.items)
}

func NewCollection(items []interface{}) (Collection, error) {
	for _, item := range items {
		switch item.(type) {
		case int:
			return NewMapNumberCollection(items), nil
		case int64:
			return NewMapNumberCollection(items), nil
		case int32:
			return NewMapNumberCollection(items), nil
		case float32:
			return NewMapNumberCollection(items), nil
		case float64:
			return NewMapNumberCollection(items), nil
		case string:
			return NewMapStringCollection(items), nil
		default:
			return nil, fmt.Errorf("collection type %s not supported", reflect.TypeOf(item).Kind())
		}
	}

	return nil, fmt.Errorf("collection items is empty")
}

func TryNewCollection(items []interface{}) Collection {
	collection, err := NewCollection(items)

	if err == nil {
		return collection
	}

	return nil
}
