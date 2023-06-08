package console

import (
	"log"

	"github.com/dop251/goja"
	"github.com/khanghh/goja-nodejs/util"
)

const ModuleName = "node:console"

var defaultModule = ConsoleModule{
	printer: DefaultPrinter,
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

type Console struct {
	runtime *goja.Runtime
	printer Printer
}

func (c *Console) formatPrinterOutput(call goja.FunctionCall) string {
	var format string
	if arg := call.Argument(0); !goja.IsUndefined(arg) {
		format = arg.String()
	}
	var args []goja.Value
	if len(call.Arguments) > 0 {
		args = call.Arguments[1:]
	}
	out := util.Format(c.runtime, format, args...)
	return out.String()
}

func (c *Console) log(call goja.FunctionCall) goja.Value {
	c.printer.Log(c.formatPrinterOutput(call))
	return goja.Undefined()
}

func (c *Console) warn(call goja.FunctionCall) goja.Value {
	c.printer.Warn(c.formatPrinterOutput(call))
	return goja.Undefined()
}

func (c *Console) error(call goja.FunctionCall) goja.Value {
	c.printer.Error(c.formatPrinterOutput(call))
	return goja.Undefined()
}

type Option func(*ConsoleModule)

type ConsoleModule struct {
	printer Printer
}

func (m *ConsoleModule) Enable(runtime *goja.Runtime) {
	console := &Console{
		runtime: runtime,
		printer: m.printer,
	}
	obj := runtime.NewObject()
	obj.Set("log", console.log)
	obj.Set("warn", console.warn)
	obj.Set("error", console.error)
	runtime.Set("console", obj)
}

func (m *ConsoleModule) Export(runtime *goja.Runtime, module *goja.Object) {
}

func WithPrinter(printer Printer) Option {
	return func(cm *ConsoleModule) {
		cm.printer = printer
	}
}

func New(opts ...Option) *ConsoleModule {
	cm := &ConsoleModule{
		printer: DefaultPrinter,
	}
	for _, opt := range opts {
		opt(cm)
	}
	return cm
}

func Default() *ConsoleModule {
	return &defaultModule
}
