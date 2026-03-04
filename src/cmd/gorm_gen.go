package db_utils

import (
	"minitwit/src/model"
	"minitwit/src/query"

	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
)


func CreateDAO() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../query",
		Mode: gen.WithoutContext|gen.WithDefaultQuery|gen.WithQueryInterface, // generate mode
	})

	gormdb, _ := gorm.Open(sqlite.Open("/tmp/minitwit.db"),&gorm.Config{})
	g.UseDB(gormdb) // reuse your gorm db

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	g.ApplyBasic(
			model.User{},
			model.Follower{},
			model.Message{},
		)

	// Generate Type Safe API with Dynamic SQL defined on Querier interface for `model.User` and `model.Company`
	g.ApplyInterface(func(query.Querier) {}, 
		model.User{}, 
		model.Follower{}, 
		model.Message{},
	)

	// Generate the code
	g.Execute()
}


func Migrate() {

}