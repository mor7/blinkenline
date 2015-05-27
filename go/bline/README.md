# bline

This is a Go library for the Blinkenline.

## Initialisation

All flags needed by the library are parsed by itself.

```
package main

import (
	"github.com/mor7/blinkenline/go/bline"
)

func main() {
	err := bline.Init()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bline.Close()
}
```
