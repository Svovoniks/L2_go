package main

import "fmt"

type EditorState interface {
	OpenDocument()
	ReadDocument()
	AddText(text string)
	SaveDocument()
	CloseDocument()
}

type Editor struct {
	noDocument    EditorState
	documentOpen  EditorState
	documentSaved EditorState

	currentState EditorState

	documentContents string
}

func NewEditor() *Editor {
	editor := &Editor{}

	noDoc := &NoDocumentState{editor: editor}
	openDoc := &OpenDocumentState{editor: editor}
	savedDoc := &SavedDocumentState{editor: editor}

	editor.SetState(noDoc)
	editor.noDocument = noDoc
	editor.documentOpen = openDoc
	editor.documentSaved = savedDoc

	return editor
}

func (e *Editor) SetState(state EditorState) {
	e.currentState = state
}

func (e *Editor) OpenDocument() {
	e.currentState.OpenDocument()
}

func (e *Editor) ReadDocument() {
	e.currentState.ReadDocument()
}

func (e *Editor) AddText(text string) {
	e.currentState.AddText(text)
}

func (e *Editor) SaveDocument() {
	e.currentState.SaveDocument()
}

func (e *Editor) CloseDocument() {
	e.currentState.CloseDocument()
}

type NoDocumentState struct {
	editor *Editor
}

func (s *NoDocumentState) OpenDocument() {
	s.editor.SetState(s.editor.documentOpen)
	fmt.Println("Document opened")
}

func (s *NoDocumentState) ReadDocument() {
	fmt.Println("Can't read, no documet is open")
}

func (s *NoDocumentState) AddText(text string) {
	fmt.Println("Can't add, no documet is open")
}

func (s *NoDocumentState) SaveDocument() {
	fmt.Println("Can't save, no documet is open")
}

func (s *NoDocumentState) CloseDocument() {
	fmt.Println("Can't close, no documet is open")
}

type OpenDocumentState struct {
	editor *Editor
}

func (s *OpenDocumentState) OpenDocument() {
	fmt.Println("Can't open new document, this one is not saved")
}

func (s *OpenDocumentState) ReadDocument() {
	fmt.Println(s.editor.documentContents)
}

func (s *OpenDocumentState) AddText(text string) {
	s.editor.documentContents += text
	fmt.Println("Added text")
}

func (s *OpenDocumentState) SaveDocument() {
	s.editor.SetState(s.editor.documentSaved)
	fmt.Println("Document saved")
}

func (s *OpenDocumentState) CloseDocument() {
	s.editor.SetState(s.editor.documentSaved)
	fmt.Println("Can't close the document it is not saved")
}

type SavedDocumentState struct {
	editor *Editor
}

func (s *SavedDocumentState) OpenDocument() {
	s.editor.documentContents = ""
	s.editor.SetState(s.editor.documentOpen)
	fmt.Println("New document opened")
}

func (s *SavedDocumentState) ReadDocument() {
	fmt.Println(s.editor.documentContents)
}

func (s *SavedDocumentState) AddText(text string) {
	s.editor.documentContents += text
	s.editor.SetState(s.editor.documentOpen)
	fmt.Println("Added text")
}

func (s *SavedDocumentState) SaveDocument() {
	fmt.Println("It's already saved, but sure why not")
}

func (s *SavedDocumentState) CloseDocument() {
	s.editor.SetState(s.editor.noDocument)
	fmt.Println("Document closed")
}

func main() {
	editor := NewEditor()

	editor.OpenDocument()

	editor.AddText("hi")

    editor.OpenDocument()

    editor.SaveDocument()

    editor.ReadDocument()

	editor.CloseDocument()

}
