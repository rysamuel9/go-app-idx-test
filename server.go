package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: helloserver [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	greeting = flag.String("g", "Hello", "Greet with `greeting`")
	addr     = flag.String("addr", "localhost:8080", "address to serve")
)

var (
	mu        sync.Mutex
	mahasiswa = make(map[string]bool)
)

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 0 {
		usage()
	}

	http.HandleFunc("/", greet)
	http.HandleFunc("/version", version)
	http.HandleFunc("/add", addMahasiswa)
	http.HandleFunc("/delete", deleteMahasiswa)

	log.Printf("serving http://%s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func version(w http.ResponseWriter, r *http.Request) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		http.Error(w, "no build information available", 500)
		return
	}

	fmt.Fprintf(w, "<!DOCTYPE html>\n<pre>\n")
	fmt.Fprintf(w, "%s\n", html.EscapeString(info.String()))
}

func greet(w http.ResponseWriter, r *http.Request) {
	name := strings.Trim(r.URL.Path, "/")
	if name == "" {
		name = "IDX"
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>Hello Server</title>
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
</head>
<body>
	<div class="container mt-5">
		<h1>%s, %s!</h1>
		<div class="mt-4">
			<form action="/add" method="post">
				<div class="form-group row">
					<label for="name" class="col-sm-2 col-form-label">Nama Mahasiswa:</label>
					<div class="col-sm-8">
						<input type="text" id="name" name="name" class="form-control">
					</div>
					<div class="col-sm-2">
						<button type="submit" class="btn btn-primary">Tambah</button>
					</div>
				</div>
			</form>
		</div>
		<div class="mt-4">
			<form action="/" method="get">
				<div class="form-group row">
					<label for="search" class="col-sm-2 col-form-label">Cari Mahasiswa:</label>
					<div class="col-sm-8">
						<input type="text" id="search" name="search" class="form-control">
					</div>
					<div class="col-sm-2">
						<button type="submit" class="btn btn-primary">Cari</button>
					</div>
				</div>
			</form>
		</div>
		<ul class="list-group mt-4">
	`, *greeting, html.EscapeString(name))

	mu.Lock()
	defer mu.Unlock()
	searchQuery := r.URL.Query().Get("search")
	for m := range mahasiswa {
		if searchQuery == "" || strings.Contains(strings.ToLower(m), strings.ToLower(searchQuery)) {
			fmt.Fprintf(w, `
				<li class="list-group-item d-flex justify-content-between align-items-center">
					%s
					<form action="/delete" method="post">
						<input type="hidden" name="name" value="%s">
						<button type="submit" class="btn btn-danger btn-sm">Delete</button>
					</form>
				</li>
			`, html.EscapeString(m), html.EscapeString(m))
		}
	}

	fmt.Fprintf(w, `</ul>
	</div>
</body>
</html>`)
}

func addMahasiswa(w http.ResponseWriter, r *http.Request) {
	// Hanya menerima metode POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := mahasiswa[name]; exists {
		http.Error(w, "Mahasiswa already exists", http.StatusBadRequest)
		return
	}

	mahasiswa[name] = true

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteMahasiswa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	delete(mahasiswa, name)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
