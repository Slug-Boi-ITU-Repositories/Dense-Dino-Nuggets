package main

import (
	"flag"
	"log"

	"minitwit/src/model"
	"minitwit/src/query"

	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
    dbPath := flag.String("db", "/tmp/minitwit.db", "database file path")
    outPath := flag.String("out", "./query", "output path for generated code")
    flag.Parse()

    // Connect to database (needed for schema introspection)
    gormdb, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    log.Println("Generating DAO code")

    g := gen.NewGenerator(gen.Config{
        OutPath: *outPath,
        Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
    })
    g.UseDB(gormdb)

    // Generate basic CRUD
    g.ApplyBasic(
        model.User{},
        model.Follower{},
        model.Message{},
    )

    // Generate custom queries from interface
    g.ApplyInterface(
        func(query.Querier) {},
        model.User{},
        model.Follower{},
        model.Message{},
    )

    g.Execute()
    log.Println("DAO code generated successfully")
}
