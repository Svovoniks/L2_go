package main

import "fmt"

type PaymentStrategy interface {
	Pay(amout float32)
}

type BankCard struct {
}

func (c *BankCard) Pay(amount float32) {
	fmt.Println("Bank card payment: ", amount)
}

type Cash struct {
}

func (c *Cash) Pay(amount float32) {
	fmt.Println("Cash payment: ", amount)
}

type QRPay struct {
}

func (c *QRPay) Pay(amount float32) {
	fmt.Println("QR payment: ", amount)
}

type PaymentEngine struct {
	strat PaymentStrategy
}

func (e *PaymentEngine) SetStrategy(strat PaymentStrategy) {
	e.strat = strat
}

func (e *PaymentEngine) ProcessPayent(amount float32) {
	e.strat.Pay(amount)
}

func main() {
	engine := &PaymentEngine{}

	engine.SetStrategy(&BankCard{})
	engine.ProcessPayent(42)

	engine.SetStrategy(&Cash{})
	engine.ProcessPayent(100)

	engine.SetStrategy(&QRPay{})
	engine.ProcessPayent(4)
}
