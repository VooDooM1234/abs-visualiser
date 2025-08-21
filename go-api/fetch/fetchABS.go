package fetch

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/VooDooM1234/abs-visualiser/go-api/db"
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

// https://data.api.abs.gov.au/rest/dataflow/all?detail=allstubs
// Help func to load into database, do not use in live server - takes ages to get response from ABS API
func (f *Fetch) ABSRestDataflowAll(db *db.Database) error {

	type ABSDataflow struct {
		ID                  string `json:"id"`
		Version             string `json:"version"`
		AgencyID            string `json:"agencyID"`
		IsExternalReference bool   `json:"isExternalReference"`
		IsFinal             bool   `json:"isFinal"`
		Name                string `json:"name"`
	}

	type ABSDataflowWrapper struct {
		Data struct {
			Dataflows []ABSDataflow `json:"dataflows"`
		} `json:"data"`
	}

	endPoint := "/rest/dataflow/all"
	path := Path{
		Endpoint: endPoint,
		Params: map[string]string{
			"detail": "allstubs",
		},
	}

	body, err := f.GetJSONHeader(path)
	if err != nil {
		return fmt.Errorf("fetching rest/dataflow/all: %w", err)
	}

	var wrapper ABSDataflowWrapper
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	//write to static file and database
	for _, absDataflow := range wrapper.Data.Dataflows {
		_, err = db.Conn.Exec(db.Ctx,
			`INSERT INTO "abs_static_dataflow" (
			id,
			version,
			agency_id,
			is_external_reference,
			is_final,
			name
		)
         VALUES ($1, $2, $3, $4, $5, $6)
         ON CONFLICT (id, version)
         DO UPDATE SET agency_id = EXCLUDED.agency_id,
		 			is_external_reference = EXCLUDED.is_external_reference,
					is_final = EXCLUDED.is_final,
					name = EXCLUDED.name`,
			absDataflow.ID, absDataflow.Version, absDataflow.AgencyID, absDataflow.IsExternalReference, absDataflow.IsFinal, absDataflow.Name,
		)

		if err != nil {
			return fmt.Errorf("upsert failed: %w", err)
		}
	}
	// log.print(res)

	err = os.WriteFile("../static/data/ABSDataflowAll.json", body, 0777)
	if err != nil {
		return fmt.Errorf("Error writing to JSON %w", err)
	}

	return err
}
