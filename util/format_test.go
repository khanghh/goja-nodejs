package util

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/khanghh/goja-nodejs/require"
)

func TestUtil_Format(t *testing.T) {
	vm := goja.New()
	ret := Format(vm, "Test: %% %д %s %d, %i, %j", vm.ToValue("string"), vm.ToValue(42), vm.ToValue(1.2), vm.NewObject())
	if res := ret.String(); res != "Test: % %д string 42, 1, {}" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_NoArgs(t *testing.T) {
	vm := goja.New()
	ret := Format(vm, "Test: %s %d, %j")
	if res := ret.String(); res != "Test: %s %d, %j" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_Circular_JSON(t *testing.T) {
	vm := goja.New()
	testObj := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john.doe@example.com",
	}
	testObj["data"] = testObj
	obj := vm.ToValue(testObj).ToObject(vm)
	ret := Format(vm, "Test: %j", obj)
	if res := ret.String(); res != `Test: {"name":"John Doe","age":30,"email":"john.doe@example.com","data":[Circular]}` {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_LessArgs(t *testing.T) {
	vm := goja.New()
	ret := Format(goja.New(), "Test: %s %d, %j", vm.ToValue("string"), vm.ToValue(42))
	if res := ret.String(); res != "Test: string 42, %j" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_MoreArgs(t *testing.T) {
	vm := goja.New()
	ret := Format(vm, "Test: %s %d, %j", vm.ToValue("string"), vm.ToValue(42), vm.NewObject(), vm.ToValue(42.42))
	if res := ret.String(); res != "Test: string 42, {} 42.42" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_RequireUtilModule(t *testing.T) {
	vm := goja.New()
	registry := require.NewRegistry()
	registry.RegisterNativeModule(ModuleName, Default())
	registry.Enable(vm)

	util, err := require.Require(vm, ModuleName)
	if err != nil {
		t.Fatalf("module %s not found", ModuleName)
	}
	utilObj := util.(*goja.Object)
	if format, ok := goja.AssertFunction(utilObj.Get("format")); ok {
		res, err := format(util)
		if err != nil {
			t.Fatal(err)
		}
		if v := res.Export(); v != "" {
			t.Fatalf("Unexpected result: %v", v)
		}
	}
}
