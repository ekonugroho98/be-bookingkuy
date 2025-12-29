package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/config"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/sync"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Parse flags
	countryCode := flag.String("country", "ID", "Country code to sync (default: ID for Indonesia)")
	limit := flag.Int("limit", 0, "Max records to sync (0 = unlimited)")
	dryRun := flag.Bool("dry-run", false, "Print what would be synced without saving")
	flag.Parse()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init()

	// Initialize database
	database, err := db.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize HotelBeds client
	hotelbedsClient := hotelbeds.NewClient(
		os.Getenv("BOOKINGKUY_HOTELBEDS_API_KEY"),
		os.Getenv("BOOKINGKUY_HOTELBEDS_SECRET"),
		os.Getenv("BOOKINGKUY_HOTELBEDS_BASE_URL"),
	)

	// Initialize sync service
	syncService := sync.NewService(database, hotelbedsClient)

	// Run sync
	ctx := context.Background()
	opts := sync.SyncOptions{
		CountryCode: *countryCode,
		Limit:       *limit,
		DryRun:      *dryRun,
	}

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("   HotelBeds Destination Sync")
	fmt.Println("==============================================")
	fmt.Printf("Country Code: %s\n", *countryCode)
	fmt.Printf("Limit: %d (0 = unlimited)\n", *limit)
	fmt.Printf("Dry Run: %v\n", *dryRun)
	fmt.Println("==============================================")
	fmt.Println()

	startTime := time.Now()

	result, err := syncService.SyncDestinations(ctx, opts)
	if err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	duration := time.Since(startTime)

	// Print results
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("   Sync Results")
	fmt.Println("==============================================")
	fmt.Printf("Total:       %d\n", result.Total)
	fmt.Printf("Inserted:    %d\n", result.Inserted)
	fmt.Printf("Updated:     %d\n", result.Updated)
	fmt.Printf("Failed:      %d\n", result.Failed)
	fmt.Printf("Skipped:     %d\n", result.Skipped)
	fmt.Printf("Duration:    %v\n", duration)
	fmt.Printf("Service:     %v\n", result.Duration)
	fmt.Println("==============================================")

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  Errors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			if i >= 10 {
				fmt.Printf("  ... and %d more errors\n", len(result.Errors)-10)
				break
			}
			fmt.Printf("  - [%s] %s: %s\n", err.Code, err.Record, err.Message)
		}
	}

	fmt.Println()

	if result.Failed > 0 {
		os.Exit(1)
	}
}
