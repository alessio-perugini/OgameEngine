package main

import (
	"errors"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
)

type TechParam struct {
	structure int64
	shield    int64
	attack    int64
	cargo     int64
}

type Slot struct {
	fleet             [14]uint32
	def               [8]uint32
	weap, shld, armor int
	name              [64]rune
	g, s, p           int
	id                int
}

type Unit struct {
	slot_id           rune
	obj_type          rune
	exploded          bool
	dummy             rune
	hull, hullmax     int64
	shield, shieldmax int64
}

type UnitPrice struct {
	m, k, d int64
}

var ResultBuffer [64 * 1024]rune
var DefenseInDebris int
var FleetInDebris = 30
var Rapidfire = 1

var N = 624
var M = 397

var FleetPrice = []UnitPrice{
	{2000, 2000, 0}, {6000, 6000, 0}, {3000, 1000, 0}, {6000, 4000, 0},
	{20000, 7000, 2000}, {45000, 15000, 0}, {10000, 20000, 10000}, {10000, 6000, 2000},
	{0, 1000, 0}, {50000, 25000, 15000}, {0, 2000, 500}, {60000, 50000, 15000},
	{5000000, 4000000, 1000000}, {30000, 40000, 15000}}

var DefensePrice = []UnitPrice{
	{2000, 0, 0}, {1500, 500, 0}, {6000, 2000, 0}, {20000, 15000, 2000},
	{2000, 6000, 0}, {50000, 50000, 30000}, {10000, 10000, 0}, {50000, 50000, 0}}

var fleetParam = [...]TechParam{ // ТТХ Флота.
	{4000, 10, 5, 5000},
	{12000, 25, 5, 25000},
	{4000, 10, 50, 50},
	{10000, 25, 150, 100},
	{27000, 50, 400, 800},
	{60000, 200, 1000, 1500},
	{30000, 100, 50, 7500},
	{16000, 10, 1, 20000},
	{1000, 0, 0, 0},
	{75000, 500, 1000, 500},
	{2000, 1, 1, 0},
	{110000, 500, 2000, 2000},
	{9000000, 50000, 200000, 1000000},
	{70000, 400, 700, 750},
}

var defenseParam = [...]TechParam{ // ТТХ Обороны.
	{2000, 20, 80, 0},
	{2000, 25, 100, 0},
	{8000, 100, 250, 0},
	{35000, 200, 1100, 0},
	{8000, 500, 150, 0},
	{100000, 300, 3000, 0},
	{20000, 2000, 1, 0},
	{100000, 10000, 1, 0},
}

func FileLoad(filename string) string {
	if filename == "" {
		return ""
	}

	file, err := os.OpenFile(filename, os.O_RDONLY, 0777)
	if err != nil {
		log.Println("error opening file", err.Error())
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("error opening file", err.Error())
	}

	return string(data)
}

func FileSave(filename string, data []byte) error {
	if filename == "" {
		return errors.New("invalid filename")
	}

	err := ioutil.WriteFile(filename, data, 0777) //TODO pericoloso assicurarsi che già esista!!!
	if err != nil {
		return errors.New("error on saving file")
	}

	return nil
}

func MyRand(min, max int) uint32 {
	return uint32(rand.Intn(max-min) + min) //TODO forse +1??
}

func SetDebrisOptions(did, fid int) {
	if did < 0 {
		did = 0
	}
	if fid < 0 {
		fid = 0
	}
	if did > 100 {
		did = 100
	}
	if fid > 100 {
		fid = 100
	}
	DefenseInDebris = did
	FleetInDebris = fid
}

func SetRapidfire(enable int) { Rapidfire = enable & 1 } //TODO check enable & 1

//TODO dangerous stuff here could explode all!!!!
func InitBattleAttackers(a []Slot, anum int, objs int) []Unit {
	u := make([]Unit, objs)
	ucnt := 0
	n := 0

	for i := 0; i < anum; i++ {
		for n = 0; n < 14; n++ {
			for obj := 0; uint32(obj) < a[i].fleet[n]; obj++ {
				u[ucnt].hull = u[ucnt].hullmax //TODO some shit with type
				u[ucnt].hullmax = int64(float64(fleetParam[n].structure) * 0.1 * (10 + float64(a[i].armor)) / 10)
				u[ucnt].obj_type = rune(100 + n) //TODO check this
				u[ucnt].slot_id = rune(i)        //aid //TODO check this
				ucnt++
			}
		}
	}

	return u
}

