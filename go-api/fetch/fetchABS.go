package fetch

import (
	"fmt"
	"strconv"

	"github.com/gocarina/gocsv"
)

type Observation struct {
	Period  string
	Value   float64
	Region  string
	Measure string
}

func (f *Fetch) ABSRestDataCSV(dataflowIdentifier, dataKey string) ([]Observation, error) {

	type csvRow struct {
		Structure      string `csv:"STRUCTURE"`
		StructureID    string `csv:"STRUCTURE_ID"`
		StructureName  string `csv:"STRUCTURE_NAME"`
		Action         string `csv:"ACTION"`
		MeasureID      string `csv:"MEASURE"`
		MeasureName    string `csv:"Measure"`
		IndexID        string `csv:"INDEX"`
		IndexName      string `csv:"Index"`
		TSEST          string `csv:"TSEST"`
		AdjustmentType string `csv:"Adjustment Type"`
		RegionID       string `csv:"REGION"`
		RegionName     string `csv:"Region"`
		FreqID         string `csv:"FREQ"`
		FreqName       string `csv:"Frequency"`
		TimePeriodID   string `csv:"TIME_PERIOD"`
		TimePeriodName string `csv:"Time Period"`
		ObsValue       string `csv:"OBS_VALUE"`
		Observation    string `csv:"Observation Value"`
	}

	endPoint := fmt.Sprintf("/rest/data/%s/%s", dataflowIdentifier, dataKey)
	path := Path{
		Endpoint: endPoint,
		Params: map[string]string{
			"format": "csvfilewithlabels",
			"detail": "dataonly",
		},
	}

	body, err := f.Get(path)
	if err != nil {
		return nil, fmt.Errorf("fetching ABS CSV: %w", err)
	}

	var rawRows []*csvRow
	if err := gocsv.UnmarshalBytes(body, &rawRows); err != nil {
		return nil, fmt.Errorf("parsing ABS CSV: %w", err)
	}

	// map to clean domain type
	observations := make([]Observation, 0, len(rawRows))
	for _, r := range rawRows {
		val, _ := strconv.ParseFloat(r.ObsValue, 64)
		observations = append(observations, Observation{
			Period:  r.TimePeriodID,
			Value:   val,
			Region:  r.RegionName,
			Measure: r.MeasureName,
		})
	}

	return observations, nil
}
