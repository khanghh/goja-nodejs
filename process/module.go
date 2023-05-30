package process

import (
	"os"
	"strings"

	"github.com/dop251/goja"
)

const ModuleName = "node:process"

var defaultModule = ProcessModule{}

func loadProcessEnv() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		env[envKeyValue[0]] = envKeyValue[1]
	}
	return env
}

type ProcessModule struct {
}

func (m *ProcessModule) Export(runtime *goja.Runtime, module *goja.Object) {
	process := runtime.Get("process").(*goja.Object)
	module.Set("exports", process.Get("env"))
}

func (m *ProcessModule) Enable(runtime *goja.Runtime) {
	env := loadProcessEnv()
	process := runtime.NewObject()
	process.Set("env", env)
	runtime.Set("process", process)
}

func Default() *ProcessModule {
	return &defaultModule
}
