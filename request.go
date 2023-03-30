package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"os"
	"io"

	_ "github.com/mattn/go-sqlite3"
)

type json_ struct {
	Temp  int    `json:"temp"`
	Feels int    `json:"feels_like"`
	Url   string `json:"url"`
}

type weather_data struct{
	id int
	Время string
	Температура int
	Ощущается int
}

type json2 struct {
	Fact json_ `json:"fact"`
	Info json_ `json:"info"`
}

func main() {				//Вечный цикл

	for {
		times := time.Now()
		if times.Minute() == 30 || times.Minute() == 00 || times.Minute() == 44 {
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

	//check_day()
	database_managed(requ.Fact.Temp, requ.Fact.Feels)

}

func check_day() {

	times := time.Now()
	now_date := fmt.Sprintf("%s", times.Format("01-02-2006"))

	
	_, err := os.Stat("curr_data.txt")
	if err != nil {
		if os.IsNotExist(err) {
			//fmt.Println("file does not exist") // это_true
			file, err := os.Create("curr_data.txt")

			if err != nil {
				fmt.Println("Unable to create file:", err)
				os.Exit(1)
			}

			defer file.Close()
			file.WriteString(now_date)

		} else {
			// другая ошибка  - это_false
		}
	} else {
		// тут файл существует

		file, err := os.Open("curr_data.txt")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		data := make([]byte, 64)

		var data_file string

		for {
			n, err := file.Read(data)
			if err == io.EOF { // если конец файла
				break // выходим из цикла
			}
			data_file = string(data[:n])

		}
		if data_file == now_date {			// если содержимое файла совпадает с нынешней датой то просто делаем запись в бд
			fmt.Println("True")
		} else {							// иначе вызываем функцию создания графа и перезаписываем файл
			fmt.Println("false")	
		}
		
	}
}

func database_managed(Temp int, Feels int){			//Внесение значений в базу данных и чтение из неё


	times := time.Now()
	namedb := fmt.Sprintf("./Weather/%s.db", times.Format("01-02-2006"))

	db, err := sql.Open("sqlite3", namedb)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Weather (id INTEGER PRIMARY KEY AUTOINCREMENT, Время TEXT, Температура INTEGER, Ощущается INTEGER)")
	if err != nil {
		log.Fatal("Ошибка создания таблицы в БД: ", err)
	}

	stmt, err := db.Prepare("INSERT INTO Weather(Время, Температура, Ощущается) values(?,?,?)")
	if err != nil {
		log.Fatal("Ошибка подготовки запроса к БД: ", err)
	}

	data := fmt.Sprintf("%d:%d", times.Hour(), times.Minute())

	_, err = stmt.Exec(data, Temp, Feels)
	if err != nil {
		log.Fatal("Ошибка выполнения запроса к БД: ", err)
		}

	//fmt.Printf("Упех\n")

	defer db.Close()

	rows, err := db.Query("select * from Weather")
    if err != nil {
        panic(err)
    }
	defer rows.Close()
    gg := []weather_data{}

	for rows.Next(){
        p := weather_data{}
        err := rows.Scan(&p.id, &p.Время, &p.Температура, &p.Ощущается)
        if err != nil{
            fmt.Println(err)
            continue
        }
        gg = append(gg, p)
    }

	for _, p := range gg{
        fmt.Printf("%d  %s  %d  %d", p.id, p.Время, p.Температура, p.Ощущается)
		fmt.Println()
		
    }
	time.Sleep(1*time.Minute)
}
