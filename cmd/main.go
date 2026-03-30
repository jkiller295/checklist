package main

import (
	"checklist/internal/db"
	"checklist/internal/handlers"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	password := os.Getenv("HOUSEHOLD_PASSWORD")
	if password == "" {
		password = "changeme"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/checklist.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	funcMap := template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			m := make(map[string]any, len(values)/2)
			for i := 0; i+1 < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		"not": func(v any) bool {
			if v == nil {
				return true
			}
			switch val := v.(type) {
			case bool:
				return !val
			case int:
				return val == 0
			case string:
				return val == ""
			case []db.List:
				return len(val) == 0
			case []db.Item:
				return len(val) == 0
			}
			return false
		},
		"percent": func(done, total int) int {
			if total == 0 {
				return 0
			}
			return (done * 100) / total
		},
		"js": func(s string) template.JS {
			result := ""
			for _, c := range s {
				switch c {
				case '\'':
					result += `\'`
				case '\\':
					result += `\\`
				case '\n':
					result += `\n`
				case '\r':
					result += `\r`
				default:
					result += string(c)
				}
			}
			return template.JS(result)
		},
	}

	tmpls := template.New("").Funcs(funcMap)

	tmpls, err = tmpls.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	tmpls, err = tmpls.ParseGlob("templates/icons/*.html")
	if err != nil {
		log.Fatalf("failed to parse icon templates: %v", err)
	}

	srv := &handlers.Server{
		DB:       database,
		Password: password,
		Tmpls:    tmpls,
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	srv.RegisterRoutes(mux)

	handler := srv.AuthMiddleware(mux)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
