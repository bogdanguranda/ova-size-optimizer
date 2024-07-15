package analyze

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/sbom"
)

func Analyze(individualArchivesDir string) (map[string]*Stats, error) {
	fmt.Println("Started analyzing individual archives...")

	ctx := context.Background()

	archivesWithSboms, err := analyzeArchives(ctx, individualArchivesDir)
	if err != nil {
		return nil, fmt.Errorf("error analyzing individual archives: %w", err)
	}

	archivesStats := map[string]*Stats{}
	for archiveName, archiveSbom := range archivesWithSboms {
		archiveStats, err := AggregateData(individualArchivesDir, archiveName, archiveSbom)
		if err != nil {
			return nil, fmt.Errorf("error aggregating data for individual archive %s: %w", archiveName, err)
		}

		archivesStats[archiveName] = archiveStats
	}

	fmt.Println("Finished analyzing individual successfully.")
	return archivesStats, nil
}

func analyzeArchives(ctx context.Context, individualArchivesDir string) (map[string]*sbom.SBOM, error) {
	archiveWithSbom := map[string]*sbom.SBOM{}
	pattern := filepath.Join(individualArchivesDir, "*.tar")

	archives, err := filepath.Glob(pattern)
	if err != nil {
		return archiveWithSbom, fmt.Errorf("error finding tar files: %w", err)
	}

	for _, archive := range archives {
		fmt.Println("Processing individual archive:", archive)

		archiveSbom, err := generateSbom(ctx, individualArchivesDir+"/"+archive)
		if err != nil {
			return archiveWithSbom, fmt.Errorf("error generating SBOM for archive %s: %w", archive, err)
		}
		archiveWithSbom[archive] = archiveSbom
	}

	return archiveWithSbom, nil
}

func generateSbom(ctx context.Context, individualArchivePath string) (*sbom.SBOM, error) {
	src, err := syft.GetSource(ctx, individualArchivePath, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting docker archive: %w", err)
	}

	sbom, err := syft.CreateSBOM(ctx, src, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating SBOM for docker archive: %w", err)
	}

	return sbom, nil
}
