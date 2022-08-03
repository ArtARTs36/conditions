package conditions

import "fmt"

type NumberCollection interface {
	Has(number float64) bool
	String() string
}

type MapNumberCollection struct {
	items map[float64]bool
}

func NewMapNumberCollection(items map[float64]bool) MapNumberCollection {
	return MapNumberCollection{items: items}
}

func NewMapNumberCollectionFromList(items []interface{}) MapNumberCollection {
	itemsMap := make(map[float64]bool)

	for _, val := range items {
		itemsMap[val.(float64)] = true
	}

	return NewMapNumberCollection(itemsMap)
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
