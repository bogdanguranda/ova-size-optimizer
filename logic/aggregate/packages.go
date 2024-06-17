package aggregate

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func DetectPackageName(pkgURL string) string {
	parts := strings.Split(pkgURL, "/")
	nameWithVersionRaw := parts[len(parts)-1]
	nameWithVersionFullUrlEncoded := strings.Split(nameWithVersionRaw, "?")[0]

	nameWithVersionUrlDecoded, err := url.QueryUnescape(nameWithVersionFullUrlEncoded)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to decode package URL: %s\n", pkgURL)
		return ""
	}

	return nameWithVersionUrlDecoded
}
