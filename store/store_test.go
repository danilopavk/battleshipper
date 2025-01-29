package store

import (
	"testing"

	"github.com/danilopavk/battleshipper/engine"
	"github.com/google/go-cmp/cmp"
)

func Test_StartGame(t *testing.T) {
	store := InitializeStore()

	player := store.StartGame("Karsa Orlong")

	if player.Name != "Karsa Orlong" {
		t.Error("Wrong player name!")
	}
}

func Test_AllWaitingPlayers(t *testing.T) {
	store := InitializeStore()

	store.StartGame("Karsa Orlong")
	store.StartGame("Fiddler")
	store.StartGame("Coltaine")

	allPlayers := store.AllWaitingPlayers()

	if len(allPlayers) != 3 {
		t.Errorf("Expected to find 3 players, but found %d", len(allPlayers))
	}

	players := map[string]bool{
		"Karsa Orlong": true,
		"Fiddler":      true,
		"Coltaine":     true,
	}

	if _, exists := players[allPlayers[0].Name]; !exists {
		t.Errorf("Unexpected player name %v", allPlayers[0].Name)
	}
	delete(players, allPlayers[0].Name)

	if _, exists := players[allPlayers[1].Name]; !exists {
		t.Errorf("Unexpected player name %v", allPlayers[1].Name)
	}
	delete(players, allPlayers[1].Name)

	if _, exists := players[allPlayers[2].Name]; !exists {
		t.Errorf("Unexpected player name %v", allPlayers[1].Name)
	}
	delete(players, allPlayers[2].Name)

}

func Test_GetWaitingPlayer(t *testing.T) {
	store := InitializeStore()

	player := store.StartGame("Karsa Orlong")

	savedPlayer, game, error := store.GetPlayerAndGame(player.Id)

	if savedPlayer.Name != "Karsa Orlong" {
		t.Errorf("Unexpected player name %v", savedPlayer.Name)
	}

	if game.Id != 0 {
		t.Errorf("Expected to return empty game, but returned %v", game)
	}

	if error != nil {
		t.Errorf("Expected not to throw erorr, bit it was %v", error)
	}
}

func Test_GetPlayerAndGame(t *testing.T) {
	store := InitializeStore()

	playerA := store.StartGame("Karsa Orlong")
	startedGame := store.JoinGame("Fiddler", playerA.Id)

	karsa, game, error := store.GetPlayerAndGame(playerA.Id)

	if diff := cmp.Diff(karsa, game.PlayerA); diff != "" {
		t.Errorf("Unexpected change in karsa: %v", diff)
	}

	if game.Id != startedGame.Id {
		t.Errorf("Unexpected game id: %d, expected: %d", game.Id, startedGame.Id)
	}

	if error != nil {
		t.Errorf("Unexpected error: %c", error)
	}
}

func Test_GetUnknownPlayer(t *testing.T) {
	store := InitializeStore()

	player, game, error := store.GetPlayerAndGame(1)

	if player.Id != 0 {
		t.Errorf("Expected to find empty player, but found %v", player)
	}

	if game.Id != 0 {
		t.Errorf("Expected to find empty game, but found %v", player)
	}

	if error == nil {
		t.Errorf("expected to find error on getting unknown player, but it was nil")
	}
}

func Test_UpdatePendingPlayer(t *testing.T) {
	store := InitializeStore()

	player := store.StartGame("Karsa Orlong")

	ship := engine.Ship{
		Cells: map[engine.Cell]bool{
			{X: 0, Y: 0}: true,
			{X: 0, Y: 1}: true,
			{X: 0, Y: 2}: true,
			{X: 0, Y: 3}: true,
			{X: 0, Y: 4}: true,
		},
	}
	_ = player.AddShip(ship)

	error := store.UpdatePlayer(player)

	if error != nil {
		t.Errorf("Unexpected error: %v", error)
	}

	result, _, _ := store.GetPlayerAndGame(player.Id)

	if diff := cmp.Diff(player, result); diff != "" {
		t.Errorf("Unexpected diff on updated player: %v", diff)
	}
}

func Test_UpdatePlayerInGame(t *testing.T) {
	store := InitializeStore()

	karsa := store.StartGame("Karsa Orlong")
	game := store.JoinGame("Fiddler", karsa.Id)
	fiddler := game.PlayerB

	ship := engine.Ship{
		Cells: map[engine.Cell]bool{
			{X: 0, Y: 0}: true,
			{X: 0, Y: 1}: true,
			{X: 0, Y: 2}: true,
			{X: 0, Y: 3}: true,
			{X: 0, Y: 4}: true,
		},
	}

	_ = karsa.AddShip(ship)

	error := store.UpdatePlayer(karsa)

	if error != nil {
		t.Errorf("Unexpected error: %v", error)
	}

	updatedKarsa, _, _ := store.GetPlayerAndGame(karsa.Id)

	if diff := cmp.Diff(updatedKarsa, karsa); diff != "" {
		t.Errorf("Unexpected diff %v", diff)
	}

	fiddlerShip := engine.Ship{
		Cells: map[engine.Cell]bool{
			{X: 0, Y: 0}: true,
			{X: 0, Y: 1}: true,
			{X: 0, Y: 2}: true,
			{X: 0, Y: 3}: true,
			{X: 0, Y: 4}: true,
		},
	}
	_ = fiddler.AddShip(fiddlerShip)

	updatedFiddler, _, _ := store.GetPlayerAndGame(fiddler.Id)

	if diff := cmp.Diff(updatedFiddler, fiddler); diff != "" {
		t.Errorf("Unexpected diff %v", diff)
	}

}

func Test_JoinGame(t *testing.T) {
	store := InitializeStore()

	karsa := store.StartGame("Karsa Orlong")
	game := store.JoinGame("Fiddler", karsa.Id)

	if game.PlayerA.Id != karsa.Id {
		t.Error("Created the game, but joined with the wrong player!")
	}
	if game.PlayerB.Name != "Fiddler" {
		t.Error("Created the game, but player b is not fiddler!")
	}

	if _, exists := store.WaitingPlayers[karsa.Id]; exists {
		t.Error("Expected to remove karsa id from the map, but it's still there")
	}
	if store.GameIdByPlayerId[karsa.Id] != game.Id {
		t.Error("Cannot find game id by Karsa's id")
	}
	if store.GameIdByPlayerId[game.PlayerB.Id] != game.Id {
		t.Error("Cannot find game id by Fiddler's id")
	}
	if _, exists := store.GamesByGameId[game.Id]; !exists {
		t.Error("Cannot find game by its id")
	}
}

func Test_UpdateGame(t *testing.T) {
	store := InitializeStore()
	karsa := store.StartGame("KarsaOrlong")
	game := store.JoinGame("Fiddler", karsa.Id)
	fiddler := game.PlayerB

	ship := engine.Ship{
		Cells: map[engine.Cell]bool{
			{X: 0, Y: 0}: true,
			{X: 0, Y: 1}: true,
			{X: 0, Y: 2}: true,
			{X: 0, Y: 3}: true,
			{X: 0, Y: 4}: true,
		},
	}
	_ = fiddler.AddShip(ship)
	*game.Turn = fiddler.Id

	err := store.UpdateGame(game)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	updatedFiddler, updatedGame, _ := store.GetPlayerAndGame(fiddler.Id)

	if diff := cmp.Diff(updatedFiddler, fiddler); diff != "" {
		t.Errorf("Unexpected diff %v", diff)
	}
	if diff := cmp.Diff(updatedGame, game); diff != "" {
		t.Errorf("Unexpected diff %v", diff)
	}
}
