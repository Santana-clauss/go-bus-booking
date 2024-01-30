package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"

	
	
)

type Bus struct {
	ID             int
	Description    string
	Seats          int
	Day            string
	Time           string
	Route          string
	TotalSeats     int
	SeatsRemaining int
}

type Student struct {
	AdmissionNumber string
	Password        string
	FavoriteWord    string
}

var (
	templates *template.Template
	db        *sql.DB
	
)
func init() {
	var err error

	db, err = sql.Open("mysql", "root:@Clauss022@tcp(127.0.0.1:3306)/transport_mgmt")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS buses (
			id INT AUTO_INCREMENT PRIMARY KEY,
			description VARCHAR(255),
			seats INT,
			day VARCHAR(20),
			time VARCHAR(5),
			route VARCHAR(255)
	)`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS students (
			admission_number VARCHAR(20) PRIMARY KEY,
			password         VARCHAR(255),
			favorite_word    VARCHAR(255)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/signup", signupHandler).Methods("GET", "POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")

	r.HandleFunc("/admin", adminHandler).Methods("GET")
	r.HandleFunc("/admin/add-bus", addBusHandler).Methods("POST")

	r.HandleFunc("/student", studentHandler).Methods("GET")
	r.HandleFunc("/payment", paymentHandler).Methods("GET")

	r.HandleFunc("/student/book-seat", bookSeatHandler).Methods("POST")
	r.HandleFunc("/student/get-buses-for-route", getBusesForRouteHandler).Methods("POST")
	r.HandleFunc("/student/complete-payment", completePaymentHandler).Methods("POST")

	//home
	r.HandleFunc("/home",homepageHandlerr).Methods("GET")
	

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	}).Methods("GET")
   //
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":7000", nil))
}
func homepageHandlerr(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		
		templates.ExecuteTemplate(w, "signup.html", nil)
	} else if r.Method == http.MethodPost {
		
		admissionNumber := r.FormValue("admissionNumber")
		password := r.FormValue("password")
		favoriteWord := r.FormValue("favoriteWord")

		
		match, _ := regexp.MatchString(`^[0-9-]+$`, admissionNumber)
		if !match {
			http.Error(w, "Invalid admission number format", http.StatusBadRequest)
			return
		}

		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		
		_, err = db.Exec("INSERT INTO students (admission_number, password, favorite_word) VALUES (?, ?, ?)",
			admissionNumber, string(hashedPassword), favoriteWord)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		admissionNumber := r.FormValue("admissionNumber")
		password := r.FormValue("password")

		
		var hashedPassword string
		err := db.QueryRow("SELECT password FROM students WHERE admission_number = ?", admissionNumber).Scan(&hashedPassword)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid admission number or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid admission number or password", http.StatusUnauthorized)
			return
		}

		
		http.Redirect(w, r, "/student", http.StatusSeeOther)
	} else {
	
		templates.ExecuteTemplate(w, "login.html", nil)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	var buses []Bus

	rows, err := db.Query("SELECT * FROM buses")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var bus Bus
		err := rows.Scan(&bus.ID, &bus.Description, &bus.Seats, &bus.Day, &bus.Time, &bus.Route)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		buses = append(buses, bus)
	}

	templates.ExecuteTemplate(w, "admin.html", buses)
}

func addBusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	description := r.FormValue("busDescription")
	seatsStr := r.FormValue("seats")
	day := r.FormValue("day")
	time := r.FormValue("time")
	route := r.FormValue("route")

	if !validateForm(description, seatsStr, day, time, route) {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	seats, err := strconv.Atoi(seatsStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid seats value", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO buses (description, seats, day, time, route) VALUES (?, ?, ?, ?, ?)",
		description, seats, day, time, route)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func studentHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "student.html", nil)
}
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "payment.html", nil)
}

func bookSeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	busIDStr := r.FormValue("busID")
	if busIDStr == "" {
		http.Error(w, "Invalid bus ID", http.StatusBadRequest)
		return
	}

	busID, err := strconv.Atoi(busIDStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid bus ID", http.StatusBadRequest)
		return
	}

	remainingSeats, err := getRemainingSeats(busID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if remainingSeats <= 0 {
		http.Error(w, "Seats are unavailable", http.StatusBadRequest)
		return
	}

	remainingSeats--

	_, err = db.Exec("UPDATE buses SET seats = ? WHERE id = ?", remainingSeats, busID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "Seat booked successfully",
		"remaining_seats":    remainingSeats,
		"seat_number":        remainingSeats + 1,
		"enable_payment_btn": true,
	})
}

func getBusesForRouteHandler(w http.ResponseWriter, r *http.Request) {
	route := r.FormValue("route")

	buses, err := getBusesForRoute(route)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, buses)
}

func getBusesForRoute(route string) ([]Bus, error) {
	var buses []Bus

	rows, err := db.Query("SELECT * FROM buses WHERE route = ?", route)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bus Bus
		err := rows.Scan(&bus.ID, &bus.Description, &bus.Seats, &bus.Day, &bus.Time, &bus.Route)
		if err != nil {
			return nil, err
		}
		bus.TotalSeats = bus.Seats
		bus.SeatsRemaining, err = getRemainingSeats(bus.ID)
		if err != nil {
			return nil, err
		}
		buses = append(buses, bus)
	}

	return buses, nil
}
func getRemainingSeats(busID int) (int, error) {
	var bookedSeats int

	err := db.QueryRow("SELECT seats FROM buses WHERE id = ?", busID).Scan(&bookedSeats)
	if err != nil {
		return 0, err
	}

	return bookedSeats, nil
}

func validateForm(description, seatsStr, day, time, route string) bool {
	seats, err := strconv.Atoi(seatsStr)
	if err != nil || seats <= 0 {
		return false
	}

	validDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	if !contains(validDays, day) {
		return false
	}
	return true
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func completePaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Payment completed successfully",
	})
}