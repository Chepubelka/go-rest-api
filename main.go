package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type WeatherInfo struct {
	List []WeatherListItem `json:"list"`
  }
  
  type WeatherListItem struct {
	Dt      int           `json:"dt"`
	Main    WeatherMain   `json:"main"`
	Weather []WeatherType `json:"weather"`
  }
  
  type WeatherMain struct {
	Temp      float32 `json:"temp"`
	FeelsLike float32 `json:"feels_like"`
	Humidity  int     `json:"humidity"`
  }
  
  type WeatherType struct {
	Icon string `json:"icon"`
  }

  func getWeather() (*WeatherInfo, error) {
	openWeatherMapApiKey := "e13a95c00d42a26e75968f9b296ab61f"
	var url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=Rostov-On-Don&cnt=4&units=metric&appid=%s", openWeatherMapApiKey)
	response, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var weatherInfo = new(WeatherInfo)

	err = json.Unmarshal(body, &weatherInfo)

	return weatherInfo, err
}

func main() {
	weather, err := getWeather()
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(weather.List); i++ {
		var weatherForDay = weather.List[i]
		fmt.Printf("Температура: %.2f °C ", weatherForDay.Main.Temp)
		fmt.Printf("Ощущается как: %.2f °C ", weatherForDay.Main.FeelsLike)
		fmt.Printf("Влажность: %d%%", weatherForDay.Main.Humidity)
		fmt.Println()
	}
}