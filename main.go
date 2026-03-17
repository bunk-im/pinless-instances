package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const instancesURL = "https://raw.githubusercontent.com/bunk-im/pinless/main/instances.json"

type Instance struct {
	URL        string   `json:"url"`
	Regions    []string `json:"regions"`
	Operators  []string `json:"operators"`
	Cloudflare bool     `json:"cloudflare"`
}

type InstancesData struct {
	Clearnet []Instance `json:"clearnet"`
	Onion    []Instance `json:"onion"`
	I2P      []Instance `json:"i2p"`
}

var cachedInstances *InstancesData
var cacheTime time.Time
var cacheDuration = 5 * time.Minute

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	router := gin.Default()
	router.LoadHTMLFiles("templates/index.html")

	router.GET("/", func(c *gin.Context) {
		instances := fetchInstances()
		c.HTML(http.StatusOK, "index.html", instances)
	})

	router.GET("/api/instances", func(c *gin.Context) {
		instances := fetchInstances()
		c.JSON(http.StatusOK, instances)
	})

	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Println("Pinless Instances Server")
	fmt.Printf("Running at http://127.0.0.1:%s\n", port)

	router.Run(":" + port)
}

func fetchInstances() *InstancesData {
	if cachedInstances != nil && time.Since(cacheTime) < cacheDuration {
		return cachedInstances
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(instancesURL)
	if err != nil {
		return cachedInstances
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cachedInstances
	}

	var data InstancesData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return cachedInstances
	}

	cachedInstances = &data
	cacheTime = time.Now()
	return cachedInstances
}
