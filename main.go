package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
	"github.com/gorilla/mux"
	"time"
	"github.com/Chepubelka/go-rest-api/Authorization/auth"
)

type WeatherInfo struct {
	List []WeatherListItem `json:"list"`
  }
  
  type WeatherListItem struct {
	Time_weather	time.Time
	Main    WeatherMain   `json:"main"`
  }
  
  type WeatherMain struct {	
	Temp      	float32 `json:"temp"`
	FeelsLike 	float32 `json:"feels_like"`
	Humidity  	int     `json:"humidity"`
  }

  func getWeather(city string) (*WeatherInfo, error) {
	openWeatherMapApiKey := "e13a95c00d42a26e75968f9b296ab61f"
	var url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q="+ city +"&cnt=3&units=metric&appid=%s", openWeatherMapApiKey)
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

func returnWeather(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	city := vars["city"]
	weather, err := getWeather(city)
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(weather.List); i++ {
		weather.List[i].Time_weather = time.Now().AddDate(0, 0, i)
	}
	json.NewEncoder(w).Encode(weather.List)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/weather/{city}", returnWeather)
	myRouter.HandleFunc("/v1/auth/token", createToken)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	handleRequests()
}