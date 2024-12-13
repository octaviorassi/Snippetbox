package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Defino un handler para el home de la aplicacion.
// Supongo que cuando se conectan al home, les responde con el mensaje dado.
// Despues vemos que significan los argumentos, por ahora podemos pensar que w
// nos da los metodos para ensamblar una respuesta HTTP y enviarla al usuario, y
// r es un puntero a una estructura que tiene informacion del puntero actual.
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Snippetbox"))
}

func snippetView(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	msg := fmt.Sprintf("Display a specific snippet with ID %d...", id)
	w.Write([]byte(msg))


}

func snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a form for creating a new snippet..."))
}
// 
func main() {

	// Inicializo un serveMux, es decir, un router que mappea URLs a sus handlers.
	mux := http.NewServeMux()

	// Le digo que para cualquier url que comience con "/", su handler es home
	// ! OBS. Si pongo "/" lo toma como un comodin, si especifico algo mas ya solo matchea con esa URL.
	mux.HandleFunc("/{$}", home)
	mux.HandleFunc("/snippet/view/{id}", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)


	// Simplemente loggeo que se esta por iniciar el servidor en :4000
	log.Print("Starting server on :4000")

	// El metodo ListenAndServe inicia un nuevo servidor web. Le pasamos dos parametros,
	// la direccion TCP a la que escucha y el router (servermux) que definimos.
	// ! OBS. El formato del primer argumento es host:port, si omito host entonces el server va a
	// ! escuchar a todas las interfaces de mi computadora.
	err := http.ListenAndServe(":4000", mux)

	// ListenAndServe es bloqueante; si sale de su ejecucion entonces hay un error.
	log.Fatal(err)
}