package require

import (
	"encoding/json"
	"errors"
	"path"
	"strings"

	"github.com/dop251/goja"
)

type ModuleResolver struct {
	registry    *Registry
	runtime     *goja.Runtime
	modules     map[string]*goja.Object
	nodeModules map[string]*goja.Object
}

// Nodejs module search algorithm described by
// https://nodejs.org/api/modules.html#modules_all_together
func (r *ModuleResolver) resolve(modpath string) (module *goja.Object, err error) {
	origPath, modpath := modpath, path.Clean(modpath)
	if modpath == "" {
		return nil, ErrInvalidModule
	}

	var start string
	err = nil
	if path.IsAbs(origPath) {
		start = "/"
	} else {
		start = r.getCurrentModulePath()
	}

	p := path.Join(start, modpath)
	if strings.HasPrefix(origPath, "./") ||
		strings.HasPrefix(origPath, "/") || strings.HasPrefix(origPath, "../") ||
		origPath == "." || origPath == ".." {
		if module = r.modules[p]; module != nil {
			return
		}
		module, err = r.loadAsFileOrDirectory(p)
		if err == nil && module != nil {
			r.modules[p] = module
		}
	} else {
		module, err = r.loadNative(origPath)
		if err == nil {
			return
		} else {
			err = nil
		}
		if module = r.nodeModules[p]; module != nil {
			return
		}
		module, err = r.loadNodeModules(modpath, start)
		if err == nil && module != nil {
			r.nodeModules[p] = module
		}
	}

	if module == nil && err == nil {
		err = ErrInvalidModule
	}
	return
}

func (r *ModuleResolver) loadNative(name string) (*goja.Object, error) {
	module := r.modules[name]
	if module != nil {
		return module, nil
	}

	if native := r.registry.natives[name]; native != nil {
		module = r.createModuleObject()
		native.Export(r.runtime, module)
		r.modules[name] = module
		return module, nil
	}
	return nil, ErrInvalidModule
}

func (r *ModuleResolver) loadAsFileOrDirectory(path string) (module *goja.Object, err error) {
	if module, err = r.loadAsFile(path); module != nil || err != nil {
		return
	}

	return r.loadAsDirectory(path)
}

func (r *ModuleResolver) loadAsFile(path string) (module *goja.Object, err error) {
	if module, err = r.loadModule(path); module != nil || err != nil {
		return
	}

	p := path + ".js"
	if module, err = r.loadModule(p); module != nil || err != nil {
		return
	}

	p = path + ".json"
	return r.loadModule(p)
}

func (r *ModuleResolver) loadIndex(modpath string) (module *goja.Object, err error) {
	p := path.Join(modpath, "index.js")
	if module, err = r.loadModule(p); module != nil || err != nil {
		return
	}

	p = path.Join(modpath, "index.json")
	return r.loadModule(p)
}

func (r *ModuleResolver) loadAsDirectory(modpath string) (module *goja.Object, err error) {
	p := path.Join(modpath, "package.json")
	buf, err := r.registry.getSource(p)
	if err != nil {
		return r.loadIndex(modpath)
	}
	var pkg struct {
		Main string
	}
	err = json.Unmarshal(buf, &pkg)
	if err != nil || len(pkg.Main) == 0 {
		return r.loadIndex(modpath)
	}

	m := path.Join(modpath, pkg.Main)
	if module, err = r.loadAsFile(m); module != nil || err != nil {
		return
	}

	return r.loadIndex(m)
}

func (r *ModuleResolver) loadNodeModule(modpath, start string) (*goja.Object, error) {
	return r.loadAsFileOrDirectory(path.Join(start, modpath))
}

func (r *ModuleResolver) loadNodeModules(modpath, start string) (module *goja.Object, err error) {
	for _, dir := range r.registry.globalFolders {
		if module, err = r.loadNodeModule(modpath, dir); module != nil || err != nil {
			return
		}
	}
	for {
		var p string
		if path.Base(start) != "node_modules" {
			p = path.Join(start, "node_modules")
		} else {
			p = start
		}
		if module, err = r.loadNodeModule(modpath, p); module != nil || err != nil {
			return
		}
		if start == ".." { // Dir('..') is '.'
			break
		}
		parent := path.Dir(start)
		if parent == start {
			break
		}
		start = parent
	}

	return
}

func (r *ModuleResolver) getCurrentModulePath() string {
	var buf [2]goja.StackFrame
	frames := r.runtime.CaptureCallStack(2, buf[:0])
	if len(frames) < 2 {
		return "."
	}
	return path.Dir(frames[1].SrcName())
}

func (r *ModuleResolver) createModuleObject() *goja.Object {
	module := r.runtime.NewObject()
	module.Set("exports", r.runtime.NewObject())
	return module
}

func (r *ModuleResolver) loadModule(path string) (*goja.Object, error) {
	module := r.modules[path]
	if module == nil {
		module = r.createModuleObject()
		r.modules[path] = module
		err := r.loadModuleFile(path, module)
		if err != nil {
			module = nil
			delete(r.modules, path)
			if errors.Is(err, ErrModuleNotExist) {
				err = nil
			}
		}
		return module, err
	}
	return module, nil
}

func (r *ModuleResolver) loadModuleFile(path string, gojaModule *goja.Object) error {

	prg, err := r.registry.getCompiledSource(path)

	if err != nil {
		return err
	}

	f, err := r.runtime.RunProgram(prg)
	if err != nil {
		return err
	}

	if call, ok := goja.AssertFunction(f); ok {
		gojaExports := gojaModule.Get("exports")
		gojaRequire := r.runtime.Get("require")

		// Run the module source, with "gojaExports" as "this",
		// "gojaExports" as the "exports" variable, "gojaRequire"
		// as the "require" variable and "gojaModule" as the
		// "module" variable (Nodegoja capable).
		_, err = call(gojaExports, gojaExports, gojaRequire, gojaModule)
		if err != nil {
			return err
		}
	} else {
		return ErrInvalidModule
	}

	return nil
}

func (r *ModuleResolver) require(call goja.FunctionCall) goja.Value {
	ret, err := r.Require(call.Argument(0).String())
	if err != nil {
		if _, ok := err.(*goja.Exception); !ok {
			panic(r.runtime.NewGoError(err))
		}
		panic(err)
	}
	return ret
}

// Require can be used to import modules from Go source (similar to goja require() function).
func (r *ModuleResolver) Require(p string) (ret goja.Value, err error) {
	module, err := r.resolve(p)
	if err != nil {
		return
	}
	ret = module.Get("exports")
	return
}

func Require(runtime *goja.Runtime, name string) (goja.Value, error) {
	if require, ok := goja.AssertFunction(runtime.Get("require")); ok {
		module, err := require(goja.Undefined(), runtime.ToValue(name))
		if err != nil {
			return nil, ErrModuleNotExist
		}
		return module, nil
	}
	panic(runtime.NewTypeError("Please enable require for this runtime using new(require.Registry).Enable(runtime)"))
}
