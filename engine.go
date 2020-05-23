package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	name              []rune //64
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

var ResultBuffer = make([]rune, 64*1024) //64 * 1024
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

func MyRand(a, b int) uint32 {
	if b-a <= 0 {
		return uint32(rand.Intn(1-0) + 0)
	}
	return uint32(rand.Intn(b-a) + a)
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

func UnitShoot(a Unit, aweap int, b *Unit, absorbed, dm, dk *uint64) int64 {
	var prc, depleted float64
	var apower, adelta int64

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
				*dm += uint64(math.Ceil(float64(FleetPrice[b.obj_type-100].m) * float64(FleetInDebris/100.0)))
				*dk += uint64(math.Ceil(float64(FleetPrice[b.obj_type-100].k) * float64(FleetInDebris/100.0)))
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
	var tmp = src
	p := 0

	for i := 0; i < amount; i++ {
		if !src[i].exploded {
			tmp[p] = src[i]
			p++
		} else {
			exploded++
		}
	}

	*slot = tmp
	/*
		for i := 0; i < amount; i++ {
			if src[i].exploded {
				src[i] = src[len(src)-1]
				src = src[:len(src)-1]
				exploded++
			}
		}
	*/
	return exploded
}

func CheckFastDraw(aunits []Unit, aobjs int, dunits []Unit, dobjs int) bool {
	for i := 0; i < aobjs; i++ {
		if aunits[i].hull != aunits[i].hullmax {
			return false
		}
	}

	for i := 0; i < dobjs; i++ {
		if dunits[i].hull != dunits[i].hullmax {
			return false
		}
	}

	return true
}

//TODO check rune on slot forse da int a rune boh
//TODo check ptr pointer
func GenSlot(ptr *strings.Builder, units []Unit, slot, objnum int, a, d []Slot, attacker, techs bool) /*strings.Builder*/ {
	var s []Slot
	if attacker {
		s = a
	} else {
		s = d
	}
	var coll Slot
	var u *Unit
	var sum uint32

	for i := 0; i < objnum; i++ {
		u = &units[i]
		if u.slot_id == rune(slot) {
			if u.obj_type < 200 {
				coll.fleet[u.obj_type-100]++
				sum++
			} else {
				coll.def[u.obj_type-200]++
				sum++
			}
		}
	}

	if techs {
		if attacker { //TODO check %d
			ptr.WriteString(fmt.Sprintf("i:%d;a:22:{", slot))
		} else {
			ptr.WriteString(fmt.Sprintf("i:%d;a:30:{", slot))
		}
	} else {
		if attacker { //TODO check %d
			ptr.WriteString(fmt.Sprintf("i:%d;a:19:{", slot))
		} else {
			ptr.WriteString(fmt.Sprintf("i:%d;a:27:{", slot))
		}
	}
	ptr.WriteString(fmt.Sprintf("s:4:\"name\";s:%d:\"%s\";", len(s[slot].name), string(s[slot].name)))
	ptr.WriteString(fmt.Sprintf("s:2:\"id\";i:%d;", s[slot].id))
	ptr.WriteString(fmt.Sprintf("s:1:\"g\";i:%d;", s[slot].g))
	ptr.WriteString(fmt.Sprintf("s:1:\"s\";i:%d;", s[slot].s))
	ptr.WriteString(fmt.Sprintf("s:1:\"p\";i:%d;", s[slot].p))

	if techs {
		ptr.WriteString(fmt.Sprintf("s:4:\"weap\";i:%d;", s[slot].weap))
		ptr.WriteString(fmt.Sprintf("s:4:\"shld\";i:%d;", s[slot].shld))
		ptr.WriteString(fmt.Sprintf("s:4:\"armr\";i:%d;", s[slot].armor))
	}

	for n := 0; n < 14; n++ {
		ptr.WriteString(fmt.Sprintf("i:%d;i:%d;", 202+n, coll.fleet[n]))
	}

	if !attacker {
		for n := 0; n < 8; n++ {
			ptr.WriteString(fmt.Sprintf("i:%d;i:%d;", 401+n, coll.def[n]))
		}
	}

	ptr.WriteString(fmt.Sprintf("}"))
	//return *ptr //TODO CHECKTHIS
}

func RapidFire(atyp, dtyp int) bool {
	rapidfire := false

	if atyp > 400 {
		return false
	}

	if atyp == 214 && (dtyp == 210 || dtyp == 212) && MyRand(1, 10000) > 8 {
		rapidfire = true
	} else if atyp != 210 && (dtyp == 210 || dtyp == 212) && MyRand(1, 100) > 20 {
		rapidfire = true
	} else if atyp == 205 && dtyp == 202 && MyRand(1, 100) > 33 {
		rapidfire = true
	} else if atyp == 206 && dtyp == 204 && MyRand(1, 1000) > 166 {
		rapidfire = true
	} else if atyp == 206 && dtyp == 401 && MyRand(1, 100) > 10 {
		rapidfire = true
	} else if atyp == 211 && (dtyp == 401 || dtyp == 402) && MyRand(1, 100) > 20 {
		rapidfire = true
	} else if atyp == 211 && (dtyp == 403 || dtyp == 405) && MyRand(1, 100) > 10 {
		rapidfire = true
	} else if atyp == 213 && dtyp == 215 && MyRand(1, 100) > 50 {
		rapidfire = true
	} else if atyp == 213 && dtyp == 402 && MyRand(1, 100) > 10 {
		rapidfire = true
	} else if atyp == 215 && (dtyp == 202 || dtyp == 203) && MyRand(1, 100) > 20 {
		rapidfire = true
	} else if atyp == 215 && (dtyp == 205 || dtyp == 206) && MyRand(1, 100) > 25 {
		rapidfire = true
	} else if atyp == 215 && dtyp == 207 && MyRand(1, 1000) > 143 {
		rapidfire = true
	} else if atyp == 214 && (dtyp == 202 || dtyp == 203 || dtyp == 208 || dtyp == 209) && MyRand(1, 1000) > 4 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 204 && MyRand(1, 1000) > 5 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 205 && MyRand(1, 1000) > 10 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 206 && MyRand(1, 1000) > 30 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 207 && MyRand(1, 1000) > 33 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 211 && MyRand(1, 1000) > 40 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 213 && MyRand(1, 1000) > 200 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 215 && MyRand(1, 1000) > 66 {
		rapidfire = true
	} else if atyp == 214 && (dtyp == 401 || dtyp == 402) && MyRand(1, 1000) > 5 {
		rapidfire = true
	} else if atyp == 214 && (dtyp == 403 || dtyp == 405) && MyRand(1, 1000) > 10 {
		rapidfire = true
	} else if atyp == 214 && dtyp == 404 && MyRand(1, 1000) > 20 {
		rapidfire = true
	}

	return rapidfire
}

func DoBattle(a []Slot, anum int, d []Slot, dnum int) int {
	var aobjs, dobjs, idx, rounds int64
	var apower, atyp, dtyp int64
	var fastdraw, rapidfire bool
	var aunits, dunits []Unit
	var unit Unit
	var res string //TODO capire a cosa serve questa var: round_patch string
	var dm, dk uint64
	var ptr strings.Builder

	shoots := make([]uint64, 2)
	spower := make([]uint64, 2)
	absorbed := make([]uint64, 2)

	for i := 0; i < anum; i++ {
		for n := 0; n < 14; n++ {
			aobjs += int64(a[i].fleet[n])
		}
	}
	for i := 0; i < dnum; i++ {
		for n := 0; n < 14; n++ {
			dobjs += int64(d[i].fleet[n])
		}
		if i == 0 {
			for n := 0; n < 8; n++ {
				dobjs += int64(d[i].def[n])
			}
		}
	}

	aunits = InitBattleAttackers(a, anum, int(aobjs))
	if len(aunits) == 0 {
		return 0
	}
	dunits = InitBattleDefenders(d, dnum, int(dobjs))
	if len(dunits) == 0 {
		return 0
	}

	ptr.WriteString("a:5:{")
	ptr.WriteString("s:6:\"before\";a:2:{")
	ptr.WriteString(fmt.Sprintf("s:9:\"attackers\";a:%d:{", anum))
	for slot := 0; slot < anum; slot++ {
		/*ptr =*/ GenSlot(&ptr, aunits, slot, int(aobjs), a, d, true, true) //TODO uhm pointer stuff not sure
	}
	ptr.WriteString("}")
	ptr.WriteString(fmt.Sprintf("s:9:\"defenders\";a:%d:{", dnum))

	for slot := 0; slot < dnum; slot++ {
		/*ptr =*/ GenSlot(&ptr, dunits, slot, int(dobjs), a, d, false, true)
	}
	ptr.WriteString("}")
	ptr.WriteString("}")

	ptr.WriteString("s:6:\"rounds\";a:X:{")

	for rounds = 0; rounds < 6; rounds++ {
		if aobjs == 0 || dobjs == 0 {
			break
		}

		shoots[0] = 0
		shoots[1] = 0
		spower[0] = 0
		spower[1] = 0
		absorbed[0] = 0
		absorbed[1] = 0

		// �������� ����.
		for i := 0; int64(i) < aobjs; i++ {
			if aunits[i].exploded {
				aunits[i].shield = 0
				aunits[i].shieldmax = 0
			} else {
				aunits[i].shieldmax = fleetParam[aunits[i].obj_type-100].shield * (10 + int64(a[aunits[i].slot_id].shld)) / 10
				aunits[i].shield = aunits[i].shieldmax
			}
		}
		for i := 0; int64(i) < dobjs; i++ {
			if dunits[i].exploded {
				dunits[i].shield = 0
				dunits[i].shieldmax = 0
			} else {
				if dunits[i].obj_type >= 200 {
					dunits[i].shieldmax = defenseParam[dunits[i].obj_type-200].shield * (10 + int64(d[dunits[i].slot_id].shld)) / 10
					dunits[i].shield = dunits[i].shieldmax
				} else {
					dunits[i].shieldmax = fleetParam[dunits[i].obj_type-100].shield * (10 + int64(d[dunits[i].slot_id].shld)) / 10
					dunits[i].shield = dunits[i].shieldmax
				}
			}
		}

		for slot := 0; slot < anum; slot++ {
			for i := 0; int64(i) < aobjs; i++ {
				rapidfire = true
				unit = aunits[i]
				if int(unit.slot_id) == slot { //TODO check this
					for rapidfire {
						idx = int64(MyRand(0, int(dobjs-1)))
						apower = UnitShoot(unit, a[slot].weap, &dunits[idx], &absorbed[1], &dm, &dk)
						shoots[0]++
						spower[0] += uint64(apower) //TODO check this

						// ��������� ID � ������� ������, ����� ���� ��������.
						atyp = int64(unit.obj_type) //TODO check this
						if atyp < 200 {
							atyp += 102
						} else {
							atyp += 201
						}
						dtyp = int64(dunits[idx].obj_type) //TODO check this
						if dtyp < 200 {
							dtyp += 102
						} else {
							dtyp += 201
						}
						rapidfire = RapidFire(int(atyp), int(dtyp))

						if Rapidfire == 0 {
							rapidfire = false
						}
					}
				}
			}
		}
		for slot := 0; slot < dnum; slot++ {
			for i := 0; int64(i) < dobjs; i++ {
				rapidfire = true
				unit = dunits[i]
				if int(unit.slot_id) == slot { //TODO check this
					// �������.
					for rapidfire {
						idx = int64(MyRand(0, int(aobjs-1)))
						apower = UnitShoot(unit, d[slot].weap, &aunits[idx], &absorbed[0], &dm, &dk)
						shoots[1]++
						spower[1] += uint64(apower)

						atyp = int64(unit.obj_type) //TODO check this
						if atyp < 200 {
							atyp += 102
						} else {
							atyp += 201
						}
						dtyp = int64(aunits[idx].obj_type) //TODO check this
						if dtyp < 200 {
							dtyp += 102
						} else {
							dtyp += 201
						}
						rapidfire = RapidFire(int(atyp), int(dtyp))

						if Rapidfire == 0 {
							rapidfire = false
						}
					}
				}
			}
		}

		fastdraw = CheckFastDraw(aunits, int(aobjs), dunits, int(dobjs))

		aobjs -= int64(WipeExploded(&aunits, int(aobjs)))
		dobjs -= int64(WipeExploded(&dunits, int(dobjs)))

		// Round.
		ptr.WriteString(fmt.Sprintf("i:%d;a:8:", rounds))
		ptr.WriteString(fmt.Sprintf("{s:6:\"ashoot\";d:%d;", int64(shoots[0])))
		ptr.WriteString(fmt.Sprintf("s:6:\"apower\";d:%d;", int64(spower[0])))
		ptr.WriteString(fmt.Sprintf("s:7:\"dabsorb\";d:%d;", int64(absorbed[1])))
		ptr.WriteString(fmt.Sprintf("s:6:\"dshoot\";d:%d;", int64(shoots[1])))
		ptr.WriteString(fmt.Sprintf("s:6:\"dpower\";d:%d;", int64(spower[1])))
		ptr.WriteString(fmt.Sprintf("s:7:\"aabsorb\";d:%d;", int64(absorbed[0])))
		ptr.WriteString(fmt.Sprintf("s:9:\"attackers\";a:%d:{", anum))
		for slot := 0; slot < anum; slot++ {
			/*ptr = */ GenSlot(&ptr, aunits, slot, int(aobjs), a, d, true, false)
		}
		ptr.WriteString("}")
		ptr.WriteString(fmt.Sprintf("s:9:\"defenders\";a:%d:{", dnum))
		for slot := 0; slot < dnum; slot++ {
			/*ptr = */ GenSlot(&ptr, dunits, slot, int(dobjs), a, d, false, false)
		}
		ptr.WriteString("}")
		ptr.WriteString("}")

		if fastdraw {
			rounds++
			break
		}
	}

	if aobjs > 0 && dobjs == 0 {
		res = "awon"
	} else if dobjs > 0 && aobjs == 0 {
		res = "dwon"
	} else {
		res = "draw"
	}

	ptr.WriteString(fmt.Sprintf("}s:6:\"result\";s:4:\"%s\";", res))
	ptr.WriteString(fmt.Sprintf("s:2:\"dm\";d:%d;", int64(dm)))
	ptr.WriteString(fmt.Sprintf("s:2:\"dk\";d:%d;}", int64(dk)))
	prepareFinal := ptr.String()
	ResultBuffer = []rune(strings.Replace(prepareFinal, "X", fmt.Sprintf("%d", rounds), 1))

	return 1
}

func stringToInt(input string, base, bitsize int) int64 {
	v, _ := strconv.ParseInt(input, base, bitsize)
	return v
}

func StartBattle(text string, battle_id int64) {
	var rf, fid, did, res, anum, dnum, firstXParam int
	var a, d []Slot
	var buf strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		lineSplit := strings.Split(line, " = ")
		name := lineSplit[0]
		data := lineSplit[1]

		switch name {
		case "Rapidfire":
			rf = int(stringToInt(data, 10, 64))
		case "FID":
			fid = int(stringToInt(data, 10, 64))
		case "DID":
			did = int(stringToInt(data, 10, 64))
		case "Attackers":
			anum = int(stringToInt(data, 10, 64))
			a = make([]Slot, anum)
		case "Defenders":
			dnum = int(stringToInt(data, 10, 64))
			d = make([]Slot, dnum)
		}

		if firstXParam >= 5 {
			if strings.Contains(name, "Attacker") {
				for i := 0; i < anum; i++ {
					buf.Reset()
					buf.WriteString(fmt.Sprintf("Attacker%d", i))
					index := strings.LastIndex(data, ">")
					a[i].name = []rune(data[2:index]) //get name in the <>
					data = data[index+2:]             //delete name
					fmt.Fscanf(strings.NewReader(data), "%d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
						&a[i].id,
						&a[i].g, &a[i].s, &a[i].p,
						&a[i].weap, &a[i].shld, &a[i].armor,
						&a[i].fleet[0],
						&a[i].fleet[1],
						&a[i].fleet[2],
						&a[i].fleet[3],
						&a[i].fleet[4],
						&a[i].fleet[5],
						&a[i].fleet[6],
						&a[i].fleet[7],
						&a[i].fleet[8],
						&a[i].fleet[9],
						&a[i].fleet[10],
						&a[i].fleet[11],
						&a[i].fleet[12],
						&a[i].fleet[13],
					)
				}
			}
			if strings.Contains(name, "Defender") {
				for i := 0; i < dnum; i++ {
					buf.Reset()
					buf.WriteString(fmt.Sprintf("Attacker%d", i))
					index := strings.LastIndex(data, ">")
					d[i].name = []rune(data[2:index]) //get name in the <>
					data = data[index+2:]             //delete name
					fmt.Fscanf(strings.NewReader(data), "%d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
						&d[i].id,
						&d[i].g, &d[i].s, &d[i].p,
						&d[i].weap, &d[i].shld, &d[i].armor,
						&d[i].fleet[0],  // MT
						&d[i].fleet[1],  // BT
						&d[i].fleet[2],  // LF
						&d[i].fleet[3],  // HF
						&d[i].fleet[4],  // CR
						&d[i].fleet[5],  // LINK
						&d[i].fleet[6],  // COLON
						&d[i].fleet[7],  // REC
						&d[i].fleet[8],  // SPY
						&d[i].fleet[9],  // BOMB
						&d[i].fleet[10], // SS
						&d[i].fleet[11], // DEST
						&d[i].fleet[12], // DS
						&d[i].fleet[13], // BC
						&d[i].def[0],    // RT
						&d[i].def[1],    // LL
						&d[i].def[2],    // HL
						&d[i].def[3],    // GS
						&d[i].def[4],    // IC
						&d[i].def[5],    // PL
						&d[i].def[6],    // SDOM
						&d[i].def[7],    // LDOM
					)
				}
			}
		}
		firstXParam++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	SetDebrisOptions(did, fid)
	SetRapidfire(rf)

	res = DoBattle(a, anum, d, dnum)

	if res > 0 {
		err := FileSave(fmt.Sprintf("battleresult/battle_%d.txt", battle_id), []byte(string(ResultBuffer)))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("too few args")
	}
	fmt.Println(os.Args)

	paramBattleId := os.Args[1]
	param := strings.Split(paramBattleId, "=")

	if len(param) < 2 {
		log.Fatal("incorrect battle id format")
	}

	battleId, err := strconv.ParseInt(param[1], 10, 64)
	if err != nil {
		log.Fatal(err.Error())
	}
	filename := fmt.Sprintf("battledata/battle_%d.txt", battleId)
	battleData := FileLoad(filename) //TODO check error in fileload
	StartBattle(battleData, battleId)
}
