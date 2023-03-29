package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type json_ struct {
	Temp  int    `json:"temp"`
	Feels int    `json:"feels_like"`
	Url   string `json:"url"`
}

type json2 struct {
	Fact json_ `json:"fact"`
	Info json_ `json:"info"`
}

func main() {				//Вечный цикл

	for {
		times := time.Now()
		if times.Minute() == 30 || times.Minute() == 00{
		
			request_API()

		}
	}

}

func request_API(){			//Запрос по API

	url := fmt.Sprintf("https://api.weather.yandex.ru/v2/forecast?lat=55.60123&lon=37.3594")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("X-Yandex-API-Key", "baea74df-7a43-4d00-9b5d-823f9be09c62")

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

	var requ json2
	err = json.Unmarshal(body, &requ)
	if err != nil {
		panic(err)
	}

	database_managed(requ.Fact.Temp, requ.Fact.Feels)	

}


func database_managed(Temp int, Feels int){			//Внесение значений в базу данных и чтение из неё


	times := time.Now()
	db, err := sql.Open("sqlite3", "./weather.db")
			if err != nil {
				panic(err)
			}
			defer db.Close()
			// Создание таблицы в базе данных, если она не существует
			_, err = db.Exec("CREATE TABLE IF NOT EXISTS Weather (id INTEGER PRIMARY KEY AUTOINCREMENT, Время TEXT, Температура INTEGER, Ощущается INTEGER)")
			if err != nil {
				log.Fatal("Ошибка создания таблицы в БД: ", err)
			}

			// Запись метрики в базу данных
			stmt, err := db.Prepare("INSERT INTO Weather(Время, Температура, Ощущается) values(?,?,?)")
			if err != nil {
				log.Fatal("Ошибка подготовки запроса к БД: ", err)
			}

			data := fmt.Sprintf("%d:%d\n", times.Hour(), times.Minute())

			_, err = stmt.Exec(data, Temp, Feels)
			if err != nil {
				log.Fatal("Ошибка выполнения запроса к БД: ", err)
			}

			fmt.Printf("Упех\n")
			time.Sleep(1*time.Minute)

}
