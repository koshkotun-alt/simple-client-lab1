package main

import (
    "encoding/json"
    "fmt"
    "net/http"
	"strconv" // добавляем этот импорт
	"net/url"
)

// Структура для парсинга ответа Nominatim
type Result struct {
    Lat string `json:"lat"`
    Lon string `json:"lon"`
    DisplayName string `json:"display_name"`
}

func fetchCoordinates(city string) (float64, float64, error) {
    escapedCity := url.QueryEscape(city)
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", escapedCity)
    client := &http.Client{}

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return 0, 0, err
    }

    // Указываем User-Agent или email для соблюдения правил
	req.Header.Set("User-Agent", "МойГеокодер/1.0 (koshkotun@gmail.com)")

    resp, err := client.Do(req)
    if err != nil {
        return 0, 0, err
    }
    defer resp.Body.Close()

    var results []Result
    if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
        return 0, 0, err
    }

    if len(results) == 0 {
        return 0, 0, fmt.Errorf("Результаты не найдены")
    }

    lat, err := parseFloat(results[0].Lat)
    if err != nil {
        return 0, 0, err
    }
    lon, err := parseFloat(results[0].Lon)
    if err != nil {
        return 0, 0, err
    }
    return lat, lon, nil
}

func parseFloat(s string) (float64, error) {
    return strconv.ParseFloat(s, 64)
}

func main() {
    city := "Москва"

    lat, lon, err := fetchCoordinates(city)
    if err != nil {
        fmt.Println("Ошибка:", err)
        return
    }

    fmt.Printf("Координаты для %s:\nШирота: %f\nДолгота: %f\n", city, lat, lon)
}