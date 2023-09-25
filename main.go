package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"
)

const openWeatherAPIURL = "https://api.openweathermap.org/data/2.5/weather"

type WeatherResponse struct {
    Coord struct {
        Lon float64 `json:"lon"`
        Lat float64 `json:"lat"`
    } `json:"coord"`
    Weather []struct {
        ID          int    `json:"id"`
        Main        string `json:"main"`
        Description string `json:"description"`
    } `json:"weather"`
    Main struct {
        Temp float64 `json:"temp"`
    } `json:"main"`
    Alerts []struct {
        Event      string `json:"event"`
        Description string `json:"description"`
    } `json:"alerts,omitempty"` // Alerts can be omitted if not present
}

func getWeatherInfo(lat, lon float64, apiKey string) (*WeatherResponse, error) {
    apiURL := fmt.Sprintf("%s?lat=%f&lon=%f&appid=%s", openWeatherAPIURL, lat, lon, apiKey)

    resp, err := http.Get(apiURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
    }

    var weatherResponse WeatherResponse
    if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
        return nil, err
    }

    return &weatherResponse, nil
}

func determineWeatherCondition(temperature float64) string {
    if temperature < 0 {
        return "cold"
    } else if temperature > 30 {
        return "hot"
    } else {
        return "moderate"
    }
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    latStr := query.Get("lat")
    lonStr := query.Get("lon")

    apiKey := os.Getenv("OPENWEATHER_API_KEY")
    if apiKey == "" {
        http.Error(w, "Open Weather API key not set", http.StatusInternalServerError)
        return
    }

    lat, err := strconv.ParseFloat(latStr, 64)
    if err != nil {
        http.Error(w, "Invalid latitude", http.StatusBadRequest)
        return
    }

    lon, err := strconv.ParseFloat(lonStr, 64)
    if err != nil {
        http.Error(w, "Invalid longitude", http.StatusBadRequest)
        return
    }

    weatherResponse, err := getWeatherInfo(lat, lon, apiKey)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error fetching weather data: %s", err), http.StatusInternalServerError)
        return
    }

    weatherCondition := determineWeatherCondition(weatherResponse.Main.Temp)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    response := map[string]string{
        "WeatherCondition": weatherResponse.Weather[0].Description,
        "Temperature":      fmt.Sprintf("%.2f", weatherResponse.Main.Temp),
        "TemperatureType":  weatherCondition,
    }

    if len(weatherResponse.Alerts) > 0 {
        response["Alerts"] = weatherResponse.Alerts
    }

    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, fmt.Sprintf("Error encoding JSON response: %s", err), http.StatusInternalServerError)
        return
    }
}

func main() {
    http.HandleFunc("/weather", weatherHandler)
    port := ":8080"
    fmt.Printf("Server is listening on %s...\n", port)
    http.ListenAndServe(port, nil)
}
