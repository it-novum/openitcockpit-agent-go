package evlog

import (
	"encoding/xml"
	"fmt"
)

//<Select Path="Application">*[System[TimeCreated[timediff(@SystemTime) &lt;= 3600000]]]</Select>
type xmlEventQuerySelect struct {
	Channel string `xml:"Path,attr"`
	Query   string `xml:",chardata"`
}

type xmlEventQuery struct {
	ID      int                   `xml:"Id,attr"`
	Channel string                `xml:"Path,attr"`
	Select  []xmlEventQuerySelect `xml:"Select"`
}

type xmlEventList struct {
	XMLName xml.Name      `xml:"QueryList"`
	Query   xmlEventQuery `xml:"Query"`
}

func queryStringFromChannels(channels []string, timeDiffSeconds uint64) (string, error) {
	if len(channels) < 1 {
		return "", fmt.Errorf("missing channel for event log query")
	}

	var selects []xmlEventQuerySelect
	for _, channel := range channels {
		query := "*"
		if timeDiffSeconds > 0 {
			query += fmt.Sprint("[System[TimeCreated[timediff(@SystemTime) <= ", timeDiffSeconds*1000, "]]]")
		}
		selects = append(selects, xmlEventQuerySelect{
			Channel: channel,
			Query:   query,
		})
	}

	data, err := xml.Marshal(&xmlEventList{
		Query: xmlEventQuery{
			ID:      0,
			Channel: channels[0],
			Select:  selects,
		},
	})
	if err != nil {
		return "", err
	}
	return string(data), nil
}
