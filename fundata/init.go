package fundata

func Init() {
	var key string = ""
	var serc string = ""
	InitClient(key, serc)
	params := map[string]string{}
	Get("/uri", params)
}
