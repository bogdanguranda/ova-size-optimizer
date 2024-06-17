package aggregate

import "strings"

func DetectRuntime(pkgURL string, runtimes map[string]*Info) {
	runtime := ""
	if strings.Contains(pkgURL, "pkg:generic/java/") || strings.Contains(pkgURL, "pkg:maven/") {
		if strings.Contains(pkgURL, "jre") || strings.Contains(pkgURL, "jdk") {
			runtime = "Java"
		}
	} else if strings.Contains(pkgURL, "pkg:python/") {
		if strings.Contains(pkgURL, "cpython") {
			runtime = "Python"
		}
	}
	// TODO: Add more runtime detection rules as needed

	if runtime != "" {
		if runtimes[runtime] == nil {
			runtimes[runtime] = &Info{
				Count: 1,
				Size:  "0", // TODO: implement size determination for runtimes
			}
		} else {
			runtimes[runtime].Count++
		}
	}
}
