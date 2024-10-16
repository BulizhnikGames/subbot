package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const SUB_LIMIT = 5

type Config struct {
	Phone     string
	Password  string
	API_ID    int
	API_hash  string
	Bot_token string
	DB_URL    string
}

func Load() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func Get() Config {
	c := Config{}

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
	if c.API_ID, err = strconv.Atoi(API_ID_str); err != nil {
		log.Fatalf("Error parsing API ID to int: %v", err)
	}

	c.API_hash = os.Getenv("API_HASH")
	if c.API_hash == "" {
		log.Fatal("API hash not found in .env")
	}

	c.Bot_token = os.Getenv("BOT_TOKEN")
	if c.Bot_token == "" {
		log.Fatal("Bot token not found in .env")
	}

	c.DB_URL = os.Getenv("DD_URL")
	if c.DB_URL == "" {
		log.Fatal("DB URL not found in .env")
	}

	return c
}
