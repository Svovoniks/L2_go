package main

import "fmt"

type Event struct {
	dragEvent   bool
	clickEvent  bool
	hoverEvent  bool
	scrollEvent bool
	processed   bool
}

type Handler interface {
	execute(*Event)
	setNext(Handler)
}

type DragHandler struct {
	next Handler
}

func (h *DragHandler) execute(event *Event) {
	if event.dragEvent {
		event.processed = true
		fmt.Println("Processed DragEvent")
		return
	}

	h.next.execute(event)
}

func (h *DragHandler) setNext(next Handler) {
	h.next = next
}

type ClickHandler struct {
	next Handler
}

func (h *ClickHandler) execute(event *Event) {
	if event.clickEvent {
		event.processed = true
		fmt.Println("Processed ClickEvent")
		return
	}

	h.next.execute(event)
}

func (h *ClickHandler) setNext(next Handler) {
	h.next = next
}

type HoverHandler struct {
	next Handler
}

func (h *HoverHandler) execute(event *Event) {
	if event.hoverEvent {
		event.processed = true
		fmt.Println("Processed HoverEvent")
		return
	}

	h.next.execute(event)
}

func (h *HoverHandler) setNext(next Handler) {
	h.next = next
}

type ScrollHandler struct {
	next Handler
}

func (h *ScrollHandler) execute(event *Event) {
	if event.scrollEvent {
		event.processed = true
		fmt.Println("Processed ScrollEvent")
		return
	}

	h.next.execute(event)
}

func (h *ScrollHandler) setNext(next Handler) {
	h.next = next
}

type EndHandler struct {
}

func (h *EndHandler) execute(event *Event) {
	fmt.Println("Couldn't processes event")
}

func (h *EndHandler) setNext(next Handler) {
}

func main() {
	dragHandler := &DragHandler{}
	clickHandler := &ClickHandler{}
	hoverHandler := &HoverHandler{}
	scrollHandler := &ScrollHandler{}
	endHandler := &EndHandler{}

	dragHandler.setNext(clickHandler)
	clickHandler.setNext(hoverHandler)
	hoverHandler.setNext(scrollHandler)
	scrollHandler.setNext(endHandler)

	event := &Event{scrollEvent: true}

	dragHandler.execute(event)

	event = &Event{clickEvent: true}

	dragHandler.execute(event)

	event = &Event{hoverEvent: true}

	dragHandler.execute(event)

	event = &Event{dragEvent: true}

	dragHandler.execute(event)

	event = &Event{}

	dragHandler.execute(event)
}
