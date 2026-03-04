package main

import (
	"context"
	"errors"
	order_v1 "github.com/DaniankaMarsianka42/boilerplates/workspace/shared/pkg/api/gen"
	inventory_v1 "github.com/DaniankaMarsianka42/boilerplates/workspace/shared/pkg/proto/inventory/v1"
	payment_v1 "github.com/DaniankaMarsianka42/boilerplates/workspace/shared/pkg/proto/payment/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	grpcPortInventory = "localhost:50051"
	grpcPortPayment   = "localhost:50052"
	httpPort          = "5050"
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

var OrderBD = make(map[string]Order)

type Order struct {
	userUuid        string
	transactionUuid string
	payMethod       PayMethod
	payment         Payment
	totalPrice      float32
}

type PayMethod int

const (
	UNKNOWN = iota
	CARD
	SBP
	CREDIT_CARD
	INVESTOR_MONEY
)

func (c PayMethod) String() string {
	switch c {
	case CARD:
		return "CARD"
	case SBP:
		return "SBP"
	case CREDIT_CARD:
		return "CREDIT_CARD"
	case INVESTOR_MONEY:
		return "INVESTOR_MONEY"
	default:
		return "unknown"
	}
}

type Payment int

const (
	unknown = iota
	PENDING_PAYMENT
	PAID
	CANCELLED
)

func (c Payment) String() string {
	switch c {
	case PENDING_PAYMENT:
		return "PENDING_PAYMENT"
	case PAID:
		return "PAID"
	case CANCELLED:
		return "CANCELLED"
	default:
		return "unknown"
	}
}

type OrderStorage struct {
	mu     sync.RWMutex
	orders map[string][]string
}

func NewOrderStorage() *OrderStorage {
	return &OrderStorage{
		orders: make(map[string][]string),
	}
}

type OrderHandler struct {
	storage         *OrderStorage
	inventoryClient inventory_v1.InventoryServiceClient
	paymentClient   payment_v1.PaymentServiceClient
}

func NewOrderHandler(storage *OrderStorage, inventory_client inventory_v1.InventoryServiceClient,
	payment_client payment_v1.PaymentServiceClient) *OrderHandler {
	return &OrderHandler{
		storage:         storage,
		inventoryClient: inventory_client,
		paymentClient:   payment_client,
	}
}

func (h *OrderHandler) CreateOrderByCity(ctx context.Context, req *order_v1.CreateOrderRequest) (order_v1.CreateOrderByCityRes, error) {
	uuids, err := h.inventoryClient.ListPart(ctx, &inventory_v1.ListRequest{Uuid: req.PartUUID})
	if err != nil || uuids == nil {
		return &order_v1.NotFoundError{
			Code:    404,
			Message: "Uuid for part '" + req.UserUUID + "' not found",
		}, nil
	}

	orderUuid := uuid.New()
	OrderBD[orderUuid.String()] = Order{
		userUuid:   req.UserUUID,
		payment:    PENDING_PAYMENT,
		totalPrice: float32(len(uuids.GetPart())) * 0.5,
	}

	return &order_v1.CreateOrder200{
		OrderUUID:  orderUuid.String(),
		TotalPrice: OrderBD[orderUuid.String()].totalPrice,
	}, nil
}

func (h *OrderHandler) PayOrder(ctx context.Context, req *order_v1.PaymentMethods, params order_v1.PayOrderParams) (order_v1.PayOrderRes, error) {
	log.Printf("оплата заказа с uuid: %v\n", params.UUID)

	order, ok := OrderBD[params.UUID]
	if ok != true {
		return &order_v1.NotFoundError{
			Code:    404,
			Message: "Order uuid '" + params.UUID + "' was not found",
		}, nil
	}

	if order.payment != PENDING_PAYMENT {
		return &order_v1.BadRequestError{
			Code:    400,
			Message: "Order uuid '" + params.UUID + "' уже оплачен или отменен",
		}, nil
	}

	transactionsUuid, err := h.paymentClient.PayOrder(ctx, &payment_v1.PayRequest{
		OrderUuid: params.UUID,
		UserUuid:  order.userUuid,
	})
	if err != nil {
		return &order_v1.BadRequestError{
			Code:    400,
			Message: "Order uuid '" + params.UUID + "' уже оплачен или отменен",
		}, nil
	}

	order.transactionUuid = transactionsUuid.TransactionUuid
	order.payment = PAID

	return &order_v1.RequestPayOrder{
		TransactionUUID: order.transactionUuid,
	}, nil
}

func (h *OrderHandler) NewError(_ context.Context, err error) *order_v1.GenericErrorStatusCode {
	return &order_v1.GenericErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: order_v1.GenericError{
			Code:    order_v1.NewOptInt(http.StatusInternalServerError),
			Message: order_v1.NewOptString(err.Error()),
		},
	}
}

func main() {
	ctx := context.Background()

	inventConn, err := grpc.NewClient(
		grpcPortInventory,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer func() {
		if cerr := inventConn.Close(); cerr != nil {
			log.Printf("failed to close connect: %v", cerr)
		}
	}()

	payConn, err := grpc.NewClient(
		grpcPortPayment,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer func() {
		if cerr := payConn.Close(); cerr != nil {
			log.Printf("failed to close connect: %v", cerr)
		}
	}()

	inventoryClient := inventory_v1.NewInventoryServiceClient(inventConn)
	paymentClient := payment_v1.NewPaymentServiceClient(payConn)

	storage := NewOrderStorage()

	orderHandler := NewOrderHandler(storage, inventoryClient, paymentClient)

	orderServer, err := order_v1.NewServer(orderHandler)
	if err != nil {
		log.Fatalf("ошибка создания сервера OpenAPI: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	//r.Use(customMiddleware.RequestLogger)

	r.Mount("/", orderServer)

	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}

	go func() {
		log.Printf("🚀 HTTP-сервер запущен на порту %s\n", httpPort)
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("❌ Ошибка запуска сервера: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("❌ Ошибка при остановке сервера: %v\n", err)
	}

	log.Println("✅ Сервер остановлен")

}
