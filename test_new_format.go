package main

import (
	"fmt"
	"thaistockanalysis/internal/services"
)

func main() {
	telegramService := services.NewTelegramService(
		"7912088515:AAFn3YbnE-84MmMgvhoc6vpJ5HiLPtH5IEg",
		"-1002240874831",
	)

	fmt.Println("Testing new Telegram message format...")

	// Test morning session
	err := telegramService.SendMarketUpdate(
		"Morning Session Open",
		"1320.75",
		"+22.50",
		"2025-10-05",
	)

	if err != nil {
		fmt.Printf("❌ Morning Error: %v\n", err)
	} else {
		fmt.Printf("✅ Morning notification sent!\n")
	}

	// Test afternoon session
	err = telegramService.SendMarketUpdate(
		"Afternoon Session Open",
		"1325.25",
		"+4.50",
		"2025-10-05",
	)

	if err != nil {
		fmt.Printf("❌ Afternoon Error: %v\n", err)
	} else {
		fmt.Printf("✅ Afternoon notification sent!\n")
	}
}
