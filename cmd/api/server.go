package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mw "github.com/broscoe7/go-school-api/internal/api/middlewares"
	"github.com/broscoe7/go-school-api/internal/api/router"
	"github.com/broscoe7/go-school-api/pkg/utils"
	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

// function sets up environment variables for local deployment
func loadEnvFromEmbeddedFile() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}
	// create a temp file to load the variables
	tempfile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}
	defer os.Remove(tempfile.Name())

	// write embedded .env file to temp file
	_, err = tempfile.Write(content)
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}
	err = tempfile.Close()
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}

	// load env variables
	err = godotenv.Load(tempfile.Name())
	if err != nil {
		log.Fatalf("Error loading .env variables: %v", err)
	}
}

func main() {
	// only for production, for running source code
	// err := godotenv.Load("cmd/api/.env")
	// if err != nil {
	// 	panic(err)
	// }

	// load environment variables from embedded .env file
	loadEnvFromEmbeddedFile()

	port := os.Getenv("API_PORT")
	// cert := "cert.pem"
	// key := "key.pem"
	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	mux := router.Router()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := mw.NewRateLimiter(15, time.Minute)
	hppOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		Whitelist:                   []string{"sortby", "sortOrder", "first_name", "last_name", "class", "email", "subject", "password", "username", "role"},
	}
	jwtMiddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login", "/execs/forgot-password", "/execs/password/reset")
	secureMux := utils.ApplyMiddlewares(
		mux,
		mw.SecurityHeaders,
		mw.Compression,
		mw.Hpp(hppOptions),
		mw.XSSMiddleware,
		jwtMiddleware,
		mw.ResponseTime,
		rl.Middleware,
		mw.Cors,
	)

	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is listening on port ", port)
	if err := server.ListenAndServeTLS(cert, key); err != nil {
		log.Fatalln("Error starting server: ", err)
	}
}
