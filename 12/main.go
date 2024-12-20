package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type ErrorReport struct {
	ErrorString string `json:"error"`
}

type ResultReport struct {
	Result interface{} `json:"result"`
}

type User struct {
	Id         int           `json:"user_id"`
	Name       string        `json:"username"`
	EventStore *Store[Event] `json:"-"`
}

func NewUser(username string) *User {
	return &User{
		Id:         -1,
		Name:       username,
		EventStore: NewStore[Event](),
	}
}

func (u *User) toJson() ([]byte, error) {
	return json.Marshal(u)
}

type Event struct {
	Id        int       `json:"event_id"`
	Title     string    `json:"event_title"`
	EventTime time.Time `json:"event_time"`
}

func (e *Event) toJson() ([]byte, error) {
	return json.Marshal(e)
}

type Store[T interface{}] struct {
	firstFreeIdx int
	objMap       map[int]*T
	mutex        sync.RWMutex
}

func NewStore[T interface{}]() *Store[T] {
	return &Store[T]{
		firstFreeIdx: 0,
		objMap:       make(map[int]*T),
	}
}

func (s *Store[T]) add(obj *T) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	idx := s.firstFreeIdx
	s.objMap[idx] = obj

	s.firstFreeIdx++
	return idx
}

func (s *Store[T]) get(id int) (*T, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if val, ok := s.objMap[id]; ok {
		return val, nil
	}

	return nil, errors.New("No such obj")
}

func (s *Store[T]) iterate(apply func(*T)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, val := range s.objMap {
		apply(val)
	}
}

func (s *Store[T]) update(id int, newObj *T) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.objMap[id]; ok {
		s.objMap[id] = newObj
		return nil
	}

	return errors.New("No such obj")
}

func (s *Store[T]) delete(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.objMap[id]; ok {
		delete(s.objMap, id)
		return nil
	}

	return errors.New("No such obj")
}

// POST /create_event
func createEvent(userIdx int, event *Event, userStore *Store[User]) (int, error) {
	user, err := userStore.get(userIdx)
	if err != nil {
		return -1, err
	}

	idx := user.EventStore.add(event)
	event.Id = idx
	return idx, nil
}

// POST /update_event
func updateEvent(userIdx int, eventIdx int, newEvent *Event, userStore *Store[User]) error {
	user, err := userStore.get(userIdx)
	if err != nil {
		return err
	}

	err = user.EventStore.update(eventIdx, newEvent)
	if err != nil {
		return err
	}

	return nil
}

// POST /delete_event
func deleteEvent(userIdx int, eventIdx int, userStore *Store[User]) error {
	user, err := userStore.get(userIdx)
	if err != nil {
		return err
	}

	err = user.EventStore.delete(eventIdx)
	if err != nil {
		return err
	}

	return nil
}

func getEventsInTimeFrame(start time.Time, end time.Time, eventStore *Store[Event]) []*Event {
	var res []*Event

	eventStore.iterate(func(ev *Event) {
		if start.Before(ev.EventTime) && end.After(ev.EventTime) {
			res = append(res, ev)
		}
	})

	return res
}

// GET /events_for_day
func eventsForDay(userIdx int, date time.Time, userStore *Store[User]) ([]*Event, error) {
	if user, err := userStore.get(userIdx); err == nil {
		end := date.AddDate(0, 0, 1)

		return getEventsInTimeFrame(date, end, user.EventStore), nil
	}
	return nil, errors.New("No such user")
}

// GET /events_for_week
func eventsForWeek(userIdx int, date time.Time, userStore *Store[User]) ([]*Event, error) {
	if user, err := userStore.get(userIdx); err == nil {
		date = date.AddDate(0, 0, -((int(date.Weekday()) + 6) % 7))
		year, month, day := date.Date()

		start := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
		end := start.AddDate(0, 0, 7)

		return getEventsInTimeFrame(start, end, user.EventStore), nil
	}
	return nil, errors.New("No such user")
}

// GET /events_for_month
func eventsForMonth(userIdx int, date time.Time, userStore *Store[User]) ([]*Event, error) {
	if user, err := userStore.get(userIdx); err == nil {
		year, month, _ := date.Date()

		start := time.Date(year, month, 0, 0, 0, 0, 0, date.Location())
		end := start.AddDate(0, 1, 0)

		return getEventsInTimeFrame(start, end, user.EventStore), nil
	}
	return nil, errors.New("No such user")
}

