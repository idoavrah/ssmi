package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

type HistoryItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

type HistoryList struct {
	Items   []HistoryItem  `json:"items"`
	Indices map[string]int `json:"indices"`
	MaxSize int            `json:"maxSize"`
}

func NewHistoryList() *HistoryList {
	return &HistoryList{
		Items:   make([]HistoryItem, 0),
		Indices: make(map[string]int),
		MaxSize: 10,
	}
}

func (h *HistoryList) Add(item HistoryItem) {
	if idx, exists := h.Indices[item.ID]; exists {
		h.Items = append(h.Items[:idx], h.Items[idx+1:]...)
	} else if len(h.Items) == h.MaxSize {
		h.Items = h.Items[:len(h.Items)-1]
	}
	h.Items = append([]HistoryItem{item}, h.Items...)
	for i, item := range h.Items {
		h.Indices[item.ID] = i
	}
}

func (h *HistoryList) Contains(id string) bool {
	_, exists := h.Indices[id]
	return exists
}

func (h *HistoryList) Save(filename string) {

	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println("Error marshaling history list: ", err)
	}

	os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("Error saving history list: ", err)
	}
}

func LoadHistoryList(filename string) *HistoryList {
	data, _ := os.ReadFile(filename)
	list := NewHistoryList()
	json.Unmarshal(data, list)
	return list
}
