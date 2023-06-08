package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/dop251/goja"
)

const CircularNotation = "[Circular]"

func jsonValueToString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return strconv.Quote(v)
	case bool, int, int32, int64, uint, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	case json.RawMessage:
		return string(v)
	case nil:
		return "null"
	default:
		return ""
	}
}

func marshalCircular(v interface{}, visit func(uintptr) bool) string {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Struct:
		ret := "{"
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Type().Field(i).Name
			value := marshalCircular(rv.Field(i).Interface(), visit)
			if rv.Field(i).Kind() == reflect.Ptr && !rv.Field(i).IsNil() && rv.Field(i).Elem().Kind() == reflect.Ptr {
				ptr := rv.Field(i).Elem().Pointer()
				if visit(ptr) {
					value = CircularNotation
				}
			}
			ret += fmt.Sprintf(`"%s":%s`, field, value)
			if i < rv.NumField()-1 {
				ret += ","
			}
		}
		ret += "}"
		return ret
	case reflect.Ptr:
		ptr := rv.Pointer()
		if ptr == 0 {
			return "null"
		}
		if visit(ptr) {
			return CircularNotation
		}
		return marshalCircular(rv.Elem().Interface(), visit)
	case reflect.Slice, reflect.Array:
		length := rv.Len()
		ret := "["
		for i := 0; i < length; i++ {
			ret += marshalCircular(rv.Index(i).Interface(), visit)
			if i < length-1 {
				ret += ","
			}
		}
		ret += "]"
		return ret
	case reflect.Map:
		ret := "{"
		mapPtr := rv.Pointer()
		if mapPtr != 0 && visit(mapPtr) {
			return CircularNotation
		}
		keys := rv.MapKeys()
		for i, key := range keys {
			value := marshalCircular(rv.MapIndex(key).Interface(), visit)
			ret += fmt.Sprintf(`"%s":%s`, key.String(), value)
			if i < len(keys)-1 {
				ret += ","
			}
		}
		ret += "}"
		return ret
	default:
		return jsonValueToString(v)
	}
}

// JSONStringify is a function that marshals a JavaScript object to JSON.
// If an entry is a circular object, the function returns invalid JSON by replacing
// the object with [Circular] and returning a marshaling error.
// ⚠️ IMPORTANT: you MUST pass a pointer to the struct object as an argument.
// Failing to do so may cause improper memory management and unnecessary copying of the struct.
func JSONStringify(v interface{}) ([]byte, error) {
	var visited = make(map[uintptr]bool)
	var err error
	visit := func(ptr uintptr) bool {
		if visited[ptr] {
			err = errors.New("converting circular structure to JSON")
			return true
		}
		visited[ptr] = true
		return false
	}
	wrappedValue := marshalCircular(v, visit)
	return []byte(wrappedValue), err
}

func replaceSpecifier(s rune, val goja.Value, buf *bytes.Buffer) bool {
	switch s {
	case 's':
		buf.WriteString(val.String())
	case 'd':
		buf.WriteString(val.ToNumber().String())
	case 'i':
		buf.WriteString(strconv.Itoa(int(val.ToInteger())))
	case 'j':
		data, _ := JSONStringify(val.Export())
		buf.WriteString(string(data))
	case '%':
		buf.WriteByte('%')
		return false
	default:
		buf.WriteByte('%')
		buf.WriteRune(s)
		return false
	}
	return true
}

// Format is a native implementation of Node.js util.format(). This function replaces format specifiers
// with the provided goja values and returns the resulting string as a goja.Value.
// Supported format specifiers: %s, %d, %i, %j, %%.
func Format(runtime *goja.Runtime, format string, args ...goja.Value) goja.Value {
	pct := false
	argNum := 0
	buf := &bytes.Buffer{}
	for _, chr := range format {
		if pct {
			pct = false
			if argNum >= len(args) {
				buf.WriteByte('%')
				buf.WriteRune(chr)
			} else if replaceSpecifier(chr, args[argNum], buf) {
				argNum++
			}
		} else if chr == '%' {
			pct = true
		} else {
			buf.WriteRune(chr)
		}
	}

	for _, arg := range args[argNum:] {
		buf.WriteByte(' ')
		buf.WriteString(arg.String())
	}
	return runtime.ToValue(buf.String())
}
