package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email`
	Status string `json:"status"`
}

func createCustomer(customer Customer) int {
	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3)  RETURNING id", customer.Name, customer.Email, customer.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		fmt.Println("Can't scan id", err)
		return id
	}
	fmt.Println("Insert Customer success id:", id)
	return id
}

func createCustomerHandler(c *gin.Context) {
	customer := Customer{}

	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": err.Error()})
		return
	}
	customerId := createCustomer(customer)
	if customerId != 0 {
		customer.ID = createCustomer(customer)
		c.JSON(http.StatusCreated, customer)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": "Something went wrong."})
	}

}

func getCustomerByID(ID int) Customer {
	customer := Customer{}
	stmt, err := db.Prepare("SELECT id,name,email,status FROM customers WHERE id=$1")
	if err != nil {
		log.Fatal("can't prepare query one row statment", err)
	}
	rowID := ID
	row := stmt.QueryRow(rowID)
	var id int
	var name, email, status string
	err = row.Scan(&id, &name, &email, &status)
	if err != nil {
		// log.Fatal("can't Scan row into variables ", err)
		fmt.Println("can't Scan row into variables ", err)
		return customer
	}
	fmt.Println("one row", id, name, email, status)
	customer.ID = id
	customer.Name = name
	customer.Email = email
	customer.Status = status
	return customer
}

func getCustomerByIDHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	customer := getCustomerByID(id)
	c.JSON(http.StatusOK, customer)
}

func getCustomers() []Customer {
	customers := []Customer{}
	stmt, err := db.Prepare("SELECT id,name,email,status FROM customers")
	if err != nil {
		log.Fatal("can't prepare query one row statment", err)
		return customers
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("can't query all customers", err)
		return customers
	}

	for rows.Next() {
		var item = Customer{}
		var id int
		var name, email, status string
		err := rows.Scan(&id, &name, &email, &status)
		if err != nil {
			log.Fatal("can't scan row into var", err)
			return customers
		}
		item.ID = id
		item.Name = name
		item.Email = email
		item.Status = status
		customers = append(customers, item)
	}
	fmt.Println("Query all customers success")
	return customers
}

func getCustomerHandler(c *gin.Context) {
	c.JSON(http.StatusOK, getCustomers())
}

func updateCustomer(c Customer) {

	stmt, err := db.Prepare("UPDATE customers SET name=$2,email=$3, status=$4 WHERE id=$1")
	if err != nil {
		log.Fatal("can't prepare query one row statment", err)
	}
	if _, err := stmt.Exec(c.ID, c.Name, c.Email, c.Status); err != nil {
		log.Fatal("error execute update ", err)
	}
	fmt.Println("Update Customer success.")

}

func updateCustomerHandler(c *gin.Context) {
	customer := Customer{}
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateCustomer(customer)

	c.JSON(http.StatusOK, getCustomerByID(customer.ID))
}

func deleteCustomer(ID int) string {
	stmt, err := db.Prepare("DELETE FROM customers WHERE id=$1")
	if err != nil {
		log.Fatal("can't prepare query one row statment", err)
	}
	rowID := ID
	_, err = stmt.Exec(rowID)
	if err != nil {
		return err.Error()
	} else {
		return "OK"
	}
}

func deleteCustomerHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	msg := deleteCustomer(id)
	if msg != "OK" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": msg})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/customers", getCustomerHandler)
	r.GET("/customers/:id", getCustomerByIDHandler)
	r.POST("/customers", createCustomerHandler)
	r.PUT("/customers", updateCustomerHandler)
	r.DELETE("/customers/:id", deleteCustomerHandler)

	return r
}

func main() {
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	defer db.Close()
	r := setupRouter()
	r.Run(":2019") // listen and serve on 127.0.0.0:8080

	log.Println("Close connection")
}
