package yak

import "encoding/xml"

import (
	"strconv"
	"fmt"
	"sync"

)


type YakDays int
var YakYear = 100 * YakDays(1); /* yakyear is 100 days */

type Yak struct {
	Name          string		`xml:"name,attr" json:"name"`
	Age           YakDays		`xml:"age,attr" json:"age"`    /* age in days */
	Sex           string		`xml:"sex,attr" json:"-"`
	AgeLastMilked YakDays 		`xml:"-" json:"-"`              /* age at which yak was last miled */
	AgeLastShaved YakDays	 	`json:"age-last-shaved"` 	/* age at which yak was last shaved */
}


func (a *YakDays) UnmarshalJSON(b []byte) (err error) {
	years := new(float64)

	if err = xml.Unmarshal(b, &years); err != nil {
		return err
	}
	*a = YakDays (100 * (*years) )
	return nil
}

func (m *YakDays) MarshalJSON() ([]byte, error){
	return []byte(fmt.Sprintf("%.2f", float64(*m)/100 )), nil
}


func (a *YakDays) UnmarshalXMLAttr(attr xml.Attr) (err error) {
	years, err := strconv.ParseFloat(attr.Value, 64)
	if err != nil {
		return err
	}

	*a = YakDays (years*100)
	return nil
}

func (a *YakDays) MarshalXMLAttr(attr xml.Attr) (err error) {
	return fmt.Errorf("not implemented")
}


/* shave if we can */
func (y *Yak) shave() {
	y.AgeLastShaved = y.Age
}


/* can shave */
func (y *Yak) CanShave() bool {
	skinDays := y.Age - y.AgeLastShaved

	canShave := float64(skinDays) >= 8 + float64(y.Age) * 0.01
	return canShave
}

/* milk the yak; return amound of milk in litres */
func (y *Yak) milk() Litres {

	y.AgeLastMilked = y.Age
	if y.Sex == "f" {
		return Litres(50 - float64(y.Age) * 0.03)
	}

	/* male doesn't give milk */
	return Litres(0)
}



type Herd struct {
	Yaks []*Yak		`xml:"labyak" json:"herd"`
}


func (h *Herd) Attend(s *Stock){

	for _, y := range h.Yaks {
		/* shave */
		if y.CanShave(){
			y.shave()
			s.Skins += 1
		}
		/* milk */
		m := y.milk()
		s.Milk += m

	}

}

func (h *Herd) DayPasses(){
	doCull := false;
	for _, y := range h.Yaks {
		y.Age += 1
		if y.Age >= 10 * YakYear {
			doCull = true
		}
	}
	/* cull the old yaks */
	if doCull {
		alive := make([]*Yak, 0)
		for _, y := range h.Yaks {
			if y.Age < 10 * YakYear {
				alive = append(alive, y)
			}
		}
		h.Yaks = alive
	}

}


type Litres float64
type Stock struct {
	Milk 	Litres `json:"milk"`
	Skins	int `json:"skins"`
	mutex 	sync.Mutex
}

/* thread safe mechanism to take milk. Returns an error if the product is not sufficiently available */
func (s *Stock) TakeMilk(q Litres) error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if q > s.Milk {
		return fmt.Errorf("not enough milk")
	}
	s.Milk -= q

	return nil
}


/* thread safe mechanism to take skins. Returns an error if the product is not sufficiently available */
func (s *Stock) TakeSkins(q int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if q > s.Skins {
		return fmt.Errorf("not enough skins")
	}

	s.Skins -= q
	return nil
}

func (m *Litres) MarshalJSON() ([]byte, error){
	return []byte(fmt.Sprintf("%.2f", float64(*m) )), nil
}

/* like stock, but serialized in different way (omits empty fields */
type OrderedGoods struct {
	Milk 	Litres `json:"milk,omitempty"`
	Skins	int `json:"skins,omitempty"`
}

type Order struct {
	Customer 	string		`json:"customer"`
	Goods		OrderedGoods	`json:"order"`
}
