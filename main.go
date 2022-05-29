package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

// templateData provides template parameters.
type templateData struct {
	Service  string
	Revision string
}

// Variables used to generate the HTML page.
var (
	data templateData
	tmpl *template.Template
)

func getIP(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	//Get IP from RemoteAddr
	ipr, _, err := net.SplitHostPort(r.RemoteAddr)
	log.Printf("X-REAL, X-Forward, Remote: %s, %s, %s", ip, ips, ipr)
	netIP := net.ParseIP(ip)
	if netIP != nil {
		log.Printf("X-REAL-IP: %s", netIP)
		return ip, nil
	}
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			log.Printf("X-FORWADER-IP: %s", netIP)
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	//ipr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ipr)
	if netIP != nil {
		log.Printf("REMOTE-IP: %s", netIP)
		return ipr, nil
	}
	return "", fmt.Errorf("No valid ip found")
}

func main() {
	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "ISV Cloud on Air Demo"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "0.1.0"
	}

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}

	// Define HTTP server.
	http.HandleFunc("/", helloRunHandler)
	http.HandleFunc("/hostname", hostName)
	http.HandleFunc("/ip", whatisMyIP)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Print("Hello from Cloud Run! The container started successfully and is listening for HTTP requests on $PORT")
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// helloRunHandler responds to requests by rendering an HTML page.
func helloRunHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func hostName(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	w.WriteHeader(200)
	w.Write([]byte(hostname))
}

func whatisMyIP(w http.ResponseWriter, r *http.Request) {
	ip, err := getIP(r)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("No valid ip"))
	}
	w.WriteHeader(200)
	w.Write([]byte(ip))
}
