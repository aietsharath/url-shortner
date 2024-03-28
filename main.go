package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	supa "github.com/nedpals/supabase-go"
)

type request struct {
	URL          string `json:"url"`
	CustomeShort string `json:"short"`
}

type response struct {
	URL            string        `json:"url"`
	CustomeShort   string        `json:"short"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_response"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {

	/*database set-up*/
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	/* Read environment variables*/
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	databaseUrl := os.Getenv("DATABASE_URL")

	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database:", err)
	}
	app := fiber.New()
	app.Post("/api/shorten", func(c *fiber.Ctx) error {

		var req request
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		actualURL := req.URL

		shortURL := ShortenURL(actualURL)

		newRequest := request{
			URL:          actualURL,
			CustomeShort: shortURL,
		}

		var results []request

		if err := supabase.DB.From("requests").Insert(newRequest).Execute(&results); err != nil {

			return err
		}

		return c.JSON(results)
	})
	app.Get("/api/actualURL/:shortURL", func(c *fiber.Ctx) error {
		shortURL := c.Params("shortURL")

		var actualURL string
		err := db.QueryRow("SELECT url FROM requests WHERE short = $1", shortURL).Scan(&actualURL)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
			}
			return err
		}

		return c.JSON(fiber.Map{"actualURL": actualURL})
	})
	log.Fatal(app.Listen(":3000"))
}

func ShortenURL(actualURL string) string {
	fmt.Printf(actualURL)
	shortURL := generateShortURL(8)
	fmt.Println("Generated short URL:", shortURL)

	return shortURL
}

func generateShortURL(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
