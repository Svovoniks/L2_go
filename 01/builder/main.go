package main

import "fmt"

type PC struct {
	cpu       string
	gpu       string
	ramGB     int
	storageGB int
}

func (p *PC) GetSpecs() string {
	return fmt.Sprintf("Cpu: %v\nGpu: %v\nRam: %v GB\nStorage: %v GB", p.cpu, p.gpu, p.ramGB, p.storageGB)
}

type PCBuilder interface {
	SetCPU()
	SetGPU()
	SetStorageCapacity()
	SetRAMCapacity()
	GetPC() PC
}

type ExpensivePCBuilder struct {
	cpu       string
	gpu       string
	ramGB     int
	storageGB int
}

func (b *ExpensivePCBuilder) SetGPU() {
	b.gpu = "RTX 4090"
}

func (b *ExpensivePCBuilder) SetCPU() {
	b.gpu = "Core i9 14900KS"
}

func (b *ExpensivePCBuilder) SetStorageCapacity() {
	b.storageGB = 10000
}

func (b *ExpensivePCBuilder) SetRAMCapacity() {
	b.ramGB = 1000
}

func (b *ExpensivePCBuilder) GetPC() PC {
	return PC{
		cpu:       b.cpu,
		gpu:       b.gpu,
		ramGB:     b.ramGB,
		storageGB: b.storageGB,
	}
}

func NewExpensivePCBuilder() *ExpensivePCBuilder {
	return &ExpensivePCBuilder{}
}

type ReasonablePCBuilder struct {
	cpu       string
	gpu       string
	ramGB     int
	storageGB int
}

func (b *ReasonablePCBuilder) SetGPU() {
	b.gpu = "RTX 4070"
}

func (b *ReasonablePCBuilder) SetCPU() {
	b.gpu = "Core i7 14700K"
}

func (b *ReasonablePCBuilder) SetStorageCapacity() {
	b.storageGB = 5000
}

func (b *ReasonablePCBuilder) SetRAMCapacity() {
	b.ramGB = 64
}

func (b *ReasonablePCBuilder) GetPC() PC {
	return PC{
		cpu:       b.cpu,
		gpu:       b.gpu,
		ramGB:     b.ramGB,
		storageGB: b.storageGB,
	}

}

func NewReasonablePCBuilder() *ReasonablePCBuilder {
	return &ReasonablePCBuilder{}
}

func GetPCBuilder(pcType string) PCBuilder {
	if pcType == "expensive" {
		return NewExpensivePCBuilder()
	}

	if pcType == "reasonable" {
		return NewReasonablePCBuilder()
	}

	return nil
}

type Director struct {
	builder PCBuilder
}

func NewDirector(builder PCBuilder) *Director {
	return &Director{
		builder: builder,
	}
}

func (d *Director) SetBuilder(builder PCBuilder) {
	d.builder = builder
}

func (d *Director) BuildPC() PC {
	d.builder.SetCPU()
	d.builder.SetGPU()
	d.builder.SetStorageCapacity()
	d.builder.SetRAMCapacity()

	return d.builder.GetPC()
}

func main() {
	expensiveBuilder := GetPCBuilder("expensive")
	reasonableBuilder := GetPCBuilder("reasonable")

	director := NewDirector(expensiveBuilder)

	expensivePC := director.BuildPC()

	fmt.Println("Expensive PC")
	fmt.Println(expensivePC.GetSpecs(), "\n")

	director.SetBuilder(reasonableBuilder)

	reasonablePC := director.BuildPC()

	fmt.Println("Reasonable PC")
	fmt.Println(reasonablePC.GetSpecs())

}
