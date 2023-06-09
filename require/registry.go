package require

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"text/template"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
)

var (
	ErrInvalidModule     = errors.New("invalid module")
	ErrInvalidModuleName = errors.New("invalid module name")
	ErrModuleNotExist    = errors.New("module does not exist")
)

// NativeModule is an interface that represents a module with native methods and objects.
// It provides the methods to enable global access and to export its functionality.
type NativeModule interface {
	// Enable registers global objects and methods of the module to the JavaScript runtime.
	// This allows using the module's features globally in the runtime environment.
	Enable(runtime *goja.Runtime)

	// Export registers native objects and methods to the module.exports object.
	// This allows importing the module's features into other modules or scripts as needed.
	Export(runtime *goja.Runtime, module *goja.Object)
}

// SourceLoader represents a function that returns a file data at a given path.
// The function should return ModuleFileDoesNotExistError if the file either doesn't exist or is a directory.
// This error will be ignored by the resolver and the search will continue. Any other errors will be propagated.
type SourceLoader func(path string) ([]byte, error)

// DefaultSourceLoader is used if none was set (see WithLoader()). It simply loads files from the host's filesystem.
func DefaultSourceLoader(filename string) ([]byte, error) {
	fp := filepath.FromSlash(filename)
	f, err := os.Open(fp)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = ErrModuleNotExist
		} else if runtime.GOOS == "windows" {
			if errors.Is(err, syscall.Errno(0x7b)) { // ERROR_INVALID_NAME, The filename, directory name, or volume label syntax is incorrect.
				err = ErrModuleNotExist
			}
		}
		return nil, err
	}

	defer f.Close()
	// On some systems (e.g. plan9 and FreeBSD) it is possible to use the standard read() call on directories
	// which means we cannot rely on read() returning an error, we have to do stat() instead.
	if fi, err := f.Stat(); err == nil {
		if fi.IsDir() {
			return nil, ErrModuleNotExist
		}
	} else {
		return nil, err
	}
	return io.ReadAll(f)
}

type Option func(*Registry)

// WithLoader sets a function which will be called by the require() function in order to get a source code for a
// module at the given path. The same function will be used to get external source maps.
// Note, this only affects the modules loaded by the require() function. If you need to use it as a source map
// loader for code parsed in a different way (such as runtime.RunString() or eval()), use (*Runtime).SetParserOptions()
func WithLoader(srcLoader SourceLoader) Option {
	return func(r *Registry) {
		r.srcLoader = srcLoader
	}
}

// WithGlobalFolders appends the given paths to the registry's list of
// global folders to search if the requested module is not found
// elsewhere.  By default, a registry's global folders list is empty.
// In the reference Node.js implementation, the default global folders
// list is $NODE_PATH, $HOME/.node_modules, $HOME/.node_libraries and
// $PREFIX/lib/node, see
// https://nodejs.org/api/modules.html#modules_loading_from_the_global_folders.
func WithGlobalFolders(globalFolders ...string) Option {
	return func(r *Registry) {
		r.globalFolders = globalFolders
	}
}

// Registry contains a cache of compiled modules which can be used by multiple Runtimes
type Registry struct {
	sync.Mutex
	natives  map[string]NativeModule
	compiled map[string]*goja.Program

	srcLoader     SourceLoader
	globalFolders []string
}

func (r *Registry) getSource(p string) ([]byte, error) {
	srcLoader := r.srcLoader
	if srcLoader == nil {
		srcLoader = DefaultSourceLoader
	}
	return srcLoader(p)
}

func (r *Registry) getCompiledSource(filepath string) (*goja.Program, error) {
	r.Lock()
	defer r.Unlock()

	prg := r.compiled[filepath]
	if prg == nil {
		buf, err := r.getSource(filepath)
		if err != nil {
			return nil, err
		}
		srouce := string(buf)

		if path.Ext(filepath) == ".json" {
			srouce = "module.exports = JSON.parse('" + template.JSEscapeString(srouce) + "')"
		}

		source := "(function(exports, require, module) {" + srouce + "\n})"
		parsed, err := goja.Parse(filepath, source, parser.WithSourceMapLoader(r.srcLoader))
		if err != nil {
			return nil, err
		}
		prg, err = goja.CompileAST(parsed, false)
		if err == nil {
			if r.compiled == nil {
				r.compiled = make(map[string]*goja.Program)
			}
			r.compiled[filepath] = prg
		}
		return prg, err
	}
	return prg, nil
}

// Enable adds the require() function to the specified runtime.
func (r *Registry) Enable(runtime *goja.Runtime) *ModuleResolver {
	resolver := &ModuleResolver{
		registry:    r,
		runtime:     runtime,
		modules:     make(map[string]*goja.Object),
		nodeModules: make(map[string]*goja.Object),
	}
	runtime.Set("require", resolver.require)
	for _, module := range r.natives {
		module.Enable(runtime)
	}
	return resolver
}

func (r *Registry) RegisterNativeModule(name string, module NativeModule) {
	r.Lock()
	defer r.Unlock()

	if r.natives == nil {
		r.natives = make(map[string]NativeModule)
	}
	name = path.Clean(name)
	r.natives[name] = module
}

func NewRegistry(opts ...Option) *Registry {
	r := &Registry{}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func NewRegistryWithLoader(srcLoader SourceLoader) *Registry {
	return NewRegistry(WithLoader(srcLoader))
}
