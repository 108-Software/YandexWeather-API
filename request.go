package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type json_ struct {
	Temp  int    `json:"temp"`
	Feels int    `json:"feels_like"`
	Url   string `json:"url"`
}

type weather_data struct {
	id          int
	Время       string
	Температура int
	Ощущается   int
}

type json2 struct {
	Fact json_ `json:"fact"`
	Info json_ `json:"info"`
}

func main() { //Вечный цикл

	for {
		times := time.Now()
		if times.Minute() == 30 || times.Minute() == 00 {
			check_day()

		}
	}

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
		if data_file == now_date { // если содержимое файла совпадает с нынешней датой то просто делаем запись в бд
			fmt.Println("True")
			request_API()
		} else { // иначе вызываем функцию создания графа и перезаписываем файл
			fmt.Println("false")

			//вызвать функцию создания графа
			read_database(data_file)

			file, err := os.Create("curr_data.txt")

			if err != nil {
				fmt.Println("Unable to create file:", err)
				os.Exit(1)
			}
			defer file.Close()
			file.WriteString(now_date)

		}
	}
}

func request_API() { //Запрос по API

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

	inser_database(requ.Fact.Temp, requ.Fact.Feels)

}

func inser_database(Temp int, Feels int) { //Внесение значений в базу данных и чтение из неё

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

	data := time.Now().Format("15:04")

	_, err = stmt.Exec(data, Temp, Feels)
	if err != nil {
		log.Fatal("Ошибка выполнения запроса к БД: ", err)
	}

	defer db.Close()
	time.Sleep(1 * time.Minute)
}

func read_database(namedb string) {
	database := fmt.Sprintf("/home/flisthdo/Рабочий стол/request/Weather/%s.db", namedb)
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	rows, err := db.Query("select * from Weather")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	gg := []weather_data{}

	for rows.Next() {
		p := weather_data{}
		err := rows.Scan(&p.id, &p.Время, &p.Температура, &p.Ощущается)
		if err != nil {
			panic(err)
		}
		gg = append(gg, p)
	}
	graf(namedb, gg)

}

func graf(namedb string, data []weather_data) {
	p := plot.New()

	p.Title.Text = "Weather"
	p.X.Label.Text = "Время"
	p.Y.Label.Text = "Температура"

	var times []time.Time //объявляем переменные для построения графиков
	for _, d := range data {
		t, _ := time.Parse("15:04", d.Время)
		times = append(times, t)
	}

	var temp []int
	for _, Currate := range data {
		Cur := Currate.Температура
		temp = append(temp, Cur)
	}

	var feels []int
	for _, d := range data {
		Cur := d.Ощущается
		feels = append(feels, Cur)
	}

	pts := make(plotter.XYs, len(temp), len(times))
	for i := range pts {
		pts[i].Y = float64(temp[i])
		pts[i].X = float64(times[i].Unix())

	}

	pts2 := make(plotter.XYs, len(feels), len(times))
	for i := range pts2 {
		pts2[i].Y = float64(feels[i])
		pts2[i].X = float64(times[i].Unix())

	}

	// Создание линии для графика
	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	line.LineStyle.Width = vg.Points(2)
	line.Color = plotutil.Color(0)

	//Создание второй линии для графика
	line2, err := plotter.NewLine(pts2)
	if err != nil {
		panic(err)
	}
	line2.LineStyle.Width = vg.Points(2)
	line2.Color = plotutil.Color(1)

	// Добавление линии на график
	p.Add(line, line2)

	point1, err := plotter.NewScatter(pts)
	if err != nil {
		panic(err)
	}

	point2, err := plotter.NewScatter(pts2)
	if err != nil {
		panic(err)
	}

	// устанавливаем точки на красный цвет
	point1.GlyphStyle.Color = plotutil.Color(0)
	point2.GlyphStyle.Color = plotutil.Color(0)

	p.Add(point1, point2)

	// Устанавливаем метки на оси X
	p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}

	grafs_weather := fmt.Sprintf("/grafs_weather/%s.png", namedb)

	// Сохранение графика в файл
	if err := p.Save(10*vg.Inch, 10*vg.Inch, grafs_weather); err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Minute)
}
