package main

import (
	"time"

	inventory_v1 "github.com/DaniankaMarsianka42/boilerplates/workspace/shared/pkg/proto/inventory/v1"
)

type Part struct {
	uuid           string
	name           string
	description    string
	price          float64
	stock_quantity int64
	category       Category
	dimensions     Dimensions
	manufacturer   Manufacturer
	tags           []string
	metadata       map[string]any
	created_at     time.Time
	update_at      time.Time
}

type Manufacturer struct {
	name    string
	country string
	website string
}

type Dimensions struct {
	length float64
	wigth  float64
	height float64
	weight float64
}

type Category int

const (
	unknown = iota
	engine
	fuel
	porthole
	wing
)

func (c Category) String() string {
	switch c {
	case engine:
		return "Engine"
	case fuel:
		return "Fuel"
	case porthole:
		return "Porthole"
	case wing:
		return "Wing"
	default:
		return "unknown"
	}
}

var BD = make(map[string]Part)

const grpcPort = 5050

type inventoryService struct {
	inventory_v1.UnimplementedInventoryServiceServer
}

func main() {

}
