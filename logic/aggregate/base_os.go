package aggregate

import "regexp"

func DetectOSNameWithVersion(distro string) (osNameWithVersion string) {
	re := regexp.MustCompile(`pkg:generic/([^@]+)@([^?]+)`)
	matches := re.FindStringSubmatch(distro)
	if len(matches) > 2 {
		osName := matches[1]
		version := matches[2]
		osNameWithVersion = osName + ":" + version
	}
	return osNameWithVersion
}
