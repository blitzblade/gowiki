package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"fmt"
)

const(
	logInfo = "INFO"
	logWarning = "WARNING"
	logError = "ERROR"
)

type logEntry struct {
	time time.Time
	severity string
	message string
}
var logCh = make(chan logEntry, 50)
var doneCh = make(chan struct{})


type Page struct {
    Title string
    Body  []byte
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
    return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
    // t, _ := template.ParseFiles("view.html")
	// t.Execute(w, p)
	logCh <- logEntry{ time.Now(), logInfo, "View displayed successfully"}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    // t, _ := template.ParseFiles("edit.html")
    // t.Execute(w, p)
	logCh <- logEntry{ time.Now(), logInfo, "Edit done successfully!"}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request){
	title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
    p := Page{Title: title, Body: []byte(body)}
    // t, _ := template.ParseFiles("edit.html")
	ptr := &p
	err := ptr.save()
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // t.Execute(w, p)
	logCh <- logEntry{ time.Now(), logInfo, "Page saved successfully!"}
	logCh <- logEntry{ time.Now(), logInfo, "Redirecting..."}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    t, _ := template.ParseFiles(tmpl + ".html")
    t.Execute(w, p)
}

func logger(){
	// for entry := range logCh {
	// 	fmt.Printf("%v - [%v] %v\n", entry.time.Format("2006-01-02T15:04:05"), entry.severity, entry.message)
	// }

	for {
		select {
		case entry := <- logCh:
			fmt.Printf("%v - [%v] %v\n", entry.time.Format("2006-01-02T15:04:05"), entry.severity, entry.message)
		case <- doneCh:
			break
		}
	}
}


func main() {
	defer func(){
		log.Fatal("Tearing down server...")
		doneCh <- struct{}{}
	}()
    // p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
    // p1.save()
    // p2, _ := loadPage("TestPage")
    // fmt.Println(string(p2.Body))
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	
	go logger()
	
	logCh <- logEntry{ time.Now(), logInfo, "App is starting..."}
	logCh <- logEntry{ time.Now(), logInfo, "App is running..."}
    err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err.Error())
	}
}