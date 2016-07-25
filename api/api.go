package api

import "net/http"
import "io/ioutil"
import "github.com/gorilla/mux"
import "github.com/urfave/negroni"

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gitlab.com/nerdalize/yak/yak"
	"strconv"
)

type Shop struct {
	CurrentDay int
	Herd       *yak.Herd
	Stock      *yak.Stock
}

func (s *Shop) AdvanceToDay(day int) {
	for ; day > s.CurrentDay; s.CurrentDay++ {
		s.Herd.Attend(s.Stock)
		s.Herd.DayPasses()
	}
}

func (s *Shop) Reset() {
	s.Herd = &yak.Herd{}
	s.Stock = &yak.Stock{}
	s.CurrentDay = 0
}

func NewShop() *Shop {
	s := Shop{}
	s.Reset()

	return &s
}

func (s *Shop) getDay(w http.ResponseWriter, r *http.Request) (int, error) {
	vars := mux.Vars(r)
	day, err := strconv.Atoi(vars["day"])
	if err != nil {
		return 0, err
	}

	if day > 100000 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Too large a day\n"))
		return 0, fmt.Errorf("Too large Day")
	}
	return day, nil
}

func (s *Shop) stateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Current Day: %d\n", s.CurrentDay)))
	return
}

func (s *Shop) loadHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	s.Reset()

	err = xml.Unmarshal(body, s.Herd)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusResetContent)

	w.Write([]byte(fmt.Sprintf("New data loaded (%d yaks)\n", len(s.Herd.Yaks))))

	w.Header().Set("Content-Type", "application/xml")

	return
}

func (s *Shop) orderHandler(w http.ResponseWriter, r *http.Request) {

	var day int
	var err error

	/**************************
	 *	proceed to current day
	 */
	if day, err = s.getDay(w, r); err != nil {
		return
	}
	s.AdvanceToDay(day)

	/**************************
	 *	read `order` from msg body
	 */
	var body []byte
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	order := &yak.Order{}
	err = json.Unmarshal(body, order)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot unmarshall order: " + err.Error()))
		return
	}

	/**************************
	 *	process `order`
	 */
	stock := s.Stock
	ship := &order.Goods

	itemsRequested := 0
	itemsUnavailable := 0

	if ship.Milk > 0 {
		itemsRequested++
		if err := stock.TakeMilk(ship.Milk); err != nil {
			itemsUnavailable++
			ship.Milk = 0
		}
	}
	if ship.Skins > 0 {
		itemsRequested++
		if err := stock.TakeSkins(ship.Skins); err != nil {
			itemsUnavailable++
			ship.Skins = 0
		}
	}

	w.Header().Set("Content-Type", "application/json")

	switch {
	case itemsRequested == 0:
		w.WriteHeader(http.StatusNotFound)
	case itemsRequested == itemsUnavailable:
		w.WriteHeader(http.StatusNotFound)
	case itemsUnavailable > 0:
		w.WriteHeader(http.StatusPartialContent)
	default:
		w.WriteHeader(http.StatusCreated)
	}

	ser, err := json.Marshal(order.Goods)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot marshall stock"))
		return
	}

	w.Write(ser)
}

func (s *Shop) herdHandler(w http.ResponseWriter, r *http.Request) {
	day, err := s.getDay(w, r)
	if err != nil {
		return
	}
	s.AdvanceToDay(day)

	ser, err := json.Marshal(s.Herd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot marshall herd"))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write(ser)
}

func (s *Shop) stockHandler(w http.ResponseWriter, r *http.Request) {
	day, err := s.getDay(w, r)

	if err != nil {
		return
	}
	s.AdvanceToDay(day)

	ser, err := json.Marshal(s.Stock)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot marshall stock"))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.Write(ser)
}

func (s *Shop) GetHandler() http.Handler {

	r := mux.NewRouter()
	r.HandleFunc("/yak-shop/load", s.loadHandler)
	r.HandleFunc("/yak-shop/herd/{day:[0-9]+}", s.herdHandler)
	r.HandleFunc("/yak-shop/stock/{day:[0-9]+}", s.stockHandler)
	r.HandleFunc("/yak-shop/order/{day:[0-9]+}", s.orderHandler)

	r.HandleFunc("/yak-shop/state", s.stateHandler)

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.UseHandler(r)
	return n
}
