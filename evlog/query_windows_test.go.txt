package evlog

import (
	"testing"
)

const testXML = `<QueryList><Query Id="0" Path="Application"><Select Path="Application">*[System[TimeCreated[timediff(@SystemTime) &lt;= 3600000]]]</Select><Select Path="System">*[System[TimeCreated[timediff(@SystemTime) &lt;= 3600000]]]</Select></Query></QueryList>`
const testXML2 = `<QueryList><Query Id="0" Path="Application"><Select Path="Application">*</Select><Select Path="System">*</Select></Query></QueryList>`

func TestQueryStringFromChannels(t *testing.T) {
	qstr, err := queryStringFromChannels([]string{"Application", "System"}, 3600)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(qstr)
	if qstr != testXML {
		t.Fatal("expected xml:", testXML)
	}
}

func TestQueryStringFromChannels2(t *testing.T) {
	qstr, err := queryStringFromChannels([]string{"Application", "System"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(qstr)
	if qstr != testXML2 {
		t.Fatal("expected xml:", testXML2)
	}
}
