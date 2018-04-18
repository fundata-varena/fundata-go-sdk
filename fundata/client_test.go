package fundata

import "testing"

func Test_Client(t *testing.T) {
	var key string = ""
	var serc string = ""
	InitClient(key, serc)
	params := map[string]string{}
	Get("/uri", params)
}
