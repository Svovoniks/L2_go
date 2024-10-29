package main

import (
	"errors"
	"fmt"
)

type BasePhone interface {
	SetName(name string)
	SetOS(os string)
	GetName() string
	GetOS() string
}

type Phone struct {
	name string
	os   string
}

func (p *Phone) SetName(name string) {
	p.name = name
}

func (p *Phone) SetOS(os string) {
	p.os = os
}

func (p *Phone) GetName() string {
	return p.name
}

func (p *Phone) GetOS() string {
	return p.os
}

type Pixel struct {
	Phone
}

func NewPixel() BasePhone {
	return &Pixel{
		Phone: Phone{
			name: "pixel 6",
			os:   "android",
		},
	}
}

type IPhone struct {
	Phone
}

func NewIPhone() BasePhone {
	return &IPhone{
		Phone: Phone{
			name: "iphone 6",
			os:   "ios",
		},
	}
}

func getPhone(phoneType string) (BasePhone, error) {
	switch phoneType {
	case "pixel":
		return NewPixel(), nil
	case "iphone":
		return NewIPhone(), nil
	default:
		return nil, errors.New("unknown phone")
	}
}

func main() {
	pixel, err := getPhone("pixel")
	if err == nil {
		fmt.Printf("Got phone with name: %v, and os: %v\n", pixel.GetName(), pixel.GetOS())
	}
	iphone, err := getPhone("iphone")
	if err == nil {
		fmt.Printf("Got phone with name: %v, and os: %v\n", iphone.GetName(), iphone.GetOS())
	}
}
