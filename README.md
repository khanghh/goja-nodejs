Nodejs compatibility library for Goja
====

This is a collection of [goja](https://github.com/dop251/goja) modules that provide nodejs compatibility.

Example:

```go
package main

import (
    "github.com/dop251/goja"
    "github.com/khanghh/goja-nodejs/require"
    "github.com/khanghh/goja-nodejs/util"
    "github.com/khanghh/goja-nodejs/console"
)

func main() {
    runtime := goja.New()
    registry := require.NewRegistry() // this can be shared by multiple runtimes
    registry.RegisterNativeModule(util.ModuleName, util.Default())
    registry.RegisterNativeModule(console.ModuleName, console.Default())
    req := registry.Enable(runtime)

    runtime.RunString(`
    var m = require("./m.js");
    m.test();
    `)

    m, err := req.Require("./m.js")
    _, _ = m, err
}

func example() {
  vm := goja.New()
  console.Default().Enable(vm)
  vm.RunString(`
  console.log("Hello, %s!", "World");
  `)
}
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.