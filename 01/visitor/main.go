package main

import (
	"fmt"
	"math"
)

type Visitor interface {
	VisitForSphere(*Sphere)
	VisitForCube(*Cube)
}

type Object interface {
	GetType() string
	Accept(Visitor)
}

type Sphere struct {
	Radius float64
}

func (s *Sphere) GetType() string {
	return "sphere"
}

func (s *Sphere) Accept(v Visitor) {
	v.VisitForSphere(s)
}

type Cube struct {
	Side float64
}

func (c *Cube) GetType() string {
	return "cube"
}

func (c *Cube) Accept(v Visitor) {
	v.VisitForCube(c)
}

type VolumeCalculator struct {
	Volume float64
}

func (v *VolumeCalculator) VisitForSphere(sp *Sphere) {
	v.Volume = 4.0 / 3.0 * math.Pi * math.Pow(sp.Radius, 3)
}

func (v *VolumeCalculator) VisitForCube(cb *Cube) {
	v.Volume = math.Pow(cb.Side, 3)
}

type AreaCalculator struct {
	Area float64
}

func (v *AreaCalculator) VisitForSphere(sp *Sphere) {
	v.Area = 4 * math.Pi * math.Pow(sp.Radius, 2)
}

func (v *AreaCalculator) VisitForCube(cb *Cube) {
	v.Area = math.Pow(cb.Side, 2) * 6
}

func main() {
	cube := &Cube{Side: 10}
	sphere := &Sphere{Radius: 11}

	areaCalc := &AreaCalculator{}

	cube.Accept(areaCalc)
	fmt.Printf("Cube area: %v\n", areaCalc.Area)

	sphere.Accept(areaCalc)
	fmt.Printf("Sphere area: %v\n\n", areaCalc.Area)

	volumeCalc := &VolumeCalculator{}

	cube.Accept(volumeCalc)
	fmt.Printf("Cube volume: %v\n", volumeCalc.Volume)

	sphere.Accept(volumeCalc)
	fmt.Printf("Sphere volume: %v\n", volumeCalc.Volume)
}
