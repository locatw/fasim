package models

// Item represents a production item
type Item struct {
	id          int
	name        string
	description string
}

// NewItem creates a new Item
func NewItem(name, description string) *Item {
	return &Item{
		name:        name,
		description: description,
	}
}

// NewItemFromParams creates an item with all parameters specified.
// Use this function only when creating objects from persisted data, and use NewItem() for other purposes.
func NewItemFromParams(id int, name string, description string) *Item {
	return &Item{
		id:          id,
		name:        name,
		description: description,
	}
}

// ID returns the item's ID
func (i *Item) ID() int {
	return i.id
}

// Name returns the item's name
func (i *Item) Name() string {
	return i.name
}

// Description returns the item's description
func (i *Item) Description() string {
	return i.description
}
