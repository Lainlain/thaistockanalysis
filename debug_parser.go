package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type MarketSession struct {
	Index  float64
	Change float64
}

func parseSessionOpeningData(filename, sessionType string) (*MarketSession, error) {
	fmt.Printf("Debug: Looking for %s session data in file: %s\n", sessionType, filename)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var targetSection string
	if sessionType == "morning" {
		targetSection = "## Morning Session"
	} else {
		targetSection = "## Afternoon Session"
	}
	fmt.Printf("Debug: Looking for section containing: '%s'\n", targetSection)

	inTargetSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fmt.Printf("Debug: Reading line: '%s'\n", line)

		// Check for session headers
		if strings.Contains(line, targetSection) {
			inTargetSection = true
			fmt.Printf("Debug: Found target section '%s' in line: %s\n", targetSection, line)
			continue
		}

		// Stop if we hit another level 2 section (##), but not level 3 (###)
		if inTargetSection && strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "###") && !strings.Contains(line, targetSection) {
			fmt.Printf("Debug: Hit another level 2 section, stopping: %s\n", line)
			break
		}

		// Look for index pattern: "* Index: 1295.80 (+5.15)" or "* Open Index: 1295.80 (+5.15)"
		// But exclude "Close Index:" which is for close data, not open data
		if inTargetSection && (strings.Contains(line, "Index:") && !strings.Contains(line, "Close Index:")) {
			fmt.Printf("Debug: Found Index line in %s section: %s\n", sessionType, line)
			// Extract index and change using regex
			re := regexp.MustCompile(`(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)`)
			matches := re.FindStringSubmatch(line)

			if len(matches) >= 3 {
				indexVal, err1 := strconv.ParseFloat(matches[1], 64)
				changeVal, err2 := strconv.ParseFloat(matches[2], 64)

				if err1 == nil && err2 == nil {
					fmt.Printf("Debug: Successfully parsed index: %.2f, change: %.2f\n", indexVal, changeVal)
					return &MarketSession{
						Index:  indexVal,
						Change: changeVal,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not find opening data for %s session in file %s", sessionType, filename)
}

func main() {
	result, err := parseSessionOpeningData("articles/2025-10-03.md", "morning")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: Index=%.2f, Change=%.2f\n", result.Index, result.Change)
	}
}
