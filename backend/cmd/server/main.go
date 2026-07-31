package main

import (
	_ "ariga.io/atlas/sql/migrate"
	_ "entgo.io/ent"
	_ "github.com/alexedwards/scs/v2"
	_ "github.com/dop251/goja"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

func main() {
	// TODO: Add server wiring after the backend migration scaffold is complete.
}
