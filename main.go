package main

import (
	"encoding/json"
	"fmt"
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
	ReassignmentCount     uint      `json:"reassignment_count"`
	ReopenCount           uint      `json:"reopen_count"`
	SysModCount           uint      `json:"sys_mod_count"`
	MadeSla               bool      `json:"made_sla"`
	CallerId              string    `json:"caller_id"`
	OpenedBy              string    `json:"opened_by" gorm:"index"`
	OpenedAt              time.Time `json:"opened_at" gorm:"index"`
	SysCreatedBy          string    `json:"sys_created_by"`
	SysCreatedAt          time.Time `json:"sys_created_at"`
	SysUpdatedBy          string    `json:"sys_updated_by"`
	SysUpdatedAt          time.Time `json:"sys_updated_at"`
	ContactType           string    `json:"contact_type"`
	Location              string    `json:"location"`
	Category              string    `json:"category" gorm:"index"`
	Subcategory           string    `json:"subcategory" gorm:"index"`
	USymptom              string    `json:"u_symptom"`
	CmdbCi                string    `json:"cmdb_ci"`
	Impact                string    `json:"impact" gorm:"index"`
	Urgency               string    `json:"urgency" gorm:"index"`
	Priority              string    `json:"priority" gorm:"index"`
	AssignmentGroup       string    `json:"assignment_group"`
	AssignedTo            string    `json:"assigned_to" gorm:"index"`
	Knowledge             bool      `json:"knowledge"`
	UPriorityConfirmation bool      `json:"u_priority_confirmation"`
	Notify                string    `json:"notify"`
	ProblemId             string    `json:"problem_id"`
	Rfc                   string    `json:"rfc"`
	Vendor                string    `json:"vendor"`
	CausedBy              string    `json:"caused_by"`
	ClosedCode            string    `json:"closed_code"`
	ResolvedBy            string    `json:"resolved_by"`
	ResolvedAt            time.Time `json:"resolved_at"`
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
		"subcategory",
		"impact",
		"urgency",
		"priority",
		"assigned_to":
		return true
	}
	return false
}

func paginate(ctx echo.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(ctx.QueryParam("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
		if limit <= 0 || limit > 100 {
			limit = 100
		}
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

func handleRequest(db *gorm.DB) {
	server := echo.New()
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
