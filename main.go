package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error

func initDB() {
	db, err = sql.Open("mysql", "username:password@tcp(IP:PORT)/namedata") //use link sqldatabase ex.(username:password@tcp(IP:PORT)/namedata) of you
	if err != nil {
		fmt.Println("failed to Connected")
	} else {
		fmt.Println("Connected to database")
	}
}

func createTable(db *sql.DB) { //func Create Table in database
	query := `CREATE TABLE beers(
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255),
		typebeer VARCHAR(255),
		details VARCHAR(255),
		imagepath VARCHAR(255)
	);`

	if _, err = db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

type Beer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Typebeer  string `json:"typebeer"`
	Details   string `json:"details"`
	ImagePath string `json:"imagepath"`
}

func saveBeerToDB(beer Beer) error {
	_, err := db.Exec(`INSERT INTO beers (name, typebeer, details, imagepath) VALUES (?, ?, ?, ?);`, beer.Name, beer.Typebeer, beer.Details, beer.ImagePath)
	return err
}

func updateBeerToDB(beer Beer) error {
	_, err := db.Exec(`UPDATE beers SET name = ?, typebeer = ?, details = ?, imagepath = ? WHERE id = ? `, beer.Name, beer.Typebeer, beer.Details, beer.ImagePath, beer.ID)
	return err
}

func deleteBeerToDB(id int) error {
	_, err := db.Exec(`DELETE FROM beers WHERE id = ?`, id)
	return err
}

func getBeers(c *gin.Context) {
	// Read the beer name given from the search parameters.
	beerName := c.Query("name")

	// Reads the specified page from the pagination parameter
	pageStr := c.Query("page")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		c.String(http.StatusInternalServerError, "This page was not found")
	}

	// Set the size of the information displayed on each page
	pageSize := 10

	query := "SELECT * FROM beers WHERE name = ? LIMIT ?, ?"

	// Calculate initial data in pagination
	startIndex := (page - 1) * pageSize

	rows, err := db.Query(query, beerName, startIndex, pageSize)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	} // SQL to the database
	defer rows.Close()

	filteredBeers := []Beer{}

	for rows.Next() {
		var beer Beer

		// Read each column in a row
		err := rows.Scan(&beer.ID, &beer.Name, &beer.Typebeer, &beer.Details, &beer.ImagePath)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		filteredBeers = append(filteredBeers, beer)
	}

	c.JSON(http.StatusOK, filteredBeers)
}

func addbeer(c *gin.Context) {
	//receive image file
	file, err := c.FormFile("image")
	if err != nil {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}

	//Save the image file in the uploads folder
	imagePath := "uploads/" + file.Filename
	err = c.SaveUploadedFile(file, imagePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	//save data to store value
	beer := Beer{
		Name:      c.PostForm("name"),
		Typebeer:  c.PostForm("typebeer"),
		Details:   c.PostForm("details"),
		ImagePath: imagePath,
	}

	//Pass the beer value to the function saveBeerToDB to insert the database
	err = saveBeerToDB(beer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to created beer in database")
		return
	}

	c.String(http.StatusOK, "Beer created successfully")
}

func updatebeer(c *gin.Context) {
	id := c.Param("id")
	// convert text to numbers
	beerID, err := strconv.Atoi(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid beer ID")
		return
	}
	//receive image file
	file, err := c.FormFile("image")
	if err != nil {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}
	//uSave the image file in the uploads folder
	imagePath := "uploads/" + file.Filename
	err = c.SaveUploadedFile(file, imagePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	//save data to store value
	beer := Beer{
		ID:        beerID,
		Name:      c.PostForm("name"),
		Typebeer:  c.PostForm("typebeer"),
		Details:   c.PostForm("details"),
		ImagePath: imagePath,
	}
	//Pass the beer value to the function updateBeerToDB to update the database
	err = updateBeerToDB(beer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update beer in database")
		return
	}

	c.String(http.StatusOK, "Beer updated successfully")
}

func deletebeer(c *gin.Context) {
	id := c.Param("id")
	// convert text to numbers
	beerID, err := strconv.Atoi(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid beer ID")
		return
	}
	// Pass the beerID value to the function deleteBeerToD to delete the database
	err = deleteBeerToDB(beerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete beer in database")
		return
	}
	c.String(http.StatusOK, "Beer Delete successfully")
}

func main() {
	initDB()
	createTable(db)
	r := gin.Default()
	r.GET("/beer", getBeers)
	r.POST("/beer", addbeer)
	r.PUT("/beer/:id", updatebeer)
	r.DELETE("/beer/:id", deletebeer)
	r.Run(":8080")
}
