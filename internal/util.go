package internal

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

// NewMap accepts a pointer to map, and create a new map to draw back the
// pointer.
func NewMap(dst interface{}) error {
	dstRef := reflect.ValueOf(dst)

	if dstRef.Kind() != reflect.Ptr {
		return errors.New("'dst' must be a pointer to map")
	}

	if dstRef.Elem().Kind() != reflect.Map {
		return errors.New("dst error")
	}
	result := reflect.MakeMap(reflect.TypeOf(dst).Elem())
	dstRef.Elem().Set(result)
	return nil
}

//InSlice .
func InSlice(haystack interface{}, searchItem interface{}) bool {
	arrRef := reflect.ValueOf(haystack)
	if arrRef.Kind() != reflect.Slice {
		return false
	}

	size := arrRef.Len()
	list := make([]interface{}, size)
	slice := arrRef.Slice(0, size)
	for index := 0; index < size; index++ {
		list[index] = slice.Index(index).Interface()
	}

	for index := 0; index < len(list); index++ {
		if list[index] == searchItem {
			return true
		}
	}
	return false
}

//NewSlice .
func NewSlice(dsc interface{}, len int) error {
	dstv := reflect.ValueOf(dsc)
	if dstv.Elem().Kind() != reflect.Slice {
		return errors.New("dsc error")
	}

	result := reflect.MakeSlice(reflect.TypeOf(dsc).Elem(), len, len)
	dstv.Elem().Set(result)
	return nil
}

//SliceDelete delete the specified subscript element of the array
func SliceDelete(arr interface{}, indexArr ...int) error {
	dstv := reflect.ValueOf(arr)
	if dstv.Elem().Kind() != reflect.Slice {
		return errors.New("dsc error")
	}
	result := reflect.MakeSlice(reflect.TypeOf(arr).Elem(), 0, dstv.Elem().Len()-len(indexArr))
	for index := 0; index < dstv.Elem().Len(); index++ {
		if InSlice(indexArr, index) {
			continue
		}
		result = reflect.Append(result, dstv.Elem().Index(index))
	}

	dstv.Elem().Set(result)
	return nil
}

// ConvertAssign .
func ConvertAssign(dest, src interface{}) error {
	// Common cases, without reflect.
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *rawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		}
	case time.Time:
		switch d := dest.(type) {
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *rawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		}
	}

	var sv reflect.Value

	switch d := dest.(type) {
	case *string:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = asString(src)
			return nil
		}
	case *[]byte:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes(nil, sv); ok {
			*d = b
			return nil
		}
	case *rawBytes:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes([]byte(*d)[:0], sv); ok {
			*d = rawBytes(b)
			return nil
		}
	case *bool:
		bv, err := driver.Bool.ConvertValue(src)
		if err == nil {
			*d = bv.(bool)
		}
		return err
	case *interface{}:
		*d = src
		return nil
	}

	if scanner, ok := dest.(Scanner); ok {
		return scanner.Scan(src)
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}

	if !sv.IsValid() {
		sv = reflect.ValueOf(src)
	}

	dv := reflect.Indirect(dpv)
	if sv.IsValid() && sv.Type().AssignableTo(dv.Type()) {
		switch b := src.(type) {
		case []byte:
			dv.Set(reflect.ValueOf(cloneBytes(b)))
		default:
			dv.Set(sv)
		}
		return nil
	}

	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}

	switch dv.Kind() {
	case reflect.Ptr:
		if src == nil {
			dv.Set(reflect.Zero(dv.Type()))
			return nil
		}
		dv.Set(reflect.New(dv.Type().Elem()))
		return ConvertAssign(dv.Interface(), src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
}

func asBytes(buf []byte, rv reflect.Value) (b []byte, ok bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10), true
	case reflect.Float32:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 32), true
	case reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64), true
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool()), true
	case reflect.String:
		s := rv.String()
		return append(buf, s...), true
	}
	return
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}

// Scanner .
type Scanner interface {
	Scan(src interface{}) error
}

// JMap .
type JMap struct {
	_map     map[interface{}]interface{}
	lock     sync.Mutex
	openLock bool
}

// NewJMap .
func NewJMap(openLock ...bool) *JMap {
	m := new(JMap)
	m._map = make(map[interface{}]interface{})
	if len(openLock) > 0 && openLock[0] {
		m.openLock = true
	}
	return m
}

// SetOrStore .
func (jm *JMap) SetOrStore(key interface{}, value interface{}) (v interface{}, set bool) {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	v, set = jm._map[key]
	if set {
		set = false
		return
	}
	set = true
	jm._map[key] = value
	v = value
	return
}

// Set .
func (jm *JMap) Set(key interface{}, value interface{}) {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	jm._map[key] = value
}

// Get .
func (jm *JMap) Get(key interface{}, value interface{}) error {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	v, ok := jm._map[key]
	if !ok {
		return errors.New("undefined")
	}

	return ConvertAssign(value, v)
}

// Exist .
func (jm *JMap) Exist(key interface{}) bool {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	_, ok := jm._map[key]
	return ok
}

