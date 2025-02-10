package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/danilopavk/battleshipper/home"
	"github.com/danilopavk/battleshipper/store"
)

func main() {
	gameStore := store.InitializeStore()
	homePage := home.Page()
	http.Handle("/", templ.Handler(homePage))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/start", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			var startPlayer StartPlayer
			err := json.NewDecoder(request.Body).Decode(&startPlayer)
			if err != nil {
				fmt.Printf("Cannot decode player name, error: %v", err)
				return
			}

			player := gameStore.StartGame(startPlayer.Name)
			fmt.Printf("Started a game with player %+v\n", player)
		}
	})

	if err := http.ListenAndServe(":3002", nil); err != nil {
		panic(fmt.Sprintf("Cannot start server, cause: %v", err))
	}
}

type StartPlayer struct {
	Name string `json:"name"`
}
