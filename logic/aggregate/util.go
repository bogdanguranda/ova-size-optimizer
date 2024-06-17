package aggregate

import (
	"fmt"
	"strings"
)

func DebugMapPrint(entries map[string]*Info) {
	fmt.Println("Entries print:")
	fmt.Println(strings.Repeat("-", 40))

	for key, info := range entries {
		fmt.Printf("%-30s\t%d %s\n", key, info.Count, info.Size)
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
