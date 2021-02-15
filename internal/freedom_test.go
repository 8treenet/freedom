package internal

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
	I     int
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
	III int
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
	allFields(a, func(v reflect.Value) {
		t.Log(v.Type(), v)
	})
	allFieldsFromValue(reflect.ValueOf(a), func(v reflect.Value) {
		t.Log(v.Type(), v)
	})
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

func TestOther(t *testing.T) {
	ot := newOneShotPool()
	ot.addProvider(func() interface{} {
		return &TestUser{Age: 100}
	})
	ot.addProvider(func() interface{} {
		var list []*B
		for i := 0; i < 3; i++ {
			list = append(list, &B{I: i})
		}
		return list
	})

	ot.resolve()

	var tu *TestUser
	ot.get(&tu)
	t.Log(tu)

	var blist []*B
	ot.get(&blist)
	for i := 0; i < len(blist); i++ {
		t.Log(blist[i])
	}
}

func TestFetchValue(t *testing.T) {
	var tu *B
	src := new(TestUser)
	err := fetchValue(&tu, src)
	t.Log(tu, err)
}
