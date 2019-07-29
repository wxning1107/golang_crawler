package parser

import (
	"testing"
	"io/ioutil"
)
func TestParseCityList(t *testing.T) {
	contents, err := ioutil.ReadFile("citylist_test_data.html")
	if err != nil {
		panic(err)
	}
	result := ParseCityList(contents)
	const resultSize = 470
	expectedURLs := []string{"http://www.zhenai.com/zhenghun/aba", "http://www.zhenai.com/zhenghun/akesu", "http://www.zhenai.com/zhenghun/alashanmeng"}
	expectedCities := []string{"City 阿坝", "City 阿克苏", "City 阿拉善盟"}
	if len(result.Requests) != resultSize {
		t.Errorf("result should have %d "+"results; but had %d", resultSize, len(result.Requests))
	}
	for i, URL := range expectedURLs {
		if result.Requests[i].Url != URL {
			t.Errorf("expected url #%d: %s; but" + "was %s", i, URL,result.Requests[i].Url)
		}
	}
	if len(result.Items) != resultSize {
		t.Errorf("result should have %d "+"results; but had %d", resultSize, len(result.Items))
	}
	for i, city:= range expectedCities {
		if result.Items[i].(string) != city {
			t.Errorf("expected city #%d: %s; but" + "was %s", i, city,result.Items[i].(string))
		}
	}
}