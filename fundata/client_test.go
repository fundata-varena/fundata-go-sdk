package fundata

import "testing"

func Test_Client(t *testing.T) {
	var key = ""
	var secret = ""
	InitClient(key, secret)

	params := make(map[string]interface{})
	params["page"] = 1
	params["limit"] = 10

	res, err := Get("/data-service/dota2/pro/league/ti/rank-player", params)
	if err != nil {
		t.Log("Request error", err)
		return
	}

	t.Log("Response got", res)
}
