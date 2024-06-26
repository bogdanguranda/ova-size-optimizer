package aggregate

import (
	"fmt"
	"strings"
)

func DebugMapPrint(entries map[string]*Info) {
	fmt.Println("Entries print:")
	fmt.Println(strings.Repeat("-", 40))

	for key, info := range entries {
		fmt.Printf("%-30s\t%d %s\t%s \n", key, info.Count, info.Size, info.InstalledSize)
	}
}

func DebugRuntimeMapPrint(entries map[string]map[string]*Info) {
	fmt.Println("Entries print:")
	fmt.Println(strings.Repeat("-", 40))

	for outerKey, innerMap := range entries {
		fmt.Printf("%-30s\n", outerKey)
		for innerKey, value := range innerMap {
			fmt.Printf("\t%-30s: %s", innerKey, fmt.Sprintf("%d %s\t%s \n", value.Count, value.Size, value.InstalledSize))
		}
		fmt.Println(strings.Repeat("-", 40)) // Separator for each outer key
	}
}

func GetOnlyDuplicates(entries map[string]*Info) map[string]*Info {
	duplicates := make(map[string]*Info)
	for key, info := range entries {
		if info.Count >= 2 {
			duplicates[key] = info
		}
	}
	return duplicates
}

func GetOnlyDuplicatesRuntimes(entries map[string]map[string]*Info) map[string]*Info {
	globalCounts := make(map[string]int)
	for _, infoMap := range entries {
		for key, info := range infoMap {
			globalCounts[key] += info.Count
		}
	}

	duplicates := make(map[string]*Info)
	for _, infoMap := range entries {
		for key, info := range infoMap {
			if globalCounts[key] >= 2 {
				info.Count = globalCounts[key]
				duplicates[key] = info
			}
		}
	}

	return duplicates
}

func ConvertSizeBytesToHumanReadableString(sizeBytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	if sizeBytes == 0 {
		return "0B"
	}
	size := float64(sizeBytes)
	i := 0
	for size >= 1024 && i < len(units)-1 {
		size /= 1024 // Divide size by 1024 to convert to the next higher unit
		i++
	}

	return fmt.Sprintf("%.2f%s", size, units[i])
}
