package general

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type B struct {
	KKKKK string
	Bdsb  DSB
}

func (b *B) BeginRequest(ctx string) {
	fmt.Println("b BeginRequest")
}

type C struct {
	CCCCCCK string
	Cdsb    DSB `inject:"req_handle"`
}

func (c *C) BeginRequest(ctx string) {
	fmt.Println("c BeginRequest")
}

type A struct {
	C
	III  int
	B2   *B
	Adsb DSB
}

type DSB interface {
	HAHAH()
}

type DSBS struct {
	a string
}

func (d *DSBS) HAHAH() {
	fmt.Println(d.a)
}

type JIEKOU = DSB

func TestUtil(t *testing.T) {
	a := new(A)
	a.B2 = new(B)

	//II := SSS()
	// IIValue := reflect.ValueOf(II)
	// IIType := reflect.TypeOf(II)
	structFields(a, func(sf reflect.StructField, v reflect.Value) {
		vtype := v.Type()
		// if vtype.Kind() == reflect.Interface && IIType.AssignableTo(vtype) && v.CanSet() {
		// 	v.Set(IIValue)
		// }
		if vtype.Kind() == reflect.Ptr {
			fmt.Println(sf, v.Type())
			fmt.Println(v.MethodByName("BeginRequest"))
		}
	})
	// a.Adsb.HAHAH()
	// a.C.Cdsb.HAHAH()
	// a.B2.Bdsb.HAHAH()
}

func SSS() JIEKOU {
	dsbs := new(DSBS)
	dsbs.a = "sb"
	return dsbs
}

func TestStore(t *testing.T) {
	store := newStore()
	store.Set("key1", "value1")
	store.Set("key2", "value2")
	t.Log(store.Keys())
	t.Log(store.Values())
	m := store.ToMap()
	m["key1"] = "dsb"
	t.Log(store.ToMap())
}

type TestUser struct {
	gorm.Model
	UserName string `gorm:"size:32"`
	Password string `gorm:"size:32"`
	Age      int
	Status   int
}

func newReorder(sort, field string, args ...string) *Reorder {
	fields := []string{field}
	fields = append(fields, args...)
	orders := []string{}
	for index := 0; index < len(fields); index++ {
		orders = append(orders, sort)
	}
	return &Reorder{
		fields: fields,
		orders: orders,
	}
}

type parsePool struct {
}

func TestParsePoolFunc(t *testing.T) {
	f := func() *parsePool {
		return new(parsePool)
	}
	t.Log(parsePoolFunc(f))

	f2 := func() {}
	t.Log(parsePoolFunc(f2))

	f3 := 20
	t.Log(parsePoolFunc(f3))

	values := reflect.ValueOf(f).Call([]reflect.Value{})
	if len(values) == 0 {
		panic("BindService func return to empty")
	}
	t.Log(values[0])
}

func TestParsePoolFunc2(t *testing.T) {
	t.Log(strings.Split("asdasdasdsad", "?"))
}

type testEntity struct {
	Entity
	name string
}

func (entity *testEntity) Test(fuck int) {
	jsonData := struct {
		Value int
	}{
		Value: fuck,
	}
	entity.DomainEvent(entity.Test, jsonData)
}

type testEventInfra struct {
}

func (tei *testEventInfra) DomainEvent(producer, topic string, data []byte, runtime Runtime, header ...map[string]string) {
	fmt.Println(topic, data, header)
}

func TestEntity(t *testing.T) {
	NewApplication()
	globalApp.eventInfra = new(testEventInfra)
	tn := new(testEntity)
	newEntity(nil, tn)
	t.Log(tn.Identity())
	tn.Test(100)
}
