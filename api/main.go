package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"browser"
)

func main() {
	headless := flag.Bool("headless", true, "headless mode")
	port := flag.Int("port", 8080, "listen port")

	flag.Parse()

	br, err := browser.New(*headless)
	if err != nil {
		log.Fatal(err)
	}
	defer br.Close()

	http.HandleFunc("POST /navigate", navigate)
	http.HandleFunc("GET /snapshot", snapshot)
	http.HandleFunc("POST /click", click)
	http.HandleFunc("POST /type", typ)
	http.HandleFunc("POST /scroll", scroll)
	http.HandleFunc("GET /url", getURL)
	http.HandleFunc("GET /html", getHTML)
	http.HandleFunc("GET /text", getText)
	http.HandleFunc("POST /back", back)
	http.HandleFunc("POST /forward", forward)
	http.HandleFunc("POST /refresh", refresh)
	http.HandleFunc("POST /screenshot", screenshot)

	log.Printf("browser api listening on :%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

var br *browser.Browser

type resp struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func jsonResp(w http.ResponseWriter, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(resp{Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(resp{Data: data})
}

func navigate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	jsonResp(w, nil, br.Navigate(req.URL))
}

func snapshot(w http.ResponseWriter, r *http.Request) {
	s, err := br.Snapshot()
	jsonResp(w, map[string]string{"snapshot": s}, err)
}

func click(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ref string `json:"ref"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	jsonResp(w, nil, br.ClickByRef(req.Ref))
}

func typ(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ref  string `json:"ref"`
		Text string `json:"text"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	jsonResp(w, nil, br.TypeByRef(req.Ref, req.Text))
}

func scroll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ref string `json:"ref"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	jsonResp(w, nil, br.ScrollIntoViewByRef(req.Ref))
}

func getURL(w http.ResponseWriter, r *http.Request) {
	u, err := br.GetCurrentURL()
	jsonResp(w, map[string]string{"url": u}, err)
}

func getHTML(w http.ResponseWriter, r *http.Request) {
	h, err := br.GetHTML()
	jsonResp(w, map[string]string{"html": h}, err)
}

func getText(w http.ResponseWriter, r *http.Request) {
	t, err := br.GetText()
	jsonResp(w, map[string]string{"text": t}, err)
}

func back(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, nil, br.NavigateBack())
}

func forward(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, nil, br.NavigateForward())
}

func refresh(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, nil, br.Refresh())
}

func screenshot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	jsonResp(w, nil, br.TakeScreenshot(req.Path))
}
