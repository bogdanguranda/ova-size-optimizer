package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"golang.org/x/sync/errgroup"
)

const (
	tempDir                  = "temp-oci"
	multiArchiveExtractedDir = tempDir + "/multi-archive-extracted"
	ociImageIndexFilePath    = multiArchiveExtractedDir + "/index.json"
	individualArchivesDir    = tempDir + "/individual-archives"
)

var (
	CompressedMultiArchive string
	MultiArchive           string
)

type ImageIndex struct {
	SchemaVersion int        `json:"schemaVersion"`
	MediaType     string     `json:"mediaType"`
	Manifests     []Manifest `json:"manifests"`
}

type Manifest struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int               `json:"size"`
	Annotations map[string]string `json:"annotations"`
	Platform    Platform          `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

func main() {
	ovaImagePath := os.Args[1]
	if ovaImagePath == "" {
		fmt.Println("Please provide ova path as an argument")
		os.Exit(1)
	}

	CompressedMultiArchive = path.Base(ovaImagePath)
	MultiArchive = tempDir + "/" + CompressedMultiArchive

	file, err := os.Open(CompressedMultiArchive)
	if err != nil {
		fmt.Printf("Error opening compressed file %s: %v\n", CompressedMultiArchive, err)
		os.Exit(1)
	}
	defer file.Close()

	if err := extractTarGz(file, MultiArchive, multiArchiveExtractedDir); err != nil {
		fmt.Printf("Error extracting archive: %v\n", err)
		os.Exit(1)
	}

	indexContents, err := unmarshallIndex(ociImageIndexFilePath)
	if err != nil {
		fmt.Printf("Error extracting index contents: %v\n", err)
		os.Exit(1)
	}

	err = transformOciToDockerImageFormat(indexContents)
	if err != nil {
		fmt.Printf("Error copying OCI image to Docker image: %v\n", err)
		os.Exit(1)
	}
}

func transformOciToDockerImageFormat(imgIndex ImageIndex) error {
	// spawn a goroutine for each skopeo copy call
	var eg errgroup.Group

	for _, manifest := range imgIndex.Manifests {
		manifest := manifest
		eg.Go(func() error {
			imageName := manifest.Annotations["io.containerd.image.name"]
			imageName = strings.Split(imageName, "/")[2]
			imageName = strings.Split(imageName, ":")[0]

			imageRef := manifest.Annotations["org.opencontainers.image.ref.name"]

			imageArchiveSrc := "oci-archive:" + MultiArchive + ":" + imageRef
			imageArchiveDst := "docker-archive:" + individualArchivesDir + "/" + imageName + "-" + imageRef + ".tar"
			err := imageCopy(imageArchiveSrc, imageArchiveDst)
			return err
		})
	}
	return eg.Wait()
}

func imageCopy(ociImage, dockerImage string) error {
	ctx := context.Background()

	policy, err := signature.DefaultPolicy(nil)
	if err != nil {
		return err
	}
	policyCtx, err := signature.NewPolicyContext(policy)
	if err != nil {
		return fmt.Errorf("error loading default policy: %v", err)
	}
	defer policyCtx.Destroy()

	srcRef, err := alltransports.ParseImageName(ociImage)
	if err != nil {
		return fmt.Errorf("invalid source oci name %s: %w", ociImage, err)
	}
	destRef, err := alltransports.ParseImageName(dockerImage)
	if err != nil {
		return fmt.Errorf("invalid destination docker name %s: %w", dockerImage, err)
	}

	fmt.Println(destRef)
	fmt.Println(srcRef)
	_, err = copy.Image(ctx, policyCtx, destRef, srcRef, &copy.Options{
		SourceCtx:      &types.SystemContext{},
		DestinationCtx: &types.SystemContext{},
	})
	if err != nil {
		return fmt.Errorf("error copying image: %v", err)
	}

	return nil
}
func unmarshallIndex(indexPath string) (ImageIndex, error) {
	var imgIndex ImageIndex

	indexFile, err := os.Open(indexPath)
	if err != nil {
		return imgIndex, fmt.Errorf("unable to open %s: %w", indexPath, err)
	}

	indexJson, err := io.ReadAll(indexFile)
	if err != nil {
		return imgIndex, fmt.Errorf("unable to read %s: %w", indexPath, err)
	}

	err = json.Unmarshal(indexJson, &imgIndex)
	if err != nil {
		return imgIndex, fmt.Errorf("unable to unmarshall: %w", err)
	}

	return imgIndex, nil
}

func decompressGzip(gzipStream io.Reader, tarFilePath string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %v", err)
	}
	defer uncompressedStream.Close()

	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("error removing directory %s: %v", tempDir, err)
	}

	err = os.Mkdir(tempDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating temp dir: %v", err)
	}

	err = os.Mkdir(individualArchivesDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %v", individualArchivesDir, err)
	}

	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return fmt.Errorf("error creating tar file: %v", err)
	}
	defer tarFile.Close()

	_, err = io.Copy(tarFile, uncompressedStream)
	if err != nil {
		return fmt.Errorf("error copying data to tar file: %v", err)
	}

	return nil
}

func extractTar(tarFilePath, dest string) error {
	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		return fmt.Errorf("error opening tar file: %v", err)
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return fmt.Errorf("error reading tar entry: %v", err)
		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", filepath.Dir(target), err)
			}
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("error creating file %s: %v", target, err)
			}
			if _, err := io.Copy(fileToWrite, tarReader); err != nil {
				fileToWrite.Close()
				return fmt.Errorf("error writing file %s: %v", target, err)
			}
			fileToWrite.Close()
		}
	}
}

func extractTarGz(gzipStream io.Reader, tarFilePath, dest string) error {
	err := decompressGzip(gzipStream, tarFilePath)
	if err != nil {
		return err
	}

	return extractTar(tarFilePath, dest)
}
