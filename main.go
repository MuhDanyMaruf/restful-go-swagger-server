package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	_ "server-go-swagger/docs" // Sesuaikan dengan nama modul Anda

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

var jwtKey = []byte("your_secret_key")

type Baju struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Size  string `json:"size"`
    Price int    `json:"price"`
}

type UserCredentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func main() {
    router := mux.NewRouter()

    router.HandleFunc("/login", loginHandler).Methods("POST")
    router.HandleFunc("/baju", getBajus).Methods("GET")
    router.HandleFunc("/baju/{id}", getBaju).Methods("GET")
    router.HandleFunc("/baju", createBaju).Methods("POST")
    router.HandleFunc("/baju/{id}", updateBaju).Methods("PUT")
    router.HandleFunc("/baju/{id}", deleteBaju).Methods("DELETE")

    // Swagger UI route
    router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type", "Authorization"},
    })

    handler := c.Handler(router)

    port := 8081 // Ubah port di sini
    fmt.Printf("Server is running on :%d\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}

// loginHandler, generateJWT, authenticateToken, getBajus, getBaju, createBaju, updateBaju, deleteBaju
// Add the remaining handler functions here

// @Summary Login
// @Description Logs user in with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body UserCredentials true "User credentials"
// @Success 200 {string} string "token"
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Router /login [post]
func loginHandler(w http.ResponseWriter, r *http.Request) {
    var user UserCredentials

    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/inv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    var storedPassword string
    err = db.QueryRow("SELECT password FROM user WHERE username = ?", user.Username).Scan(&storedPassword)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Invalid username or password", http.StatusUnauthorized)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if storedPassword != user.Password {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    tokenString, err := generateJWT()
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func generateJWT() (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)

    claims := token.Claims.(jwt.MapClaims)
    claims["authorized"] = true
    claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

func authenticateToken(tokenString string) (bool, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil {
        return false, err
    }
    return token.Valid, nil
}

// @Summary Menampilkan daftar baju
// @Description Mengembalikan daftar semua baju
// @Tags baju
// @Accept  json
// @Produce  json
// @Success 200 {array} Baju
// @Router /baju [get]
func getBajus(w http.ResponseWriter, r *http.Request) {
    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/inv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    rows, err := db.Query("SELECT id, name, size, price FROM baju")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var bajus []Baju
    for rows.Next() {
        var b Baju
        err := rows.Scan(&b.ID, &b.Name, &b.Size, &b.Price)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        bajus = append(bajus, b)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bajus)
}

// @Summary Get a Baju by ID
// @Description Get a Baju from the database by its ID
// @Tags Baju
// @Accept json
// @Produce json
// @Param id path int true "Baju ID"
// @Success 200 {object} Baju
// @Failure 404 {string} string "Baju not found"
// @Failure 500 {string} string "Internal server error"
// @Router /baju/{id} [get]
func getBaju(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    bajuID := params["id"]

    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/inv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    row := db.QueryRow("SELECT id, name, size, price FROM baju WHERE id = ?", bajuID)

        var baju Baju
    err = row.Scan(&baju.ID, &baju.Name, &baju.Size, &baju.Price)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Baju not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(baju)
}

// @Summary Create a new Baju
// @Description Create a new Baju and add it to the database
// @Tags Baju
// @Accept json
// @Produce json
// @Param input body Baju true "New Baju details"
// @Success 201 {string} string "Baju successfully created"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /baju [post]
func createBaju(w http.ResponseWriter, r *http.Request) {
    var newBaju Baju
    err := json.NewDecoder(r.Body).Decode(&newBaju)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/inv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("INSERT INTO baju (name, size, price) VALUES (?, ?, ?)", newBaju.Name, newBaju.Size, newBaju.Price)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Baju successfully created"))
}

// @Summary Update a Baju
// @Description Update an existing Baju in the database
// @Tags Baju
// @Accept json
// @Produce json
// @Param id path int true "Baju ID"
// @Param input body Baju true "Updated Baju details"
// @Success 200 {string} string "Baju successfully updated"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /baju/{id} [put]
func updateBaju(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    bajuID := params["id"]

    var updatedBaju Baju
    err := json.NewDecoder(r.Body).Decode(&updatedBaju)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/inv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("UPDATE baju SET name = ?, size = ?, price = ? WHERE id = ?", updatedBaju.Name, updatedBaju.Size, updatedBaju.Price, bajuID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Baju successfully updated"))
}

// @Summary Delete a Baju
// @Description Delete a Baju from the database by its ID
// @Tags Baju
// @Accept json
// @Produce json
// @Param id path int true "Baju ID"
// @Success 200 {string} string "Baju successfully deleted"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /baju/{id} [delete]
func deleteBaju(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w, "Invalid Baju ID", http.StatusBadRequest)
        return
    }

    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/inv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    _, err = db.Exec("DELETE FROM baju WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Baju with ID " + strconv.Itoa(id) + " deleted successfully"))
}
