package main

import (
	"bufio"
	"os"
	"strings"
)

// LoadLocalizationFile loads localization strings from american.txt
func LoadLocalizationFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "{") || strings.HasPrefix(line, "}") {
			continue
		}

		if strings.Contains(line, " = ") {
			parts := strings.SplitN(line, " = ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				localizationMap[key] = value
			}
		}
	}

	return scanner.Err()
}
