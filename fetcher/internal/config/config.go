package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	BotUsername string
	APIURL      string
	IP          string
	Port        string
	Phone       string
	Password    string
	APIID       int
	APIHash     string
}

func Load() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func Get() Config {
	c := Config{}

	c.BotUsername = os.Getenv("BOT_USERNAME")
	if c.BotUsername == "" {
		log.Fatal("BOT_USERNAME not found in .env")
	}

	c.APIURL = os.Getenv("API_URL")
	if c.APIURL == "" {
		log.Fatal("API URL not found in .env")
	}

	c.IP = os.Getenv("IP")
	if c.IP == "" {
		log.Fatal("IP not found in .env")
	}

	c.Port = os.Getenv("PORT")
	if c.Port == "" {
		log.Fatal("Port not found in .env")
	}

	c.Phone = os.Getenv("PHONE")
	if c.Phone == "" {
		log.Fatal("Phone not found in .env")
	}

	c.Password = os.Getenv("PASSWORD")
	if c.Password == "" {
		log.Fatal("Password not found in .env")
	}

	API_ID_str := os.Getenv("API_ID")
	if API_ID_str == "" {
		log.Fatal("API ID not found in .env")
	}
	var err error
	if c.APIID, err = strconv.Atoi(API_ID_str); err != nil {
		log.Fatalf("Error parsing API ID to int: %v", err)
	}

	c.APIHash = os.Getenv("API_HASH")
	if c.APIHash == "" {
		log.Fatal("API hash not found in .env")
	}

	return c
}
