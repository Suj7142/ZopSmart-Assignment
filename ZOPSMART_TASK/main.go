package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"gofr.dev/pkg/gofr"
)

var db *sql.DB

type Guest struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	RoomNumber int    `json:"room_number"`
}

type Room struct {
	ID       int    `json:"id"`
	Number   int    `json:"number"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

func createDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "hotel.db")
	if err != nil {
		fmt.Println("error opening the database")
	}
	createGuestTableQuery := `CREATE TABLE guests(id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(255), room_number INTEGER);`
	createRoomTableQuery := `CREATE TABLE rooms(id INTEGER PRIMARY KEY AUTOINCREMENT, number INTEGER, capacity INTEGER, status VARCHAR(50));`
	_, err = db.Exec(createGuestTableQuery)
	if err != nil {
		fmt.Println("Error creating guests table")
	}
	_, err = db.Exec(createRoomTableQuery)
	if err != nil {
		fmt.Println("Error creating rooms table")
	}
}

func checkInGuest(guest Guest) error {
	_, err := db.Exec("INSERT INTO guests(name, room_number) VALUES (?, ?)", guest.Name, guest.RoomNumber)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE rooms SET status='occupied' WHERE number=?", guest.RoomNumber)
	return err
}

func checkOutGuest(id int) error {
	var roomNumber int
	err := db.QueryRow("SELECT room_number FROM guests WHERE id=?", id).Scan(&roomNumber)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM guests WHERE id=?", id)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE rooms SET status='available' WHERE number=?", roomNumber)
	return err
}

func viewRooms() ([]Room, error) {
	rows, err := db.Query("SELECT * FROM rooms")
	if err != nil {
		fmt.Println("Error while executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var r Room
		err := rows.Scan(&r.ID, &r.Number, &r.Capacity, &r.Status)
		if err != nil {
			fmt.Println("Error while scanning row:", err)
			return nil, err
		}
		rooms = append(rooms, r)
	}

	return rooms, nil
}

func main() {
	app := gofr.New()
	createDatabase()

	app.GET("/", func(ctx *gofr.Context) (interface{}, error) {
		return "Welcome to Hotel Management API", nil
	})

	app.POST("/checkin", func(ctx *gofr.Context) (interface{}, error) {
		var guest Guest
		if err := json.NewDecoder(ctx.Request().Body).Decode(&guest); err != nil {
			return nil, err
		}
		err := checkInGuest(guest)
		if err != nil {
			return nil, err
		}
		return guest, nil
	})

	app.GET("/checkout/:id", func(ctx *gofr.Context) (interface{}, error) {
		idParam := ctx.Param("id")
		if idParam == "" {
			return nil, fmt.Errorf("ID not provided")
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			return nil, fmt.Errorf("invalid format")
		}

		err = checkOutGuest(id)
		if err != nil {
			fmt.Println("Couldn't check out guest:", err)
			return nil, err
		}

		return "Guest checked out successfully", nil
	})

	app.GET("/viewrooms", func(ctx *gofr.Context) (interface{}, error) {
		rooms, err := viewRooms()
		if err != nil {
			fmt.Println("Could not view rooms")
			return nil, err
		}
		return rooms, nil
	})

	app.Start()
}
