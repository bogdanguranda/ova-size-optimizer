package analyze

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"ova-size-optimizer/logic/load"
)

func DetectRuntime(packageProperties load.Metadata, runtimes map[string]map[string]*Info, fileName string) {
	var runtime string
	const (
		javaInstallLocLib = "/usr/lib/jvm/.*/lib/"
		javaInstallLocBin = "/usr/lib/jvm/.*/bin/"
		pythonInstallBin  = "/usr/bin/python.*"
	)

	libReJava, err := regexp.Compile(javaInstallLocLib)
	if err != nil {
		fmt.Printf("Error compiling java regex: %v\n", err)
	}

	binReJava, err := regexp.Compile(javaInstallLocBin)
	if err != nil {
		fmt.Printf("Error compiling java regex: %v\n", err)
	}

	binRePython, err := regexp.Compile(pythonInstallBin)
	if err != nil {
		fmt.Printf("Error compiling python regex: %v\n", err)
	}

	for _, v := range packageProperties.Files {
		if libReJava.MatchString(v.Path) ||
			binReJava.MatchString(v.Path) ||
			binRePython.MatchString(v.Path) {
			// we want to store as a runtime the distribution of java or python or etc
			// eg. /usr/lib/jvm/zulu-openjdk-5.3/lib/ => [3] is zulu-openjdk-5.3
			// if multiple runtimes are on the same container/image we will have that in the map
			runtimeSlc := strings.FieldsFunc(v.Path, func(c rune) bool {
				return c == '/'
			})

			for _, elem := range runtimeSlc {
				if strings.Contains(elem, "jdk") ||
					strings.Contains(elem, "jre") ||
					strings.Contains(elem, "python") ||
					strings.Contains(elem, "rust") ||
					strings.Contains(elem, "golang") {
					runtime = elem
				}
			}

			if runtimes[fileName][runtime] == nil {
				// new runtime
				// store the count of appearence and add the size
				runtimes[fileName][runtime] = &Info{
					Count: 1,
					Size:  strconv.FormatInt(int64(v.Size), 10),
				}
			} else {
				sizeInt64, err := strconv.ParseInt(runtimes[fileName][runtime].Size, 10, 64)
				if err != nil {
					fmt.Println("Failed to convert string to int")
					return
				}
				runtimes[fileName][runtime].Size = strconv.FormatInt(sizeInt64+int64(v.Size), 10)
			}
		}
	}

}

// runtime := ""
// if strings.Contains(pkgURL, "pkg:generic/java/") || strings.Contains(pkgURL, "pkg:maven/") {
// 	if strings.Contains(pkgURL, "jre") || strings.Contains(pkgURL, "jdk") {
// 		runtime = "Java"
// 	}
// } else if strings.Contains(pkgURL, "pkg:python/") {
// 	if strings.Contains(pkgURL, "cpython") {
// 		runtime = "Python"
// 	}
// }
