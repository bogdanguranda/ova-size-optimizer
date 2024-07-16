package ociimage

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
	tempDir                  = "temp"
	multiArchiveExtractedDir = tempDir + "/multi-archive-extracted"
	IndividualArchivesDir    = tempDir + "/individual-archives"
	ociImageIndexFilePath    = multiArchiveExtractedDir + "/index.json"
	defaultPolicyFilePath    = "/etc/containers/policy.json"
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

type Policy struct {
	Default []PolicyEntry `json:"default"`
}

// PolicyEntry defines a policy entry.
type PolicyEntry struct {
	Type string `json:"type"`
}

func TransformAndCopyOciToDockerImage(ovaImagePath string) error {
	CompressedMultiArchive = path.Base(ovaImagePath)
	MultiArchive = tempDir + "/" + CompressedMultiArchive

	file, err := os.Open(CompressedMultiArchive)
	if err != nil {
		return fmt.Errorf("error opening compressed file %s: %v", CompressedMultiArchive, err)
	}
	defer file.Close()

	if err := extractTarGz(file, MultiArchive, multiArchiveExtractedDir); err != nil {
		return fmt.Errorf("error extracting archive: %v", err)
	}

	indexContents, err := unmarshallIndex(ociImageIndexFilePath)
	if err != nil {
		return fmt.Errorf("error extracting index contents: %v", err)
	}

	err = transformOciToDockerImageFormat(indexContents)
	if err != nil {
		return fmt.Errorf("error copying OCI image to Docker image: %v", err)
	}

	return nil
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
			imageArchiveDst := "docker-archive:" + IndividualArchivesDir + "/" + imageName + "-" + imageRef + ".tar"
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
		return fmt.Errorf("error getting default policy: %v", err)
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
	defer indexFile.Close()

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

	err = os.Mkdir(IndividualArchivesDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory %s: %v", IndividualArchivesDir, err)
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

func checkPolicyAndCreateIfMissing() error {
	if _, err := os.Stat(defaultPolicyFilePath); err == nil {
		fmt.Printf("File %s exists.\n", defaultPolicyFilePath)
		return nil
	} else if os.IsNotExist(err) {
		fmt.Printf("File %s does not exist.\n", defaultPolicyFilePath)
	} else {
		return fmt.Errorf("error checking file %s: %v", defaultPolicyFilePath, err)
	}

	err := os.MkdirAll(filepath.Dir(defaultPolicyFilePath), 0755)
	if err != nil {
		return fmt.Errorf("error creating directories: %v", err)
	}

	policyFile, err := os.Create(defaultPolicyFilePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer policyFile.Close()

	policy := Policy{
		Default: []PolicyEntry{
			{Type: "insecureAcceptAnything"},
		},
	}

	policyJSON, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy to JSON: %v", err)
	}

	_, err = policyFile.Write(policyJSON)
	if err != nil {
		return fmt.Errorf("failed to write policy to %s : %v", defaultPolicyFilePath, err)
	}
	return nil
}
