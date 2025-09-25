package main

import (
    "database/sql"
    "html/template"
    "log"
    "net/http"
    "os"

    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))
var db *sql.DB

func main() {
    // Init DB
    initDB()

    // Routes
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/signup", signupHandler)
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/product", productHandler)
    http.HandleFunc("/cart", cartHandler)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    log.Println("Server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
    var err error
    db, err = sql.Open("sqlite3", "./db.sqlite")
    if err != nil {
        log.Fatal(err)
    }
    // Create tables if not exist
    db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    )`)
    db.Exec(`CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        price REAL,
        description TEXT
    )`)
    db.Exec(`CREATE TABLE IF NOT EXISTS cart (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        product_id INTEGER
    )`)
    // Seed products
    db.Exec(`INSERT OR IGNORE INTO products (id, name, price, description) VALUES
        (1, 'Go T-Shirt', 19.99, 'A comfy Go-branded T-shirt'),
        (2, 'Go Mug', 9.99, 'A stylish mug for Go lovers')
    `)
}

// --- Handlers ---

func homeHandler(w http.ResponseWriter, r *http.Request) {
    rows, _ := db.Query("SELECT id, name, price FROM products")
    defer rows.Close()
    var products []map[string]interface{}
    for rows.Next() {
        var id int
        var name string
        var price float64
        rows.Scan(&id, &name, &price)
        products = append(products, map[string]interface{}{
            "ID":    id,
            "Name":  name,
            "Price": price,
        })
    }
    templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
        "Products": products,
        "Username": getSessionUser(r),
    })
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        username := r.FormValue("username")
        password := r.FormValue("password")
        hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        _, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hash))
        if err != nil {
            templates.ExecuteTemplate(w, "signup.html", "Username already taken!")
            return
        }
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    templates.ExecuteTemplate(w, "signup.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        username := r.FormValue("username")
        password := r.FormValue("password")
        row := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username)
        var id int
        var hash string
        err := row.Scan(&id, &hash)
        if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
            templates.ExecuteTemplate(w, "login.html", "Invalid credentials!")
            return
        }
        setSessionUser(w, username)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    templates.ExecuteTemplate(w, "login.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    clearSessionUser(w)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func productHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    row := db.QueryRow("SELECT name, price, description FROM products WHERE id = ?", id)
    var name string
    var price float64
    var desc string
    err := row.Scan(&name, &price, &desc)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    // Handle Add to Cart
    if r.Method == "POST" && getSessionUser(r) != "" {
        user := getSessionUser(r)
        var uid int
        db.QueryRow("SELECT id FROM users WHERE username = ?", user).Scan(&uid)
        db.Exec("INSERT INTO cart (user_id, product_id) VALUES (?, ?)", uid, id)
        http.Redirect(w, r, "/cart", http.StatusSeeOther)
        return
    }
    templates.ExecuteTemplate(w, "product.html", map[string]interface{}{
        "ID":          id,
        "Name":        name,
        "Price":       price,
        "Description": desc,
        "Username":    getSessionUser(r),
    })
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
    user := getSessionUser(r)
    if user == "" {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    var uid int
    db.QueryRow("SELECT id FROM users WHERE username = ?", user).Scan(&uid)
    rows, _ := db.Query("SELECT products.id, products.name, products.price FROM cart JOIN products ON cart.product_id = products.id WHERE cart.user_id = ?", uid)
    defer rows.Close()
    var items []map[string]interface{}
    for rows.Next() {
        var id int
        var name string
        var price float64
        rows.Scan(&id, &name, &price)
        items = append(items, map[string]interface{}{
            "ID":    id,
            "Name":  name,
            "Price": price,
        })
    }
    templates.ExecuteTemplate(w, "cart.html", map[string]interface{}{
        "Items":   items,
        "Username": user,
    })
}

// --- Session helpers (cookie-based, simple) ---

func setSessionUser(w http.ResponseWriter, username string) {
    cookie := &http.Cookie{
        Name:  "username",
        Value: username,
        Path:  "/",
        HttpOnly: true,
    }
    http.SetCookie(w, cookie)
}

func getSessionUser(r *http.Request) string {
    cookie, err := r.Cookie("username")
    if err != nil {
        return ""
    }
    return cookie.Value
}

func clearSessionUser(w http.ResponseWriter) {
    cookie := &http.Cookie{
        Name:   "username",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(w, cookie)
}