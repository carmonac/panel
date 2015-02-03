// panel project main.go
package main

import (
	"flag"
	"fmt"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

var (
	dbfile *string
)

func main() {
	// Lista de variables de configuración
	puerto := flag.String("puerto", "8081", "Selecciona el puerto de respuesta del panel")
	dbpath := flag.String("dbpath", "./database.db", "Indica la ruta a la base de datos")
	users := flag.String("users", "./access.pass", "Indica la ruta a la base de datos")

	flag.Parse()
	dbfile = dbpath
	fmt.Printf("Escuchando en puerto: %s\nCargando usuarios de: %s\nCargando SQLite: %s\n\n", *puerto, *users, *dbpath)

	m := martini.Classic()
	m.Use(gzip.All())

	store := sessions.NewCookieStore([]byte("golangninjasession"))

	m.Use(sessions.Sessions("dash_session", store))

	m.Use(render.Renderer(render.Options{
		Extensions: []string{".tmpl", ".html"},
	}))

	m.Get("/dash", Auth, enterDash)
	m.Get("/entries", Auth, enterEntries)
	m.Get("/newentry", Auth, enterNewEntry)
	m.Get("/edit/:id", Auth, enterEdit)
	m.Get("/", homeAccess)
	m.Post("/login", login)
	m.Post("/saveentry", saveEntry)
	m.Post("/editsave/:id", editPost)

	m.RunOnAddr(":" + *puerto)
}
