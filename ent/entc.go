//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	if err := entc.Generate("./schema", &gen.Config{
		Target:  ".",
		Package: "github.com/shuTwT/nex-api/ent",
	}); err != nil {
		log.Fatal("running ent codegen: ", err)
	}
}
