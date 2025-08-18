package fetch

import (
	"github.com/gocarina/gocsv"
)

type restDataCSVResponseRaw struct {
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

// Get data from /rest/data endpoint expecting CSV response
func (f *FetchData) ABSRestDataCSV(apiKey, qParam string) ([]*restDataCSVResponseRaw, error) {
	// var data *restDataCSVResponseRaw
	dataArr := []*restDataCSVResponseRaw{}
	//body is returning empty - fix here when no 12am on a monday xdd
	body, err := f.Get("/rest/data", apiKey, qParam)

	if err := gocsv.UnmarshalBytes(body, &dataArr); err != nil {
		panic(err)
	}

	return dataArr, err
}
