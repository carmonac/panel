package main

import (
	"bufio"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type anyError struct {
	Message string
}

type infoPage struct {
	User    string
	Content []contentPage
}

type contentPage struct {
	Title      string
	Id         string
	Content    string
	Author     string
	Dateformat string
	Editable   bool
	Img        string
}

func Auth(session sessions.Session, rw http.ResponseWriter) {
	if session.Get("admin") == nil {
		http.Error(rw, "Not Authorized", http.StatusUnauthorized)
	}
}

func homeAccess(r render.Render, session sessions.Session, rw http.ResponseWriter, req *http.Request) {
	if session.Get("admin") == nil {
		r.HTML(200, "index", anyError{""})
	} else {
		http.Redirect(rw, req, "/dash", http.StatusMovedPermanently)
	}
}

func login(session sessions.Session, r render.Render, rw http.ResponseWriter, req *http.Request) {

	body, _ := ioutil.ReadAll(req.Body)
	v, _ := url.ParseQuery(string(body))

	file, err := os.Open("./access.pass")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	passwords := make(map[string]string)

	for scanner.Scan() {
		s := strings.Split(scanner.Text(), " ")
		passwords[s[0]] = s[1]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	data := []byte(v.Get("password"))
	passwordByte := md5.Sum(data)

	if passwords[v.Get("usuario")] == fmt.Sprintf("%x", passwordByte) {
		session.Set("admin", v.Get("usuario"))
		http.Redirect(rw, req, "/dash", http.StatusMovedPermanently)
	} else {
		r.HTML(200, "index", anyError{"Usuario o contraseña inválida"})
	}
}

func enterDash(session sessions.Session, r render.Render, rw http.ResponseWriter) {
	r.HTML(200, "control", nil)
}

func enterEntries(session sessions.Session, r render.Render, rw http.ResponseWriter) {
	db, err := sql.Open("sqlite3", *dbfile)
	rows, err := db.Query("select id, author, title, content, dateformat from entries order by timestamp desc")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	e := make([]contentPage, 0)
	for rows.Next() {
		var title, author, content, id, dateformat string
		var editable bool
		rows.Scan(&id, &author, &title, &content, &dateformat)
		if author == session.Get("admin") {
			editable = true
		} else {
			editable = false
		}
		e = append(e, contentPage{title, id, content, author, dateformat, editable, "nil"})
	}
	rows.Close()
	//	i := infoPage{}
	//	i.content = e
	//	i.user = fmt.Sprintf("%s", session.Get("admin"))
	r.HTML(200, "entries", infoPage{fmt.Sprintf("%s", session.Get("admin")), e})

}

func enterNewEntry(session sessions.Session, r render.Render, rw http.ResponseWriter) {
	r.HTML(200, "newentry", nil)
}

func saveEntry(session sessions.Session, r render.Render, rw http.ResponseWriter, req *http.Request) {

	body, _ := ioutil.ReadAll(req.Body)
	v, _ := url.ParseQuery(string(body))
	log.Println("Saving post with id ", v.Get("identry"))

	db, err := sql.Open("sqlite3", *dbfile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	t := time.Now().Local()
	sqlStmt := fmt.Sprintf("INSERT INTO entries (id_name, author, title, content, timestamp, dateformat, img) VALUES ('%s', '%s', '%s', '%s', %s, '%s', '%s')", strings.Replace(v.Get("identry"), " ", "-", -1), session.Get("admin"), v.Get("title"), v.Get("content"), t.Format("20060102150405"), t.Format("02/01/2006"), v.Get("img"))
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
		return
	}
	http.Redirect(rw, req, "/entries", http.StatusMovedPermanently)

}

func enterEdit(session sessions.Session, r render.Render, rw http.ResponseWriter, req *http.Request, params martini.Params) {
	db, err := sql.Open("sqlite3", *dbfile)
	query := fmt.Sprintf("select id_name, author, title, content, img from entries where id_name = '%s'", params["id"])
	row, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	//var author, title, content string
	c := contentPage{}
	row.Next()
	row.Scan(&c.Id, &c.Author, &c.Title, &c.Content, &c.Img)
	c.Id = params["id"]
	if c.Author != session.Get("admin") {
		http.Error(rw, "Not Authorized", http.StatusUnauthorized)
	} else {
		r.HTML(200, "edit", c)
	}
}

func editPost(rw http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	v, _ := url.ParseQuery(string(body))
	query := fmt.Sprintf("update entries set author = '%s', title = '%s', content = '%s', img = '%s'", v.Get("author"), v.Get("title"), v.Get("content"), v.Get("img"))
	db, err := sql.Open("sqlite3", *dbfile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("%q: %s\n", err, query)
		return
	}
	http.Redirect(rw, req, "/entries", http.StatusMovedPermanently)
}
