package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/middleware"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
)

type Incident struct {
	gorm.Model            `json:"model"`
	Number                string    `json:"number" gorm:"index"`
	IncidentState         string    `json:"incident_state" gorm:"index"`
	Active                bool      `json:"active" gorm:"index"`
	CallerId              string    `json:"caller_id"`
	OpenedBy              string    `json:"opened_by" gorm:"index"`
	OpenedAt              time.Time `json:"opened_at" gorm:"index"`
	ContactType           string    `json:"contact_type"`
	Location              string    `json:"location"`
	Category              string    `json:"category" gorm:"index"`
	Urgency                string    `json:"urgency" gorm:"index"`
	AssignmentGroup       string    `json:"assignment_group"`
	ClosedCode            string    `json:"closed_code"`
	ClosedAt              time.Time `json:"closed_at"`
}

type Secret struct {
	DBUser     string `json:"DB_USER"`
	DBPassword string `json:"DB_PASSWORD"`
}

func allIncidents(db *gorm.DB) func(echo.Context) error {
	return func(ctx echo.Context) error {
		var incidents []Incident
		queryParams := make(map[string]interface{})
		for filter, value := range ctx.QueryParams() {
			if isValidFilter(filter) {
				queryParams[filter] = value[0]
			}
		}
		db.Scopes(paginate(ctx)).Where(queryParams).Find(&incidents)
		return ctx.JSON(http.StatusOK, incidents)
	}
}

func singleIncident(db *gorm.DB) func(echo.Context) error {
	return func(ctx echo.Context) error {
		var incidents []Incident
		db.Scopes(paginate(ctx)).Where("number = ?", ctx.Param("number")).Find(&incidents)
		return ctx.JSON(http.StatusOK, incidents)
	}
}

func healthCheck(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "OK")
}

func isValidFilter(filter string) bool {
	switch filter {
	case
		"incident_state",
		"opened_by",
		"category",
		"urgency",
		"assignment_group":
		return true
	}
	return false
}

func paginate(ctx echo.Context) func(db *gorm.DB) *gorm.DB {
	maxPageSize, err := strconv.Atoi(os.Getenv("MAX_PAGE_SIZE"))
	if err != nil {
		fmt.Println(err.Error())
		panic("invalid max page size number")
	}
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(ctx.QueryParam("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
		if limit <= 0 || limit > maxPageSize {
			limit = maxPageSize
		}
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

func handleRequest(db *gorm.DB) {
	server := echo.New()
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://brandonregard.info"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	server.GET("/incidents", allIncidents(db))
	server.GET("/incidents/:number", singleIncident(db))
	server.GET("/health", healthCheck)
	server.Logger.Fatal(server.Start(":3000"))
}

func initialMigration(db *gorm.DB) {
	db.AutoMigrate(&Incident{})
}

func connectionString() string {
	secret := getSecret(os.Getenv("SECRET"))
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fmt.Println(err.Error())
		panic("invalid port number")
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		secret.DBUser,
		secret.DBPassword,
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_NAME"),
	)
}

func getSecret(secretId string) Secret {
	var secretData Secret
	svc := secretsmanager.New(session.New())
	secret := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretId),
	}
	secretValue, err := svc.GetSecretValue(secret)
	if err != nil {
		fmt.Println(err.Error())
		panic("could not fetch database credentials")
	}
	err = json.Unmarshal([]byte(*secretValue.SecretString), &secretData)
	if err != nil {
		fmt.Println(err.Error())
		panic("could not parse secret")
	}
	return secretData
}

func main() {
	db, err := gorm.Open("mysql", connectionString())
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	defer db.Close()
	initialMigration(db)
	handleRequest(db)
}
