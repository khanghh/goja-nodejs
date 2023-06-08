package util

import (
	"github.com/dop251/goja"
)

const ModuleName = "node:util"

var defaultModule = UtilModule{}

type Util struct {
	runtime *goja.Runtime
}

func (u *Util) format(call goja.FunctionCall) goja.Value {
	var fmt string
	if arg := call.Argument(0); !goja.IsUndefined(arg) {
		fmt = arg.String()
	}

	var args []goja.Value
	if len(call.Arguments) > 0 {
		args = call.Arguments[1:]
	}

	return Format(u.runtime, fmt, args...)
}

type UtilModule struct {
}

func (m *UtilModule) Enable(runtime *goja.Runtime) {
}

func (m *UtilModule) Export(runtime *goja.Runtime, module *goja.Object) {
	util := &Util{runtime}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("format", util.format)
}

func Default() *UtilModule {
	return &defaultModule
}
