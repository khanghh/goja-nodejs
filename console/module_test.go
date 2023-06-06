package console

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/khanghh/goja-nodejs/require"
	"github.com/khanghh/goja-nodejs/util"
)

func TestConsole(t *testing.T) {
	vm := goja.New()

	registry := require.NewRegistry()
	registry.RegisterNativeModule(util.ModuleName, util.Default())
	registry.RegisterNativeModule(ModuleName, Default())
	registry.Enable(vm)

	if c := vm.Get("console"); c == nil {
		t.Fatal("console not found")
	}

	if _, err := vm.RunString("console.log('')"); err != nil {
		t.Fatal("console.log() error", err)
	}

	if _, err := vm.RunString("console.error('')"); err != nil {
		t.Fatal("console.error() error", err)
	}

	if _, err := vm.RunString("console.warn('')"); err != nil {
		t.Fatal("console.warn() error", err)
	}
}

func TestConsoleWithPrinter(t *testing.T) {
	vm := goja.New()

	var lastPrint string
	printer := PrinterFunc(func(s string) {
		lastPrint = s
	})

	registry := new(require.Registry)
	registry.RegisterNativeModule(util.ModuleName, util.Default())
	registry.RegisterNativeModule(ModuleName, NewWithPrinter(printer))
	registry.Enable(vm)

	if c := vm.Get("console"); c == nil {
		t.Fatal("console not found")
	}

	if _, err := vm.RunString("console.log('log')"); err != nil {
		t.Fatal("console.log() error", err)
	}
	if lastPrint != "log" {
		t.Fatal("lastPrint not 'log'", lastPrint)
	}

	if _, err := vm.RunString("console.error('error')"); err != nil {
		t.Fatal("console.error() error", err)
	}
	if lastPrint != "error" {
		t.Fatal("lastPrint not 'error'", lastPrint)
	}

	if _, err := vm.RunString("console.warn('warn')"); err != nil {
		t.Fatal("console.warn() error", err)
	}
	if lastPrint != "warn" {
		t.Fatal("lastPrint not 'warn'", lastPrint)
	}
}
