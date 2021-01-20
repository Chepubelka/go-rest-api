package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"database/sql"
	"github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

const (
	APP_KEY = "chepubelka"
)

type WeatherInfo struct {
	List []WeatherListItem `json:"list"`
}

type WeatherListItem struct {
	Time_weather time.Time
	Main         WeatherMain `json:"main"`
}

type WeatherMain struct {
	Temp      float32 `json:"temp"`
	FeelsLike float32 `json:"feels_like"`
	Humidity  int     `json:"humidity"`
}

type User struct {
	id 			int
	email 		string
	password 	string
}

var (
	ctx context.Context
	db  *sql.DB
)

func getWeather(city string) (*WeatherInfo, error) {
	openWeatherMapApiKey := "e13a95c00d42a26e75968f9b296ab61f"
	var url = fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q="+city+"&cnt=3&units=metric&appid=%s", openWeatherMapApiKey)
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
	myRouter.HandleFunc("/token", TokenHandler)
	myRouter.Handle("/weather/{city}", AuthMiddleware(http.HandlerFunc(returnWeather)))
	log.Fatal(http.ListenAndServe(":9999", myRouter))
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")
	r.ParseForm()
    db, err := sql.Open("mysql", "root:XGalHeg7.@/golang")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	var user User
	err = db.QueryRow("select id, email, password from users where email = ?", email).Scan(&user.id, &user.email, &user.password)
	if err != nil {
		panic(err.Error())
	}
	if password != user.password {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"invalid_credentials"}`)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": email,
		"exp":  time.Now().Add(time.Hour * time.Duration(1)).Unix(),
		"iat":  time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(APP_KEY))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error":"token_generation_failed"}`)
		return
	}
	_, err = db.Exec("update users set token = ? where id = ?", tokenString, user.id)
	if err != nil{
        panic(err)
    }
	io.WriteString(w, `{"token":"`+tokenString+`"}`)
	return
}

func AuthMiddleware(next http.Handler) http.Handler {
	if len(APP_KEY) == 0 {
		log.Fatal("HTTP server unable to start, expected an APP_KEY for JWT auth")
	}
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(APP_KEY), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	return jwtMiddleware.Handler(next)
}

func main() {
	handleRequests()
}