func SendError(w http.ResponseWriter, err error, errorCode int) {
	w.Header().Set("Content-Type", "application/json")
	if json, errE := json.Marshal(ErrorReport{ErrorString: err.Error()}); errE == nil {
		w.WriteHeader(errorCode)
		w.Write(json)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	return
}

func SendResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if json, errE := json.Marshal(ResultReport{Result: result}); errE == nil {
		w.Write(json)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	return
}
func parseEventId(body []byte) (int, error) {
	event := Event{Id: -1}
	err := json.Unmarshal(body, &event)
	if err != nil || event.Id == -1 {
		return -1, err
	}

	return event.Id, nil
}

func parseEvent(body []byte, needId bool) (*Event, error) {
	event := Event{Id: -1}
	err := json.Unmarshal(body, &event)
	if err != nil {
		return nil, err
	}

	if event.Title == "" {
		return nil, errors.New("Missing title")
	}

	if event.EventTime.IsZero() {
		return nil, errors.New("Missing time")
	}

	if needId && event.Id == -1 {
		return nil, errors.New("Missing id")
	}

	return &event, nil
}

func parseUsername(body []byte) (string, error) {
	user := User{}
	err := json.Unmarshal(body, &user)
	if err != nil {
		return "", err
	}
	if user.Name == "" {
		return "", errors.New("Invalid user id")
	}

	return user.Name, nil
}

func parseUserIdx(body []byte) (int, error) {
	user := User{Id: -1}
	err := json.Unmarshal(body, &user)
	if err != nil {
		return -1, err
	}
	if user.Id == -1 {
		return -1, errors.New("Invalid user id")
	}

	return user.Id, nil
}

func HandleCreateEvent(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	userIdx, err := parseUserIdx(body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	event, err := parseEvent(body, false)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	idx, err := createEvent(userIdx, event, userStore)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	SendResult(w, idx)
}

func HandleUpdateEvent(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	userIdx, err := parseUserIdx(body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	event, err := parseEvent(body, true)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	if err := updateEvent(userIdx, event.Id, event, userStore); err != nil {
		SendError(w, err, 500)
		return
	}

	SendResult(w, "Success")
}

func HandleDeleteEvent(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	userIdx, err := parseUserIdx(body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	eventId, err := parseEventId(body)

	if err != nil {
		SendError(w, err, 500)
		return
	}

	if err := deleteEvent(userIdx, eventId, userStore); err != nil {
		SendResult(w, "Success")
		return
	}

	SendError(w, errors.New("No such event"), 500)
}

func HandleEvnetsForTheDay(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	if !r.URL.Query().Has("user_id") {
		SendError(w, errors.New("Missing user id"), 400)
		return
	}

	if !r.URL.Query().Has("date") {
		SendError(w, errors.New("Missing date"), 400)
		return
	}

	userIdxStr := r.URL.Query().Get("user_id")
	userIdx, err := strconv.Atoi(userIdxStr)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	events, err := eventsForDay(userIdx, date, userStore)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	SendResult(w, events)
}

func HandleEvnetsWeek(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	if !r.URL.Query().Has("user_id") {
		SendError(w, errors.New("Missing user id"), 400)
		return
	}

	if !r.URL.Query().Has("date") {
		SendError(w, errors.New("Missing date"), 400)
		return
	}

	userIdxStr := r.URL.Query().Get("user_id")
	userIdx, err := strconv.Atoi(userIdxStr)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	events, err := eventsForWeek(userIdx, date, userStore)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	SendResult(w, events)
}

func HandleEvnetsMonth(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	if !r.URL.Query().Has("user_id") {
		SendError(w, errors.New("Missing user id"), 400)
		return
	}

	if !r.URL.Query().Has("date") {
		SendError(w, errors.New("Missing date"), 400)
		return
	}

	userIdxStr := r.URL.Query().Get("user_id")
	userIdx, err := strconv.Atoi(userIdxStr)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	events, err := eventsForMonth(userIdx, date, userStore)
	if err != nil {
		SendError(w, err, 500)
		return
	}

	SendResult(w, events)
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request, userStore *Store[User]) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	username, err := parseUsername(body)
	if err != nil {
		SendError(w, err, 400)
		return
	}

	user := NewUser(username)
	idx := userStore.add(user)

	SendResult(w, idx)
}

func StorageWrapper(fn func(http.ResponseWriter, *http.Request, *Store[User]), userStore *Store[User]) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, userStore)
	}
}

func LoggerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %v, Method: %v\n", r.URL, r.Method)
		handler.ServeHTTP(w, r)
	})
}

func runServer(port int) {
	userStore := NewStore[User]()

	createUserHandler := http.HandlerFunc(StorageWrapper(HandleCreateUser, userStore))
	createEventHandler := http.HandlerFunc(StorageWrapper(HandleCreateEvent, userStore))
	updateEventHandler := http.HandlerFunc(StorageWrapper(HandleUpdateEvent, userStore))
	deleteEventHandler := http.HandlerFunc(StorageWrapper(HandleDeleteEvent, userStore))
	dayEventsHandler := http.HandlerFunc(StorageWrapper(HandleEvnetsForTheDay, userStore))
	weekEventsHandler := http.HandlerFunc(StorageWrapper(HandleEvnetsWeek, userStore))
	monthEventsHandler := http.HandlerFunc(StorageWrapper(HandleEvnetsMonth, userStore))

	http.Handle("/create_user", LoggerMiddleware(createUserHandler))
	http.Handle("/create_event", LoggerMiddleware(createEventHandler))
	http.Handle("/update_event", LoggerMiddleware(updateEventHandler))
	http.Handle("/delete_event", LoggerMiddleware(deleteEventHandler))
	http.Handle("/events_for_day", LoggerMiddleware(dayEventsHandler))
	http.Handle("/events_for_week", LoggerMiddleware(weekEventsHandler))
	http.Handle("/events_for_month", LoggerMiddleware(monthEventsHandler))

	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	fmt.Println(err.Error())
}

type config struct {
	Port int
}

func getConfig() (*config, error) {
	file, err := os.ReadFile(".cfg")
	if err != nil {
		return nil, err
	}

	config := &config{Port: -1}
	if err := json.Unmarshal(file, config); err != nil || config.Port == -1 {
		return nil, err
	}

	return config, nil
}
func main() {
	config, err := getConfig()
	if err != nil {
		fmt.Println("Couldn't parse config")
		return
	}

	runServer(config.Port)
}
