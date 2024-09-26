package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type User struct {
	UserId            int    `json:"user_id"`
	ProfilePictureUrl string `json:"profile_picture_url"`
	Name              string `json:"name"`
	Score             int    `json:"score"`
}

type Response struct {
	question_id int
	option_id   int
}

type UserResponse struct {
	user_id   int
	responses []Response
	results   map[int]User
}

var db *sql.DB

func initializeDatabaseConnection() {

	dbError := godotenv.Load(".env")

	if dbError != nil {
		log.Fatalf("Error loading .env file")
	}

	DbHost := os.Getenv("DB_HOST")
	DbUser := os.Getenv("DB_USER")
	DbPassword := os.Getenv("DB_PASSWORD")
	DbName := os.Getenv("DB_NAME")
	DbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	var err error
	db, err = sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}

	fmt.Println("Database connection established")

}

func main() {
	initializeDatabaseConnection()
	r := gin.Default()
	r.GET("/calculate-result", calculateUserResult)

	r.Run(":8020")

}

func calculateUserResult(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Query("user_id"))
	userResponses := fetchUserResponse(userId)

	for _, answer := range userResponses.responses {
		getUsersWithSameResponse(answer, userId, &userResponses.results)
	}

	sortedResults := make([]User, 0, len(userResponses.results))

	for _, userResult := range userResponses.results {
		sortedResults = append(sortedResults, userResult)
	}

	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Score > sortedResults[j].Score
	})

	c.JSON(200, gin.H{
		"error": false,
		"data":  sortedResults,
	})
}

func fetchUserResponse(user_id int) UserResponse {
	userResponse := UserResponse{
		user_id: user_id,
		results: make(map[int]User),
	}

	stmt, err := db.Prepare("SELECT question_id, option_id FROM user_responses WHERE user_id = ?")

	if err != nil {
		log.Fatalf(err.Error())
	}

	defer stmt.Close()

	queryOptions, err := stmt.Query(user_id)

	if err != nil {
		log.Fatalf(err.Error())
	}

	for queryOptions.Next() {
		var question_id int
		var option_id int
		queryOptions.Scan(&question_id, &option_id)

		response := Response{
			option_id:   option_id,
			question_id: question_id,
		}

		userResponse.responses = append(userResponse.responses, response)

	}

	return userResponse
}

func getUsersWithSameResponse(answer Response, userId int, result *map[int]User) {

	stmt, err := db.Prepare("SELECT u.id, u.name, u.profile_picture_url FROM users u JOIN user_responses ur on u.id = ur.user_id WHERE ur.user_id != ? AND ur.question_id = ? AND ur.option_id = ?")

	if err != nil {
		log.Fatalf(err.Error())
	}

	defer stmt.Close()

	queryOptions, err := stmt.Query(userId, answer.question_id, answer.option_id)

	if err != nil {
		log.Fatalf(err.Error())
	}

	for queryOptions.Next() {
		var comparedUserId int
		var userName string
		var profilePictureUrl string

		queryOptions.Scan(&comparedUserId, &userName, &profilePictureUrl)
		if _, exists := (*result)[comparedUserId]; !exists {
			(*result)[comparedUserId] = User{
				UserId:            comparedUserId,
				Name:              userName,
				ProfilePictureUrl: profilePictureUrl,
				Score:             0,
			}
		}

		userResult := (*result)[comparedUserId]
		userResult.Score++
		(*result)[comparedUserId] = userResult
	}
}
