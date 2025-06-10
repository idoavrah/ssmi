package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

type FavoriteItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Profile  string `json:"profile"`
}

type FavoritesArray struct {
	Items [10]FavoriteItem `json:"items"`
}

func NewFavoritesList() *FavoritesArray {
	return &FavoritesArray{}
}

func (h *FavoritesArray) Add(item FavoriteItem, position int) {

	if position < 0 || position > 9 {
		return
	}

	for i := 0; i < len(h.Items); i++ {
		if h.Items[i] == item && i != position {
			h.Items[i] = FavoriteItem{}
		}
	}

	h.Items[position] = item
}

func (h *FavoritesArray) Save(filename string) {

	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println("Error marshaling favorites list: ", err)
	}

	os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("Error saving favorites list: ", err)
	}
}

func LoadFavoritesList(filename string) *FavoritesArray {
	data, _ := os.ReadFile(filename)
	list := NewFavoritesList()
	json.Unmarshal(data, list)
	return list
}
