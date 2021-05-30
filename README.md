# goper

-----
Goper is wrapped a goroutine and channal structure, 
it's safety to close goroutine and channal after handle all task,
witch had putting into channal.

-----

## 使用说明

- frist, import goper into code: `import github/min1324/goper`
- use Default to run.

- use Deliver(arg), send arg to handle.

- finally,Close() to release and exit goroutine.



## API

```
type Handler func(interface{})
Close()
Default(maxGo int, hd Handler) error
Deliver(arg interface{}) error
```



## 使用示例

```go
package main

import (
	"fmt"
	"github/min1324/goper"
)

func main() {
	var g goper.Goper
	g.Default(1, Router)
	defer g.Close()

	g.Deliver("hello world.")
}

func Router(i interface{}) {
	s, ok := i.(string)
	if ok {
		fmt.Println(s)
	}
}


```