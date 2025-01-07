package main

import (
	"net/http"
	"os/exec"
	"text/template"
)

type PageData struct {
	Mode string
}

var templates = template.Must(template.ParseFiles("index.html"))

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/submit", submitHandler)
	http.ListenAndServe(":8000", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{Mode: "Automatic"}
	templates.ExecuteTemplate(w, "index.html", data)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		mode := r.FormValue("mode")
		filePath := r.FormValue("filePath")
		// Here you would handle the logic for each mode
		switch mode {
		case "automatic":
			// Call your Go function or server logic here
			go exec.Command("your-command", filePath).Run()
		case "interactive":
			go exec.Command("your-command", filePath).Run()
		case "ngram":
			go exec.Command("your-command", filePath).Run()
		case "heuristic":
			go exec.Command("your-command", filePath).Run()
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
