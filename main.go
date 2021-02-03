package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"path"
	"html/template"
	"net/http/pprof"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	APP_KEY = "chepubelka"
)

type WeatherInfo struct {
	List []WeatherListItem `json:"list"`
}

type WeatherListItem struct {
	Time_weather string
	Main         WeatherMain `json:"main"`
}

type WeatherMain struct {
	Temp      float32 `json:"temp"`
	FeelsLike float32 `json:"feels_like"`
	Humidity  int     `json:"humidity"`
}

type User struct {
	id       int
	email    string
	password string
}

type LogsListItem struct {
	User_id		int 	`json:"user_id"`
	Email		string	`json:"email"`
	Date_time	string	`json:"date_time"`
	Ip_address	string	`json:"ip_address"`
}

type LogsList struct {
	List []LogsListItem `json:"list"`
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

func addLog(email string, operation_type string, ip_addr string, city string) {
	db, err := sql.Open("mysql", "root:XGalHeg7.@/golang")
	if err != nil {
		panic(err)
	}
	user := getUser(email)
	_, err = db.Exec("insert into logs(user_id, operation, date_time, ip_address, city) values (?, ?, ?, ?, ?)", user.id, operation_type, time.Now(), ip_addr, city)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}

func returnWeather(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	city := vars["city"]
	weather, err := getWeather(city)
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(weather.List); i++ {
		weather.List[i].Time_weather = time.Now().AddDate(0, 0, i).Format(time.ANSIC)
	}
	token, err := jwt.Parse(r.Header.Get("Authorization")[7:], nil)
	if token == nil {
		panic(err.Error())
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	addLog(claims["user"].(string), "get weather", r.RemoteAddr, city)
	json.NewEncoder(w).Encode(weather.List)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", render)
	myRouter.HandleFunc("/token", TokenHandler)
	myRouter.Handle("/weather/{city}", AuthMiddleware(http.HandlerFunc(returnWeather)))
	myRouter.Handle("/logs/{city}", AuthMiddleware(http.HandlerFunc(getLogs)))
	myRouter.HandleFunc("/debug/pprof/", pprof.Index)
    myRouter.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    myRouter.HandleFunc("/debug/pprof/profile", pprof.Profile)
    myRouter.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    myRouter.HandleFunc("/debug/pprof/trace", pprof.Trace)
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	user := getUser(email)
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
	db, err = sql.Open("mysql", "root:XGalHeg7.@/golang")
	_, err = db.Exec("update users set token = ? where id = ?", tokenString, user.id)
	if err != nil {
		panic(err)
	}
	addLog(email, "user authorized", r.RemoteAddr, "")
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

func getUser(email string) *User {
	db, err := sql.Open("mysql", "root:XGalHeg7.@/golang")

	if err != nil {
		panic(err)
	}
	defer db.Close()
	var user = new(User)
	err = db.QueryRow("select id, email, password from users where email = ?", email).Scan(&user.id, &user.email, &user.password)
	if err != nil {
		panic(err.Error())
	}
	return user
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	city := vars["city"]
	db, err := sql.Open("mysql", "root:XGalHeg7.@/golang")

	if err != nil {
		panic(err)
	}
	defer db.Close()
	rows, err := db.Query("select users.id, users.email, logs.date_time, logs.ip_address from users join logs on logs.user_id = users.id where logs.city = ?", city)
	logs := LogsList{}
	for rows.Next() {
		l := LogsListItem{}
		err := rows.Scan(&l.User_id, &l.Email, &l.Date_time, &l.Ip_address)
		if err != nil{
            fmt.Println(err)
            continue
		}
		logs.List = append(logs.List, l)
	}
	w.Header().Set("Content-Type", "application/json")
	responseLogs, _ := json.Marshal(logs)
	w.Write(responseLogs)
	return
}

func render(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("views", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
	if err := tmpl.Execute(w, ""); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
	handleRequests()
}
