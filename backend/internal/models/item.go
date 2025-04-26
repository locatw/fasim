package models

// Item represents a material or product that can be consumed or produced by facilities
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
