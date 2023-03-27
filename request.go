package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type json_ struct {
	Temp int `json:"temp"`
}

type json2 struct {
	Fact json_ `json:"fact"`
}

func main() {
	url := fmt.Sprintf("https://api.weather.yandex.ru/v2/forecast?lat=55.60163&lon=37.34665")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("X-Yandex-API-Key", "KEY-VALUE")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	//fmt.Println(string(body))

	var requ json2
	err = json.Unmarshal(body, &requ)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Температура сейчас: %d\n", requ.Fact.Temp)

}
