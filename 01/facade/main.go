package main

import "fmt"

type Output interface {
	SendData(str string)
}

type Input interface {
	ReceveData() string
}

type Keyboard struct {
	output Output
}

func (k *Keyboard) SetOutput(output Output) {
	k.output = output
	fmt.Println("keyboard connected")
}

func (k *Keyboard) Shutdown() {
	fmt.Println("keyboard shutdown")
}

type Mouse struct {
	output Output
}

func (m *Mouse) SetOutput(output Output) {
	m.output = output
	fmt.Println("mouse connected")
}

func (k *Mouse) Shutdown() {
	fmt.Println("mouse shutdown")
}

type Monitor struct {
	input Input
}

func (m *Monitor) SetInput(input Input) {
	m.input = input
	fmt.Println("monitor connected")
}

func (k *Monitor) Shutdown() {
	fmt.Println("monitor shutdown")
}

type Speaker struct {
	input Input
}

func (s *Speaker) SetInput(input Input) {
	s.input = input
	fmt.Println("speaker connected")
}

func (k *Speaker) Shutdown() {
	fmt.Println("speaker shutdown")
}

type Processor struct {
}

func (p *Processor) SendData(str string) {
	fmt.Println(str)
}

func (p *Processor) ReceveData() string {
	return "data"
}

func (k *Processor) Shutdown() {
	fmt.Println("processor shutdown")
}

type ComputerFacade struct {
	processor *Processor
	keyboard  *Keyboard
	mouse     *Mouse
	monitor   *Monitor
	speakers  []*Speaker
}

func (c *ComputerFacade) Boot() {
	c.keyboard.SetOutput(c.processor)
	c.mouse.SetOutput(c.processor)
	c.monitor.SetInput(c.processor)
	for _, sp := range c.speakers {
		sp.SetInput(c.processor)
	}
}

func (c *ComputerFacade) Shutdown() {
	c.keyboard.Shutdown()
	c.mouse.Shutdown()
	c.monitor.Shutdown()
	for _, sp := range c.speakers {
		sp.Shutdown()
	}
	c.processor.Shutdown()
}

func NewComputerFacade(processor *Processor, keyboard *Keyboard, mouse *Mouse, monitor *Monitor, speakers []*Speaker) *ComputerFacade {
	return &ComputerFacade{
		processor: processor,
		keyboard:  keyboard,
		mouse:     mouse,
		monitor:   monitor,
		speakers:  speakers,
	}
}

func main() {
	comuter := NewComputerFacade(new(Processor), new(Keyboard), new(Mouse), new(Monitor), []*Speaker{new(Speaker), new(Speaker)})

	comuter.Boot()
	comuter.Shutdown()
}
