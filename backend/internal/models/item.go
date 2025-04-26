package models

// Item は生産アイテムを表す構造体です
type Item struct {
	id          int
	name        string
	description string
}

// NewItem は新しいItemを生成します
func NewItem(name string) *Item {
	return &Item{
		name: name,
	}
}

// NewItemFromParams はパラメータからItemを生成します
func NewItemFromParams(id int, name string, description string) *Item {
	return &Item{
		id:          id,
		name:        name,
		description: description,
	}
}

// ID はアイテムのIDを返します
func (i *Item) ID() int {
	return i.id
}

// Name はアイテムの名前を返します
func (i *Item) Name() string {
	return i.name
}

// Description はアイテムの説明を返します
func (i *Item) Description() string {
	return i.description
}
