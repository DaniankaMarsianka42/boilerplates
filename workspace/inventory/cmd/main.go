package main

import (
	"context"
	"fmt"
	inventory_v1 "github.com/DaniankaMarsianka42/boilerplates/workspace/shared/pkg/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Part struct {
	Uuid           string
	Name           string
	Description    string
	Price          float64
	Stock_quantity int64
	Category       Category
	Dimensions     Dimensions
	Manufacturer   Manufacturer
	Tags           []string
	Metadata       map[string]any
	Created_at     time.Time
	Update_at      time.Time
}

type Manufacturer struct {
	Name    string
	Country string
	Website string
}

type Dimensions struct {
	Length float64
	Width  float64
	Height float64
	Weight float64
}

type Category int

const (
	unknown = iota
	Engine
	Fuel
	Porthole
	Wing
)

func (c Category) String() string {
	switch c {
	case Engine:
		return "Engine"
	case Fuel:
		return "Fuel"
	case Porthole:
		return "Porthole"
	case Wing:
		return "Wing"
	default:
		return "unknown"
	}
}

var BD = make(map[string]*Part)

const grpcPort = 50051

func CmdPartToProtoPart(cmdPart *Part) *inventory_v1.Part {
	return &inventory_v1.Part{
		Uuid:        cmdPart.Uuid,
		Name:        cmdPart.Name,
		Description: cmdPart.Description,
		Price:       cmdPart.Price,
		Category:    inventory_v1.Category(cmdPart.Category),
		Dimensions: &inventory_v1.Dimensions{
			Length: cmdPart.Dimensions.Length,
			Wight:  cmdPart.Dimensions.Width,
			Height: cmdPart.Dimensions.Height,
			Weight: cmdPart.Dimensions.Weight,
		},
		Manufacturer: &inventory_v1.Manufacturer{
			Name:    cmdPart.Manufacturer.Name,
			Country: cmdPart.Manufacturer.Country,
			Website: cmdPart.Manufacturer.Website,
		},
		Tags:      cmdPart.Tags,
		CreatedAt: timestamppb.New(cmdPart.Created_at),
		UpdateAt:  timestamppb.New(cmdPart.Update_at),
	}
}

type inventoryService struct {
	inventory_v1.UnimplementedInventoryServiceServer

	mu   sync.RWMutex
	part Part
}

func (s *inventoryService) GetPart(_ context.Context, req *inventory_v1.GetRequest) (*inventory_v1.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	part, ok := BD[req.GetUuid()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "part with uuid %s not found", req.GetUuid())
	}

	return &inventory_v1.GetResponse{
		Part: CmdPartToProtoPart(part),
	}, nil

}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("Ошибка регестрации сервера: %v\n", err)
		return
	}

	addBD()

	//Создание gRPC сервера
	s := grpc.NewServer()

	//Регистрация нашего сервера
	service := &inventoryService{}

	inventory_v1.RegisterInventoryServiceServer(s, service)

	//Включаем рефлексию(чтобы postmen видел какие методы у нас есть)
	reflection.Register(s)

	go func() {
		log.Printf("🚀gRPC сервер поднят на порту %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("Сервер незапущен (fail): %v\n", err)
			return
		}
	}()

	//Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑Остановка gRPC сервера Shutting down...")
	s.GracefulStop()
	log.Println("✅ Сервер остановлен")
}

func addBD() {
	BD["550e8400-e29b-41d4-a716-446655440001"] = &Part{
		Uuid:           "550e8400-e29b-41d4-a716-446655440001",
		Name:           "Двигатель Pratt & Whitney PW4000",
		Description:    "Турбовентиляторный двигатель для Boeing 777",
		Price:          1250000.50,
		Stock_quantity: 2,
		Category:       Category(2),
		Dimensions: Dimensions{ // предполагаемая структура
			Length: 3.5,
			Width:  2.8,
			Height: 4.1,
			Weight: 5.1,
		},
		Manufacturer: Manufacturer{ // предполагаемая структура
			Name:    "Pratt & Whitney",
			Country: "USA",
			Website: "http/idinahu",
		},
		Tags: []string{"engine", "turbofan", "boeing777", "aviation"},
		Metadata: map[string]any{
			"thrust_kn":    400.0,
			"bypass_ratio": 5.0,
			"certified":    true,
		},
		Created_at: time.Now().Add(-30 * 24 * time.Hour),
		Update_at:  time.Now(),
	}
}
