package main

import (
    "flag"
    "log"
    "os"
    
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
	"minitwit/src/model"
)

func main() {
    // Parse command line flags
    dbPath := flag.String("db", "/tmp/minitwit.db", "database file path")
    reset := flag.Bool("reset", false, "kill database (drop tables)")
    flag.Parse()

    // Connect to database
    db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Handle reset if requested
    if *reset {
        log.Println("Killing database")
        err = db.Migrator().DropTable(&model.User{}, &model.Follower{}, &model.Message{})
        if err != nil {
            log.Fatal("Failed to drop tables:", err)
        }
        log.Println("Tables dropped")
    }

    // Run migrations
    log.Println("Running migrations")
    err = db.AutoMigrate(
        &model.User{},
        &model.Follower{},
        &model.Message{},
    )
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    log.Println("Migrations complete!")

	// Seeding the development database
    if os.Getenv("APP_ENV") == "development" {
        seedData(db)
    }
}

func seedData(db *gorm.DB) {
    log.Println("Seeding development data")
    
    // Add some test users
    users := []model.User{
        {Username: "alice", Email: "alice@example.com", PwHash: "hash1"},
        {Username: "bob", Email: "bob@example.com", PwHash: "hash2"},
    }
    
    for _, user := range users {
        db.FirstOrCreate(&user, model.User{Username: user.Username})
    }
    
    log.Println("✅ Seed data complete")
}
