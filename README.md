# go-rpmdb
Library for enumerating packages in an RPM DB `Packages` file (without bindings).

This is a fork of [knqyf263/go-rpmdb](https://github.com/knqyf263/go-rpmdb) with a
[patch](https://github.com/erikvarga/go-rpmdb/commit/9d196a3ed4c687e68257f5d1a990c8b565ee106c)
to avoid infinite loops on corrupt bdb files.

```go
package main

import (
	"fmt"
	"log"

	rpmdb "github.com/erikvarga/go-rpmdb/pkg"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	db, err := rpmdb.Open("./Packages")
	if err != nil {
		return err
	}
	pkgList, err := db.ListPackages()
	if err != nil {
		return err
	}

	fmt.Println("Packages:")
	for _, pkg := range pkgList {
		fmt.Printf("\t%+v\n", *pkg)
		// {Epoch:0 Name:m4 Version:1.4.16 Release:10.el7 Arch:x86_64}
		// {Epoch:0 Name:zip Version:3.0 Release:11.el7 Arch:x86_64}
		// ...
	}
	fmt.Printf("[Total Packages: %d]\n", len(pkgList))
	return nil
}
```
