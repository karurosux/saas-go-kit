package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	subscriptiongorm "github.com/karurosux/saas-go-kit/subscription-go/repositories/gorm"
	teamgorm "github.com/karurosux/saas-go-kit/team-go/repositories/gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var (
		outputDir = flag.String("output", "./migrations", "Output directory for migration files")
		dsn       = flag.String("dsn", "", "Database connection string (optional - for schema diff)")
		module    = flag.String("module", "all", "Module to generate migrations for (subscription, team, all)")
	)
	flag.Parse()

	fmt.Printf("🚀 SaaS Go Kit Migration Generator\n\n")

	if *dsn != "" {
		fmt.Printf("🔍 Generating migrations from database schema...\n")
		generateFromDB(*dsn, *outputDir, *module)
	} else {
		fmt.Printf("📄 Generating migrations from model definitions...\n")
		generateFromModels(*outputDir, *module)
	}
}

func generateFromModels(outputDir, module string) {
	switch module {
	case "subscription":
		if err := subscriptiongorm.GenerateMigrationSQL(outputDir); err != nil {
			log.Fatalf("Failed to generate subscription migrations: %v", err)
		}
	case "team":
		if err := teamgorm.GenerateMigrationSQL(outputDir); err != nil {
			log.Fatalf("Failed to generate team migrations: %v", err)
		}
	case "all":
		if err := subscriptiongorm.GenerateMigrationSQL(outputDir); err != nil {
			log.Fatalf("Failed to generate subscription migrations: %v", err)
		}
		if err := teamgorm.GenerateMigrationSQL(outputDir); err != nil {
			log.Fatalf("Failed to generate team migrations: %v", err)
		}
	default:
		log.Fatalf("Unknown module: %s. Use 'subscription', 'team', or 'all'", module)
	}

	fmt.Printf("\n✅ Migration generation completed!\n")
	fmt.Printf("📁 Files saved to: %s\n\n", outputDir)
	fmt.Printf("🔧 Next steps:\n")
	fmt.Printf("   1. Review the generated SQL files\n")
	fmt.Printf("   2. Test in a staging environment\n")
	fmt.Printf("   3. Apply to production:\n")
	fmt.Printf("      psql -d your_db -f %s/*_up.sql\n\n", outputDir)
}

func generateFromDB(dsn, outputDir, module string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	switch module {
	case "subscription":
		if err := subscriptiongorm.GenerateMigrationSQLFromDB(db, outputDir); err != nil {
			log.Fatalf("Failed to generate subscription migrations: %v", err)
		}
	case "team":
		if err := teamgorm.GenerateMigrationSQLFromDB(db, outputDir); err != nil {
			log.Fatalf("Failed to generate team migrations: %v", err)
		}
	case "all":
		if err := subscriptiongorm.GenerateMigrationSQLFromDB(db, outputDir); err != nil {
			log.Fatalf("Failed to generate subscription migrations: %v", err)
		}
		if err := teamgorm.GenerateMigrationSQLFromDB(db, outputDir); err != nil {
			log.Fatalf("Failed to generate team migrations: %v", err)
		}
	default:
		log.Fatalf("Unknown module: %s. Use 'subscription', 'team', or 'all'", module)
	}

	fmt.Printf("\n✅ Migration generation from database completed!\n")
	fmt.Printf("📁 Files saved to: %s\n\n", outputDir)
}