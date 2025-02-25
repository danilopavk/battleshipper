// Package that represents an in-memory store for battlershipper game.
//
// It tracks all the games in memory, and performs the basic management operations
// for a game, like start game, join game, retrieve and update.
package store

import (
	"fmt"
	"sync"

	"github.com/danilopavk/battleshipper/engine"
)

// Store type stores all the data related to the game.
//
// Has one lock for both pending and running games, so potentially we could
// separate this into 2 structs if we see that locks are slowing
// down the game, so at least one of the game phases are spared.
type Store struct {
	mutex            sync.RWMutex
	GamesByGameId    map[int]engine.Game
	GameIdByPlayerId map[int]int
	WaitingPlayers   map[int]engine.Player
}

// InitializeStore builds the empty store
func InitializeStore() Store {
	return Store{
		GamesByGameId:    map[int]engine.Game{},
		GameIdByPlayerId: map[int]int{},
		WaitingPlayers:   map[int]engine.Player{},
	}
}

// StartGame starts a new game.
//
// Adds the player to waiting players map,
// where they will wait for someone to join their game.
func (store *Store) StartGame(playerName string) engine.Player {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	player := engine.InitializePlayer(playerName)
	store.WaitingPlayers[player.Id] = player

	return player
}

// AllWaitingPlayers method returns a list of all waiting players.
//
// Should be used to offer list of
// potential players whom games one might join
func (store *Store) AllWaitingPlayers() []engine.Player {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	var players []engine.Player
	for _, player := range store.WaitingPlayers {
		players = append(players, player)
	}

	return players
}

// GetPlayerAndGame method retrieves a player and the corresponding game by player id.

// If player is in a waiting state,game will be nil.
func (store *Store) GetPlayerAndGame(playerId int) (engine.Player, engine.Game, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if waitingPlayer, ok := store.WaitingPlayers[playerId]; ok {
		return waitingPlayer, engine.Game{}, nil
	}

	if gameId, ok := store.GameIdByPlayerId[playerId]; ok {
		if game, ok := store.GamesByGameId[gameId]; ok {
			switch playerId {
			case game.PlayerA.Id:
				return game.PlayerA, game, nil
			case game.PlayerB.Id:
				return game.PlayerB, game, nil
			default:
				return engine.Player{}, engine.Game{}, fmt.Errorf("Internal error, game found for player id %d, but it doesn't contain the same player id", playerId)
			}
		}
		return engine.Player{}, engine.Game{}, fmt.Errorf("Internal error, game id found for player id %d, but can't find the game with that id: %d", playerId, gameId)
	}

	return engine.Player{}, engine.Game{}, fmt.Errorf("Game not found for player %d", playerId)
}

// UpdatePlayer updates a player in the db.

// It can be either a waiting player, or one in the active game.
func (store *Store) UpdatePlayer(player engine.Player) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.WaitingPlayers[player.Id]; ok {
		store.WaitingPlayers[player.Id] = player
		return nil
	}

	if gameId, ok := store.GameIdByPlayerId[player.Id]; ok {
		if game, ok := store.GamesByGameId[gameId]; ok {
			switch player.Id {
			case game.PlayerA.Id:
				{
					game.PlayerA = player
					return nil
				}
			case game.PlayerB.Id:
				{
					game.PlayerB = player
					return nil
				}
			}
		}
	}
	return fmt.Errorf("Not found player with id %d to update", player.Id)
}

// JoinGame joins a game that the opponent already started
func (store *Store) JoinGame(playerName string, opponentId int) engine.Game {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	playerA := store.WaitingPlayers[opponentId]
	playerB := engine.InitializePlayer(playerName)

	game := engine.InitializeGame(playerA, playerB, playerA.Id)

	delete(store.WaitingPlayers, playerA.Id)
	store.GameIdByPlayerId[playerA.Id] = game.Id
	store.GameIdByPlayerId[playerB.Id] = game.Id
	store.GamesByGameId[game.Id] = game

	return game
}

// UpdateGame updates a game.

// It can update anything about the game - player data, or some game metadata
func (store *Store) UpdateGame(game engine.Game) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.GamesByGameId[game.Id]; !ok {
		return fmt.Errorf("Cannot find game with id %d", game.Id)
	}

	store.GamesByGameId[game.Id] = game

	return nil
}
