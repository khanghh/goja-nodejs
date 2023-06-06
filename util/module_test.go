package util

import (
	"bytes"
	"testing"

	"github.com/dop251/goja"
	"github.com/khanghh/goja-nodejs/require"
)

func TestUtil_Format(t *testing.T) {
	vm := goja.New()
	util := &Util{vm}

	var b bytes.Buffer
	util.Format(&b, "Test: %% %д %s %d, %j", vm.ToValue("string"), vm.ToValue(42), vm.NewObject())

	if res := b.String(); res != "Test: % %д string 42, {}" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_NoArgs(t *testing.T) {
	vm := goja.New()
	util := &Util{vm}

	var b bytes.Buffer
	util.Format(&b, "Test: %s %d, %j")

	if res := b.String(); res != "Test: %s %d, %j" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_LessArgs(t *testing.T) {
	vm := goja.New()
	util := &Util{vm}

	var b bytes.Buffer
	util.Format(&b, "Test: %s %d, %j", vm.ToValue("string"), vm.ToValue(42))

	if res := b.String(); res != "Test: string 42, %j" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestUtil_Format_MoreArgs(t *testing.T) {
	vm := goja.New()
	util := &Util{vm}

	var b bytes.Buffer
	util.Format(&b, "Test: %s %d, %j", vm.ToValue("string"), vm.ToValue(42), vm.NewObject(), vm.ToValue(42.42))

	if res := b.String(); res != "Test: string 42, {} 42.42" {
		t.Fatalf("Unexpected result: '%s'", res)
	}
}

func TestRequireUtilModule(t *testing.T) {
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
