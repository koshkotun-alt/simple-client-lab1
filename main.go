package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// Структура для парсинга ответа Nominatim
type Result struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

func fetchCoordinates(address string, results chan<- Result, errs chan<- error) {
	// Экранирование адреса
	escapedAddress := url.QueryEscape(address)
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", escapedAddress)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		errs <- err
		return
	}

	req.Header.Set("User-Agent", "GoGeoClient/1.0 (koshkotun@gmail.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errs <- err
		return
	}
	defer resp.Body.Close()

	var resultsSlice []Result
	if err := json.NewDecoder(resp.Body).Decode(&resultsSlice); err != nil {
		errs <- err
		return
	}

	if len(resultsSlice) == 0 {
		errs <- fmt.Errorf("Нет результатов для адреса: %s", address)
		return
	}

	results <- resultsSlice[0]
}

func main() {
	// Обработка флагов
	// Например, флаг вывода: --format=plain|json
	formatFlag := flag.String("format", "json", "Формат вывода: json или plain")
	flag.Parse()

	addresses := flag.Args()
	if len(addresses) == 0 {
		fmt.Println("Пожалуйста, укажите хотя бы один адрес.")
		os.Exit(1)
	}

	// Каналы для результатов, ошибок и завершения
	resultsCh := make(chan Result)
	errorsCh := make(chan error)
	doneCh := make(chan struct{})

	// Запуск горутин для каждой адреса
	for _, addr := range addresses {
		go fetchCoordinates(addr, resultsCh, errorsCh)
	}

	// Координаты для вывода и счетчик
	expected := len(addresses)
	found := 0

	// Обработка результатов и ошибок
	go func() {
		for {
			select {
			case res := <-resultsCh:
				if *formatFlag == "json" {
					data, _ := json.MarshalIndent(res, "", "  ")
					fmt.Println(string(data))
				} else { // plain
					fmt.Printf("Адрес: %s\nШирота: %s\nДолгота: %s\n---\n", res.DisplayName, res.Lat, res.Lon)
				}
				found++
				if found == expected {
					close(doneCh)
					return
				}
			case err := <-errorsCh:
				fmt.Println("Ошибка:", err)
				found++
				if found == expected {
					close(doneCh)
					return
				}
			}
		}
	}()

	// Ожидаем завершения
	<-doneCh
}
