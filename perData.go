package statuscake

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
)

// PerfData represents a statuscake Performance Data
type TestPerfData struct {
	Time        int    `json:"Time" querystring:"Time" querystringoptions:"omitempty"`
	Status      int    `json:"Status" querystring:"Status" querystringoptions:"omitempty"`
	Location    string `json:"Location" querystring:"Location" querystringoptions:"omitempty"`
	Performance int    `json:"Performance" querystring:"Performance" querystringoptions:"omitempty"`
}

type TestLocationData struct {
	Status      int `json:"Status"`
	Performance int `json:"Performance"`
	Time        int `json:"Time"`
}
type TestsLocation map[string]TestLocationData

type TestsPerfData []TestPerfData

type perfDataBody struct {
	Body map[string]TestPerfData
}

type perfData struct {
	client apiClient
}

type PerfData interface {
	AllWithFilter(filterOptions url.Values) ([]TestPerfData, error)
}

func newPerfClient(c apiClient) PerfData {
	return &perfData{
		client: c,
	}
}

func (pd *perfData) AllWithFilter(filterOptions url.Values) ([]TestPerfData, error) {

	resp, err := pd.client.get("/Tests/Checks", filterOptions)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respJSON := perfDataBody{}

	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, err2
	}

	err3 := json.Unmarshal([]byte(body), &respJSON.Body)
	if err3 != nil {
		return nil, err3
	}

	var testsData []TestPerfData

	// dedup data
	// TODO: refact urgent =(
	metricsLocation := make(map[string]TestsLocation)

	for m := range respJSON.Body {
		l := respJSON.Body[m].Location
		lm, ok := metricsLocation[l]
		if !ok {
			newlm := make(map[string]TestLocationData)
			newlm[l] = TestLocationData{
				Status:      respJSON.Body[m].Status,
				Time:        respJSON.Body[m].Time,
				Performance: respJSON.Body[m].Performance,
			}
			metricsLocation[l] = newlm
		} else {
			if lm[l].Time <= respJSON.Body[m].Time {
				delete(metricsLocation, l)
				newlm := make(map[string]TestLocationData)
				newlm[l] = TestLocationData{
					Status:      respJSON.Body[m].Status,
					Time:        respJSON.Body[m].Time,
					Performance: respJSON.Body[m].Performance,
				}
				metricsLocation[l] = newlm
			}
		}
	}
	for l := range metricsLocation {
		t := TestPerfData{
			Location:    l,
			Performance: metricsLocation[l][l].Performance,
			Time:        metricsLocation[l][l].Time,
			Status:      metricsLocation[l][l].Status,
		}
		testsData = append(testsData, t)
	}

	return testsData, err
}