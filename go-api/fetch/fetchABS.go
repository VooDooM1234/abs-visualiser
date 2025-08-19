package fetch

import (
	"fmt"

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

// Get data from /rest/data endpoint expecting CSV response, ABS get data from /rest/data/ but then has extra information for its dataflowidenitifer and datakey
func (f *FetchData) ABSRestDataCSV(apiKey, dataflowIdentifier, dataKey, qParam string) ([]*restDataCSVResponseRaw, error) {
	endPoint := fmt.Sprintf("/rest/data/%s/%s", dataflowIdentifier, dataKey)
	path := Path{
		Endpoint: endPoint,
		Params: map[string]string{
			// "q":      qParam,
			// "key":    apiKey,
			// "startPeriod": "",
			// "endPeriod":   "",
			"format": "csvfilewithlabels",
			"detail": "dataonly",
		},
	}

	dataArr := []*restDataCSVResponseRaw{}
	body, err := f.Get(path)

	if err := gocsv.UnmarshalBytes(body, &dataArr); err != nil {
		panic(err)
	}

	return dataArr, err
}
