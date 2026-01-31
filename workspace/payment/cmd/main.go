package main

import (
	"context"
	"fmt"
	payment_v1 "github.com/DaniankaMarsianka42/boilerplates/workspace/shared/pkg/proto/payment/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const grpcPort = 50052

type PaymentServer struct {
	payment_v1.UnimplementedPaymentServiceServer

	mu sync.RWMutex
}

type PayReq struct {
	orderUuid     string
	userUuid      string
	paymentMethod PaymentMethod
}

type PaymentMethod int

func (p PaymentMethod) String() string {
	switch p {
	case Card:
		return "Card"
	case SBP:
		return "SBP"
	case CreditCard:
		return "CreditCard"
	case InvestorMoney:
		return "InvestorMoney"
	default:
		return "Unknown"
	}
}

const (
	Unknown = iota
	Card
	SBP
	CreditCard
	InvestorMoney
)

func (p *PaymentServer) PayOrder(_ context.Context, req *payment_v1.PayRequest) (*payment_v1.PayResponse, error) {

	payReq := PayReq{
		req.OrderUuid,
		req.UserUuid,
		PaymentMethod(int32(req.PaymentMethod)),
	}

	log.Printf("оплата от user: %v, к order: %v, способ оплаты: %v(%v)\n", payReq.userUuid, payReq.orderUuid, payReq.paymentMethod, int(payReq.paymentMethod))

	transactionUuid := uuid.New()

	log.Printf("Сгенерированный uuid по данной транзакции: %v\n", transactionUuid)

	return &payment_v1.PayResponse{
		TransactionUuid: transactionUuid.String(),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("Ошибка регистрации сервера: %v\n", err)
		return
	}

	//Создание gRPC сервера
	s := grpc.NewServer()

	//Регистрация нашего сервера
	service := &PaymentServer{}

	payment_v1.RegisterPaymentServiceServer(s, service)

	//Включаем рефлексию
	reflection.Register(s)

	go func() {
		log.Printf("🚀gRPC сервер поднят на порту %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("Сервер незапущен (fail): %v\n", err)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑Остановка gRPC сервера Shutting down...")
	s.GracefulStop()
	log.Println("✅ Сервер остановлен")

}
