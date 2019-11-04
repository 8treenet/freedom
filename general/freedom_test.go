package general

import (
	"fmt"
	"reflect"
	"testing"
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

func TestTotalPage(t *testing.T) {
	var (
		result   int
		size     = 90
		pageSize = 20
	)

	if size%pageSize == 0 {
		result = size / pageSize
	} else {
		result = size/pageSize + 1
	}
	fmt.Println(result)
	return
}
