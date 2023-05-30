package console

import (
	"log"

	"github.com/dop251/goja"
	"github.com/khanghh/goja-nodejs/require"
	"github.com/khanghh/goja-nodejs/util"
)

const ModuleName = "node:console"

var defaultModule = ConsoleModule{
	printer: DefaultPrinter,
}

type Console struct {
	runtime *goja.Runtime
	util    *goja.Object
}

type Printer interface {
	Log(string)
	Warn(string)
	Error(string)
}

type PrinterFunc func(s string)

func (print PrinterFunc) Log(s string) { print(s) }

func (print PrinterFunc) Warn(s string) { print(s) }

func (print PrinterFunc) Error(s string) { print(s) }

var DefaultPrinter Printer = PrinterFunc(func(s string) { log.Print(s) })

func (c *Console) log(print func(string)) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if format, ok := goja.AssertFunction(c.util.Get("format")); ok {
			ret, err := format(c.util, call.Arguments...)
			if err != nil {
				panic(err)
			}

			print(ret.String())
		} else {
			panic(c.runtime.NewTypeError("util.format is not a function"))
		}

		return nil
	}
}

type ConsoleModule struct {
	printer Printer
}

func (m *ConsoleModule) Enable(runtime *goja.Runtime) {
	util, err := require.Require(runtime, util.ModuleName)
	if err != nil {
		panic(err)
	}
	console := &Console{
		runtime: runtime,
		util:    util.(*goja.Object),
	}
	obj := runtime.NewObject()
	obj.Set("log", console.log(m.printer.Log))
	obj.Set("error", console.log(m.printer.Error))
	obj.Set("warn", console.log(m.printer.Warn))
	runtime.Set("console", obj)
}

func (m *ConsoleModule) Export(runtime *goja.Runtime, module *goja.Object) {
}

func NewWithPrinter(printer Printer) *ConsoleModule {
	if printer == nil {
		printer = DefaultPrinter
	}
	return &ConsoleModule{printer}
}

func Default() *ConsoleModule {
	return &defaultModule
}