func InitBattleDefenders(d []Slot, dnum int, objs int) []Unit {
	u := make([]Unit, objs)
	ucnt := 0
	n := 0

	for i := 0; i < dnum; i++ {
		for n = 0; n < 14; n++ {
			for obj := 0; uint32(obj) < d[i].fleet[n]; obj++ {
				u[ucnt].hull = u[ucnt].hullmax //TODO some shit with type
				u[ucnt].hullmax = int64(float64(fleetParam[n].structure) * 0.1 * (10 + float64(d[i].armor)) / 10)
				u[ucnt].obj_type = rune(100 + n) //TODO check this
				u[ucnt].slot_id = rune(i)        //aid //TODO check this
				ucnt++
			}
		}
		for n = 0; n < 8; n++ {
			for obj := 0; uint32(obj) < d[i].def[n]; obj++ {
				u[ucnt].hull = u[ucnt].hullmax
				u[ucnt].hullmax = int64(float64(defenseParam[n].structure) * 0.1 * (10 + float64(d[i].armor)) / 10)
				u[ucnt].obj_type = rune(200 + n)
				u[ucnt].slot_id = rune(i) //did
				ucnt++
			}
		}
	}

	return u
}

func UniShoot(a *Unit, aweap int, b *Unit, absorbed, dm, dk *uint64) int64 {
	var prc float64
	var depleted float64
	var apower int64
	var adelta int64

	if a.obj_type < 200 {
		apower = fleetParam[a.obj_type-100].attack * (10 + int64(aweap)) / 10
	} else {
		return apower
	}

	if b.exploded {
		return apower
	}

	if b.shield == 0 {
		if apower >= b.hull {
			b.hull = 0
		} else {
			b.hull -= apower
		}
	} else {
		prc = float64(b.shieldmax) * 0.01
		depleted = math.Floor(float64(apower) / prc)
		if b.shield < int64(depleted*prc) {
			*absorbed += uint64(b.shield) //TODO check pointer
			adelta = apower - b.shield
			if adelta >= b.hull {
				b.hull = 0
			} else {
				b.hull -= adelta
			}
			b.shield = 0
		} else {
			b.shield -= int64(depleted * prc)
			*absorbed += uint64(apower) //TODO check pointer
		}
	}

	if b.hull <= int64(float64(b.hullmax)*0.7) && b.shield == 0 { // �������� � �������� ����.
		if MyRand(0, 99) >= uint32((b.hull*100)/b.hullmax) || b.hull == 0 {
			if b.obj_type >= 200 {
				*dm += uint64(math.Ceil(float64(DefensePrice[b.obj_type-200].m) * float64(DefenseInDebris/100.0)))
				*dk += uint64(math.Ceil(float64(DefensePrice[b.obj_type-200].k) * float64(DefenseInDebris/100.0)))
			} else {
				*dm += uint64(math.Ceil(float64(DefensePrice[b.obj_type-100].m) * float64(FleetInDebris/100.0)))
				*dk += uint64(math.Ceil(float64(DefensePrice[b.obj_type-100].k) * float64(FleetInDebris/100.0)))
			}
			b.exploded = true
		}
	}

	return apower
}

//TODO check if really wipes the exploded not really sure
func WipeExploded(slot *[]Unit, amount int) int {
	var src = *slot
	exploded := 0
	/*var tmp = src
	p :=0

	for i:=0; i<amount; i++ {
		if !src[i].exploded {
			tmp[p] = src[i]
			p++
		} else {
			exploded++
		}
	}

	*slot = tmp*/

	for i := 0; i < amount; i++ {
		if src[i].exploded {
			src[i] = src[len(src)-1]
			src = src[:len(src)-1]
			exploded++
		}
	}

	return exploded
}

func CheckFastDraw(aunits []Unit, aobjs int, dunits []Unit, dobjs int) int {
	for i := 0; i < aobjs; i++ {
		if aunits[i].hull != aunits[i].hullmax {
			return 0
		}
	}

	for i := 0; i < dobjs; i++ {
		if dunits[i].hull != dunits[i].hullmax {
			return 0
		}
	}

	return 1
}

func main() {

}