// Interface .
func (jm *JMap) Interface(key interface{}) interface{} {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	v, ok := jm._map[key]
	if !ok {
		return nil
	}

	return v
}

// Remove .
func (jm *JMap) Remove(key interface{}) {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	delete(jm._map, key)
}

// DelAll .
func (jm *JMap) DelAll() {
	all := jm.AllKey()
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	for _, item := range all {
		jm.Remove(item)
	}
}

// AllKey .
func (jm *JMap) AllKey() []interface{} {
	if jm.openLock {
		defer jm.lock.Unlock()
		jm.lock.Lock()
	}
	list := make([]interface{}, 0, len(jm._map))
	for k := range jm._map {
		list = append(list, k)
	}
	return list
}

var errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error
type rawBytes []byte

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	}
	return fmt.Sprintf("%v", src)
}

func structFields(dest interface{}, call func(reflect.StructField, reflect.Value)) {
	destVal := indirect(reflect.ValueOf(dest))
	destType := indirectType(destVal.Type())
	ts := deepFields(destType)
	for index, field := range deepValues(destVal, false) {
		call(ts[index], field)
	}
}

// objectFields . if the member variable is nil, then create the object.
func objectFields(dest interface{}, call func(reflect.StructField, reflect.Value)) {
	destVal := indirect(reflect.ValueOf(dest))
	destType := indirectType(destVal.Type())
	ts := deepFields(destType)
	for index, field := range deepValues(destVal, true) {
		call(ts[index], field)
	}
}

func deepFields(reflectType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < reflectType.NumField(); i++ {
		v := reflectType.Field(i)
		vt := v.Type
		if vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
		}
		if (v.Anonymous && vt.Kind() == reflect.Struct) || vt.Kind() == reflect.Struct {
			fields = append(fields, deepFields(vt)...)
			fields = append(fields, v)
		} else {
			fields = append(fields, v)
		}
	}

	return fields
}

func deepValues(reflectValue reflect.Value, object bool) []reflect.Value {
	var fields []reflect.Value
	for index := 0; index < reflectValue.NumField(); index++ {
		val := reflectValue.Field(index)
		vt := val
		if vt.Kind() == reflect.Ptr {
			if vt.IsNil() && object {
				vt.Set(reflect.New(val.Type().Elem()))
			}
			if vt.IsNil() && !object {
				panic(fmt.Sprint(val.Type(), " is nil"))
			}
			vt = vt.Elem()

		}
		if vt.Kind() == reflect.Struct {
			fields = append(fields, deepValues(vt, object)...)
			fields = append(fields, val)
		} else {
			fields = append(fields, val)
		}
	}

	return fields
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr || reflectValue.Kind() == reflect.Interface {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

// allFields
func allFields(dest interface{}, call func(reflect.Value)) {
	destVal := indirect(reflect.ValueOf(dest))
	destType := destVal.Type()
	if destType.Kind() != reflect.Struct && destType.Kind() != reflect.Interface {
		return
	}

	for index := 0; index < destVal.NumField(); index++ {
		if destType.Field(index).Anonymous {
			allFields(destVal.Field(index).Addr().Interface(), call)
			continue
		}
		val := destVal.Field(index)
		call(val)
	}
}

// allFieldsFromValue
func allFieldsFromValue(val reflect.Value, call func(reflect.Value)) {
	destVal := indirect(val)
	destType := destVal.Type()
	if destType.Kind() != reflect.Struct && destType.Kind() != reflect.Interface {
		return
	}
	for index := 0; index < destVal.NumField(); index++ {
		if destType.Field(index).Anonymous {
			allFieldsFromValue(destVal.Field(index).Addr(), call)
			continue
		}
		val := destVal.Field(index)
		call(val)
	}
}

func parsePoolFunc(f interface{}) (outType reflect.Type, e error) {
	ftype := reflect.TypeOf(f)
	if ftype.Kind() != reflect.Func {
		e = errors.New("It's not a func")
		return
	}
	if ftype.NumOut() != 1 {
		e = errors.New("Return must be an object pointer")
		return
	}
	outType = ftype.Out(0)
	if outType.Kind() != reflect.Ptr {
		e = errors.New("Return must be an object pointer")
		return
	}
	return
}

func fetchValue(dest, src interface{}) bool {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return false
	}
	value = value.Elem()
	srvValue := reflect.ValueOf(src)
	if value.Type() == srvValue.Type() {
		value.Set(srvValue)
		return true
	}
	return false
}

func parseCallServiceFunc(f interface{}) (inType reflect.Type, e error) {
	ftype := reflect.TypeOf(f)
	if ftype.Kind() != reflect.Func {
		e = errors.New("It's not a func")
		return
	}
	if ftype.NumIn() != 1 {
		e = errors.New("The pointer parameter must be a service object")
		return
	}
	inType = ftype.In(0)
	if inType.Kind() != reflect.Ptr {
		e = errors.New("The pointer parameter must be a service object")
		return
	}
	return
}
