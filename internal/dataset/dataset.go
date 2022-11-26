package dataset

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/jedib0t/go-pretty/table"
	"github.com/montanaflynn/stats"
)

const (
	rowLen int = 8

	groupTypeMine       groupType = "mine"
	groupTypeBackground groupType = "background"
)

var tableStyle table.Style = table.Style{
	Name:    "StyleRounded",
	Box:     table.StyleBoxRounded,
	Color:   table.ColorOptionsDefault,
	Format:  table.FormatOptionsDefault,
	Options: table.OptionsDefault,
	Title:   table.TitleOptionsDefault,
}

type (
	groupType string

	sample struct {
		group groupType
		code  string
		xUtm  string
		yUtm  string
		pbPpm float64
		asPpm float64
		sbPpm float64
		vPpm  float64
	}

	variances struct {
		pb, as, sb, v float64
	}

	fDistributions struct {
		pb, as, sb, v float64
	}

	dataset struct {
		samples        map[groupType][]sample
		variances      map[groupType]variances
		fDistributions fDistributions
	}
)

func Load(f io.ReadCloser) (dataset, error) {
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("could not close dataset file:", err)
		}
	}()

	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'

	ds := dataset{
		samples:   map[groupType][]sample{},
		variances: map[groupType]variances{},
	}

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		sample, err := newSample(rec)
		if err != nil {
			log.Printf("could not parse data row '%+#v': %s", rec, err)
			continue
		}

		ds.samples[sample.group] = append(ds.samples[sample.group], sample)
	}
	return ds, nil
}

func newSample(input []string) (sample, error) {
	if len(input) != rowLen {
		return sample{}, fmt.Errorf("invalid row length: %d", len(input))
	}

	s := sample{
		code: input[0],
		xUtm: input[1],
		yUtm: input[2],
	}

	inputGroup := input[3]

	switch inputGroup {
	case "mine":
		s.group = groupTypeMine
	case "background":
		s.group = groupTypeBackground
	default:
		return sample{}, fmt.Errorf("invalid group: %s", inputGroup)
	}

	pbPPM, err := strconv.ParseFloat(input[4], 64)
	if err != nil {
		return sample{}, fmt.Errorf("could not parse pb_ppm to float64: %s", err)
	}
	s.pbPpm = pbPPM

	asPPM, err := strconv.ParseFloat(input[5], 64)
	if err != nil {
		return sample{}, fmt.Errorf("could not parse as_ppm to float64: %s", err)
	}
	s.asPpm = asPPM

	sbPPM, err := strconv.ParseFloat(input[6], 64)
	if err != nil {
		return sample{}, fmt.Errorf("could not parse sb_ppm to float64: %s", err)
	}
	s.sbPpm = sbPPM

	vPPM, err := strconv.ParseFloat(input[7], 64)
	if err != nil {
		return sample{}, fmt.Errorf("could not parse v_ppm to float64: %s", err)
	}
	s.vPpm = vPPM

	return s, nil
}

func (ds *dataset) CalculateVariances() error {
	var pbMine, pbBg, asMine, asBg, sbMine, sbBg, vMine, vBg stats.Float64Data

	for k, groupSamples := range ds.samples {
		for _, sample := range groupSamples {
			switch k {
			case "mine":
				pbMine = append(pbMine, sample.pbPpm)
				asMine = append(asMine, sample.asPpm)
				sbMine = append(sbMine, sample.sbPpm)
				vMine = append(vMine, sample.vPpm)
			case "background":
				pbBg = append(pbBg, sample.pbPpm)
				asBg = append(asBg, sample.asPpm)
				sbBg = append(sbBg, sample.sbPpm)
				vBg = append(vBg, sample.vPpm)
			}
		}
	}

	mineVariances, err := calcGroupVariances(pbMine, asMine, sbMine, vMine)
	if err != nil {
		return fmt.Errorf("could not calculate group variances for group 'mine': %s", err)
	}

	ds.variances[groupTypeMine] = mineVariances

	bgVariances, err := calcGroupVariances(pbBg, asBg, sbBg, vBg)
	if err != nil {
		return fmt.Errorf("could not calculate group variances for group 'background': %s", err)
	}

	ds.variances[groupTypeBackground] = bgVariances

	return nil
}

