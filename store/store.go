package store

import (
	"fmt"
	"sync"

	"github.com/danilopavk/battleshipper/engine"
)

type Store struct {
	mutex            sync.RWMutex
	GamesByGameId    map[int]engine.Game
	GameIdByPlayerId map[int]int
	WaitingPlayers   map[int]engine.Player
}

func InitializeStore() Store {
	return Store{
		GamesByGameId:    map[int]engine.Game{},
		GameIdByPlayerId: map[int]int{},
		WaitingPlayers:   map[int]engine.Player{},
	}
}

func (store *Store) StartGame(playerName string) engine.Player {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	player := engine.InitializePlayer(playerName)
	store.WaitingPlayers[player.Id] = player

	return player
}

func (store *Store) AllWaitingPlayers() []engine.Player {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	players := []engine.Player{}
	// return deep copy of the players list
	for _, player := range store.WaitingPlayers {
		// while waiting for the game, target will always be empty
		target := engine.Target{
			SankShips: &[]engine.Ship{},
			Hits:      map[engine.Cell]bool{},
			Misses:    map[engine.Cell]bool{},
		}
		// deep copy current ships
		ships := []engine.Ship{}
		ships = append(ships, *player.Ships...)
		players = append(
			players,
			engine.Player{Id: player.Id, Name: player.Name, Ships: &ships, Target: &target},
		)
	}

	return players
}

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

func (store *Store) UpdateGame(game engine.Game) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.GamesByGameId[game.Id]; !ok {
		return fmt.Errorf("Cannot find game with id %d", game.Id)
	}

	store.GamesByGameId[game.Id] = game

	return nil
}
