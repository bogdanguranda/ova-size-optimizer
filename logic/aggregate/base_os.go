package aggregate

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

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

func GetBaseImageSize(individualArchivePath string) int64 {
	ctx := context.Background()

	ref, err := alltransports.ParseImageName(fmt.Sprintf("docker-archive:%s", individualArchivePath))
	if err != nil {
		log.Fatalf("Error parsing image name: %v", err)
	}

	sysCtx := &types.SystemContext{}
	imgSrc, err := ref.NewImageSource(ctx, sysCtx)
	if err != nil {
		log.Fatalf("Error getting image source: %v", err)
	}
	defer imgSrc.Close()

	img, err := image.FromUnparsedImage(ctx, sysCtx, image.UnparsedInstance(imgSrc, nil))
	if err != nil {
		log.Fatalf("Error parsing manifest for image: %v", err)
	}

	imgInspect, err := img.Inspect(ctx)
	if err != nil {
		log.Fatalf("Error during inspect: %v", err)
	}

	// the first layer contain information regarding the base os of the container
	// fist layer is the FROM directive

	return imgInspect.LayersData[0].Size
}
