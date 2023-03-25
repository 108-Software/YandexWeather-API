package main

import (
	"context"
	"fmt"
	"log"
	"time"

	weather "github.com/3crabs/go-yandex-weather-api/wheather"
)

func main() {
	yandexWeatherApiKey := "baea74df-7a43-4d00-9b5d-823f9be09c62"
	w, err := weather.GetWeatherWithCache(context.TODO(), yandexWeatherApiKey, 55.60163, 37.34665, time.Hour)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(w)
	fmt.Printf("Сегодня %s\n", w.Fact.GetCondition())
	fmt.Printf("Температура %d°C\n", w.Fact.Temp)
	fmt.Printf("Ощущается как %d°C\n", w.Fact.FeelsLike)
	fmt.Printf("Порывы ветра до %.1f м/с\n", w.Fact.WindGust)
}
