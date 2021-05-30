# goper

-----
Goper is wrapped a goroutine and channal structure, 
it's safety to close goroutine and channal after handle all task,
witch had putting into channal.

-----

## 使用说明

- frist, import goper into code: `import github/min1324/goper`
- use Auto or Default mode to run.

- use Put(arg), send arg to handle.

- finally,Close() to release and exit goroutine.



## API

```
type Handler func(interface{})
Close()
Default(maxGo int, hd Handler) error
Put(arg interface{}) error
```



## 使用示例

```go
package main

import (
	"fmt"
	"github.com/min1324/goper"
)

func main() {
	var p goper.Goper
	p.Auto(1024,Handler)
	defer p.Close()

	p.Put("hello world.")
}

func Handler(i interface{}) {
	v, ok := i.(string)
	if !ok {
		return
	}
	fmt.Println(v)
}

```