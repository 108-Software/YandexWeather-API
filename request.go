package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
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

func main() {
	// 3
	s := gocron.NewScheduler(time.Now().Location())

	// 4
	s.Every(1).Minutes().Do(func() {
		err := cycle()
		if err != nil {
			log.Println("Не был запущен цикл проверен времени", err)
		}
	})

	// 5
	s.StartBlocking()

}

func cycle() error { //Вечный цикл

	times := time.Now()
	if times.Minute() == 30 || times.Minute() == 00 {
		err := check_day()
		if err != nil {
			return err
		}

	}
	return nil
}

func check_day() error {

	times := time.Now()
	now_date := fmt.Sprintf("%s", times.Format("01-02-2006"))

	_, err := os.Stat("curr_data.txt")
	if err != nil {
		if os.IsNotExist(err) {
			//fmt.Println("file does not exist") // это_true
			file, err := os.Create("curr_data.txt")

			if err != nil {
				log.Print("Ошибка создания файла: ")
				return err
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
			log.Print("Ошибка открытия файла: ")
			return err
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
			err = request_API()

			if err != nil {
				log.Println("Ошибка в API запросе", err)
			}
		} else { // иначе вызываем функцию создания графа и перезаписываем файл
			fmt.Println("false")

			//вызвать функцию создания графа
			read_database(data_file)

			file, err := os.Create("curr_data.txt")

			if err != nil {
				log.Print("Ошибка создания нового файла: ")
				return err
			}
			defer file.Close()
			file.WriteString(now_date)
			err = request_API()

			if err != nil {
				log.Println("Ошибка в API запросе", err)
			}
		}
	}
	return nil
}

func request_API() error { //Запрос по API

	url := fmt.Sprintf("https://api.weather.yandex.ru/v2/forecast?lat=55.60123&lon=37.3594")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print("Ошибка создание requet запроса : ")
		return err
	}

	req.Header.Set("X-Yandex-API-Key", "baea74df-7a43-4d00-9b5d-823f9be09c62")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Ошибка request запроса: ")
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	var requ json2
	err = json.Unmarshal(body, &requ)
	if err != nil {
		log.Print("Ошибка десериализации данных JSON: ")
		return err
	}

	err = inser_database(requ.Fact.Temp, requ.Fact.Feels)
	if err != nil {
		log.Print("Ошибка записи в бд")
		return err
	}

	return nil

}

func inser_database(Temp int, Feels int) error { //Внесение значений в базу данных и чтение из неё

	times := time.Now()
	namedb := fmt.Sprintf("./Weather/%s.db", times.Format("01-02-2006"))

	db, err := sql.Open("sqlite3", namedb)
	if err != nil {
		log.Print("Ошибка открытия базы данных: ")
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Weather (id INTEGER PRIMARY KEY AUTOINCREMENT, Время TEXT, Температура INTEGER, Ощущается INTEGER)")
	if err != nil {
		log.Fatal("Ошибка создания таблицы в БД: ")
		return err
	}

	stmt, err := db.Prepare("INSERT INTO Weather(Время, Температура, Ощущается) values(?,?,?)")
	if err != nil {
		log.Fatal("Ошибка подготовки запроса к БД: ")
		return err
	}

	data := time.Now().Format("15:04")

	_, err = stmt.Exec(data, Temp, Feels)
	if err != nil {
		log.Fatal("Ошибка выполнения запроса к БД: ")
		return err
	}

	defer db.Close()

	return nil

}

func read_database(namedb string) error {
	database := fmt.Sprintf("./Weather/%s.db", namedb)
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		log.Print("Ошибка открытия базы данных: ")
		return err
	}

	defer db.Close()

	rows, err := db.Query("select * from Weather")
	if err != nil {
		log.Print("Ошибка запроса к  базе данных: ")
		return err
	}
	defer rows.Close()

	gg := []weather_data{}

	for rows.Next() {
		p := weather_data{}
		err := rows.Scan(&p.id, &p.Время, &p.Температура, &p.Ощущается)
		if err != nil {
			log.Print("Ошибка чтения из базы данных: ")
			return err
		}
		gg = append(gg, p)
	}

	err = graf(namedb, gg)
	if err != nil {
		log.Print("Ошибка создания графа: ")
		return err
	}
	return nil
}

func graf(namedb string, data []weather_data) error {
	p := plot.New()

	p.Title.Text = namedb
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
		log.Print("Ошибка задания линии 1 для графа: ")
		return err
	}
	line.LineStyle.Width = vg.Points(2)
	line.Color = plotutil.Color(0)

	//Создание второй линии для графика
	line2, err := plotter.NewLine(pts2)
	if err != nil {
		log.Print("Ошибка задания линии 2 для графа: ")
		return err
	}
	line2.LineStyle.Width = vg.Points(2)
	line2.Color = plotutil.Color(1)

	// Добавление линии на график
	p.Add(line, line2)

	point1, err := plotter.NewScatter(pts)
	if err != nil {
		log.Print("Ошибка обрисовка линии 1 на графе: ")
		return err
	}

	point2, err := plotter.NewScatter(pts2)
	if err != nil {
		log.Print("Ошибка обрисовка линии 2 на графе: ")
		return err
	}

	// устанавливаем точки на красный цвет
	point1.GlyphStyle.Color = plotutil.Color(0)
	point2.GlyphStyle.Color = plotutil.Color(0)

	p.Add(point1, point2)

	// Устанавливаем метки на оси X
	p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}

	grafs_weather := fmt.Sprintf("grafs_weather/%s.png", namedb)

	// Сохранение графика в файл
	if err := p.Save(10*vg.Inch, 10*vg.Inch, grafs_weather); err != nil {
		log.Print("Ошибка сохранения графа: ")
		return err
	}

	err = send_graf(grafs_weather, namedb)
	if err != nil {
		log.Println("Ошибка отправки графа")
		return err
	}

	return nil
}

func send_graf(grafs_weather string, namedb string) error {
	file, err := os.Open(grafs_weather)
	if err != nil {
		log.Print("Ошибка открытия графа: ")
		return err
	}
	defer file.Close()

	// создаем новый буфер для формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// добавляем файл в форму
	image_name := fmt.Sprintf("%s.png", namedb)
	part, err := writer.CreateFormFile("image", image_name)
	if err != nil {
		log.Print("Ошибка добавления графа в форму: ")
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Print("Ошибка копирования графа: ")
		return err
	}
	// закрываем форму
	writer.Close()

	// создаем новый POST-запрос на API imgbb.com
	url := fmt.Sprintf("https://api.imgbb.com/1/upload?&key=%s", "a6727c8c01cab9bea1aeaf867b7588ab")
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Print("Ошибка запроса к сервису: ")
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// отправляем запрос и получаем ответ
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Ошибка отправки графа: ")
		return err
	}
	defer resp.Body.Close()

	return nil

}
