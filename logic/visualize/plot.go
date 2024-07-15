package visualize

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"ova-size-optimizer/logic/analyze"
)

func GenerateReport(archivesStats map[string]*analyze.Stats) error {
	fmt.Println("Started generating report...")

	for archiveName, archiveStats := range archivesStats {
		duplicateBaseOS := analyze.GetOnlyDuplicates(archiveStats.BaseOS)
		duplicatePackages := analyze.GetOnlyDuplicates(archiveStats.Packages)
		duplicateRuntimes := analyze.GetOnlyDuplicatesRuntimes(archiveStats.Runtimes)

		archiveBaseName := filepath.Base(archiveName)
		archiveName := strings.Split(archiveBaseName, ".")[0]
		if err := PlotStats(duplicateBaseOS, "BaseOS Statistics", fmt.Sprintf("%s-stats-base-os.png", archiveName), 10); err != nil {
			return fmt.Errorf("Error generating bar chart for BaseOS: %v\n", err)
		}

		if err := PlotStats(duplicatePackages, "Package Statistics", fmt.Sprintf("%s-stats-packages.png", archiveName), 10); err != nil {
			return fmt.Errorf("Error generating bar chart for packages: %v\n", err)
		}

		if err := PlotStats(duplicateRuntimes, "Runtime Statistics", fmt.Sprintf("%s-stats-runtimes.png", archiveName), 10); err != nil {
			return fmt.Errorf("Error generating bar chart for runtimes: %v\n", err)
		}
	}

	fmt.Println("Finished generating report successfully.")
	return nil
}

func PlotStats(stats map[string]*analyze.Info, title string, filename string, topN int) error {
	if len(stats) == 0 {
		fmt.Fprintf(os.Stderr, "Warn: No duplicates to generate stats for: %s\n", title)
		return nil
	}

	p := plot.New()
	p.Title.Text = title

	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range stats {
		ss = append(ss, kv{k, v.Count})
	}

	sort.Slice(ss, func(i, j int) bool {
		if ss[i].Value == ss[j].Value {
			return ss[i].Key < ss[j].Key
		}
		return ss[i].Value > ss[j].Value
	})

	if topN > len(ss) {
		topN = len(ss)
	}

	ss = ss[:topN]

	bars := make(plotter.Values, topN)
	labels := make([]string, topN)
	for i := 0; i < topN; i++ {
		bars[i] = float64(ss[i].Value)
		labels[i] = ss[i].Key
	}

	bar, err := plotter.NewBarChart(bars, vg.Points(20))
	if err != nil {
		return err
	}
	bar.Horizontal = true

	p.Add(bar)
	p.NominalY(labels...)

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, vg.Length(float64(topN)*0.5)*vg.Inch, filename); err != nil {
		return err
	}

	return nil
}
