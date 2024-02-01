package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	db := sqlx.MustConnect("postgres", getDbUrl())

	r := gin.Default()
	r.LoadHTMLGlob("views/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/stats", gin.BasicAuth(gin.Accounts{
		"admin": "admin",
	}), func(c *gin.Context) {
		cities, err := getLastCities(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "stats.html", cities)
	})

	r.GET("/weather", func(c *gin.Context) {
		city := c.Query("city")
		latlong, err := getLatLong(db, city)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		weather, err := getWeather(*latlong)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		weatherDisplay, err := extractWeatherData(city, weather)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "weather.html", weatherDisplay)
	})

	r.Run()
}

func getLatLong(db *sqlx.DB, city string) (*LatLong, error) {
	var latLong *LatLong
	err := db.Get(&latLong, "SELECT lat, long FROM cities WHERE name = $1", city)
	if err == nil {
		fmt.Println("found city in db")
		return latLong, nil
	}

	latLong, err = fetchLatLong(city)
	if err != nil {
		return nil, err
	}

	err = insertCity(db, city, *latLong)
	if err != nil {
		return nil, err
	}

	return latLong, nil
}

func fetchLatLong(city string) (*LatLong, error) {
	endpoint := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", url.QueryEscape(city))
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error making request to Geo API: %w", err)
	}
	defer resp.Body.Close()

	var response GeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response from Geo API: %w", err)
	}

	if len(response.Results) < 1 {
		return nil, fmt.Errorf("no results returned from Geo API")
	}

	return &response.Results[0], nil
}

func getWeather(latLong LatLong) (string, error) {
	endpoint := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.6f&longitude=%.6f&hourly=temperature_2m", latLong.Latitude, latLong.Longitude)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("error making request to Weather API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response from Weather API: %w", err)
	}

	return string(body), nil
}

func extractWeatherData(city, rawWeather string) (WeatherDisplay, error) {
	var weatherResponse WeatherResponse
	if err := json.Unmarshal([]byte(rawWeather), &weatherResponse); err != nil {
		return WeatherDisplay{}, fmt.Errorf("error decoding response from Weather API: %w", err)
	}

	var forecasts []Forecast
	layout := "2006-01-02T15:04"
	for i, t := range weatherResponse.Hourly.Time {
		date, err := time.Parse(layout, t)
		if err != nil {
			return WeatherDisplay{}, fmt.Errorf("error parsing time: %w", err)
		}
		forecast := Forecast{
			Date:        date.Format("Mon 15:04"),
			Temperature: fmt.Sprintf("%.1f°C", weatherResponse.Hourly.Temperature2m[i]),
		}
		forecasts = append(forecasts, forecast)
	}
	return WeatherDisplay{
		City:      city,
		Forecasts: forecasts,
	}, nil
}

func insertCity(db *sqlx.DB, name string, latLong LatLong) error {
	_, err := db.Exec("INSERT INTO cities (name, lat, long) VALUES ($1, $2, $3)", name, latLong.Latitude, latLong.Longitude)
	return err
}

func getLastCities(db *sqlx.DB) ([]string, error) {
	var cities []string
	err := db.Select(&cities, "SELECT name FROM cities ORDER BY id DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	return cities, nil
}