func calcGroupVariances(pb, as, sb, v stats.Float64Data) (variances, error) {
	pbVari, err := stats.Variance(pb)
	if err != nil {
		return variances{}, fmt.Errorf("could not calculate variance for pb_ppm: %s", err)
	}

	asVari, err := stats.Variance(as)
	if err != nil {
		return variances{}, fmt.Errorf("could not calculate variance for as_ppm: %s", err)
	}

	sbVari, err := stats.Variance(sb)
	if err != nil {
		return variances{}, fmt.Errorf("could not calculate variance for sb_ppm: %s", err)
	}

	vVari, err := stats.Variance(v)
	if err != nil {
		return variances{}, fmt.Errorf("could not calculate variance for v_ppm: %s", err)
	}

	return variances{
		pb: pbVari,
		as: asVari,
		sb: sbVari,
		v:  vVari,
	}, nil
}

func (ds *dataset) CalculateFDistributions() {
	ds.fDistributions.pb = ds.variances[groupTypeMine].pb / ds.variances[groupTypeBackground].pb
	ds.fDistributions.as = ds.variances[groupTypeMine].as / ds.variances[groupTypeBackground].as
	ds.fDistributions.sb = ds.variances[groupTypeMine].sb / ds.variances[groupTypeBackground].sb
	ds.fDistributions.v = ds.variances[groupTypeMine].v / ds.variances[groupTypeBackground].v
}

func (ds *dataset) PrintDataset() {
	tw := table.NewWriter()
	tw.SetStyle(tableStyle)
	tw.SetTitle("DATASET")
	tw.AppendHeader(table.Row{"group", "sample", "x_utm", "y_utm", "pb_ppm", "as_ppm", "sb_ppm", "v_ppm"})

	for _, samples := range ds.samples {
		for _, sample := range samples {
			tw.AppendRow(table.Row{
				sample.group,
				sample.code,
				sample.xUtm,
				sample.yUtm,
				fmt.Sprintf("%.2f", sample.pbPpm),
				fmt.Sprintf("%.2f", sample.asPpm),
				fmt.Sprintf("%.2f", sample.sbPpm),
				fmt.Sprintf("%.2f", sample.vPpm),
			})
		}
	}

	fmt.Println(tw.Render())
}

func (ds *dataset) PrintVariances() {
	tw := table.NewWriter()
	tw.SetStyle(tableStyle)
	tw.SetTitle("VARIANCE")
	tw.AppendHeader(table.Row{"group", "pb", "as", "sb", "v"})

	for group, vari := range ds.variances {
		tw.AppendRow(table.Row{
			group,
			fmt.Sprintf("%.3f", vari.pb),
			fmt.Sprintf("%.3f", vari.as),
			fmt.Sprintf("%.3f", vari.sb),
			fmt.Sprintf("%.3f", vari.v),
		})
	}
	fmt.Println(tw.Render())
}

func (ds *dataset) PrintFDistributions() {
	tw := table.NewWriter()
	tw.SetStyle(tableStyle)
	tw.SetTitle("F-DISTRIBUTION")
	tw.AppendHeader(table.Row{"pb", "as", "sb", "v"})

	tw.AppendRow(table.Row{
		fmt.Sprintf("%.3f", ds.fDistributions.pb),
		fmt.Sprintf("%.3f", ds.fDistributions.as),
		fmt.Sprintf("%.3f", ds.fDistributions.sb),
		fmt.Sprintf("%.3f", ds.fDistributions.v),
	})
	fmt.Println(tw.Render())
}
