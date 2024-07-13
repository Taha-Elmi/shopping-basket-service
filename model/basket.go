package model

import (
	"gorm.io/gorm"
)

type Basket struct {
	gorm.Model
	Data    string `json:"data"`
	State   string `json:"state"`
	OwnerID uint   `json:"owner_id"`
}
