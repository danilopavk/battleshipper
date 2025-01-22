// Package engine provides a way to manage the game of battleship.
// It stores the game in the object and allows basic operations on it.
package engine

import (
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
)

const width = 10
const height = 10

// Game is an object holding the whole data related to a single game.
// Two objects representing two players, Turn int representing an id
// of the player whose turn it is, and Winner int representing an id
// of the winning player.
type Game struct {
	Id               int
	PlayerA, PlayerB Player
	Turn             *int
	Winner           *int
}

// Player type holds the data on one contestant of the game. Ships
// pointer represents ships that belong to the player. Target pointer
// represents all the data that this player has on the opposing player's
// board
type Player struct {
	Id     int
	Name   string
	Ships  *[]Ship
	Target *Target
}

// Target type holds the data that one player has on the opposing player's board.
// SankShips object contains ships that are completely sank. Hits contain hits
// on ships that aren't completely sunk yet. Misses contains cells where this player
// knows that are empty - either by direct hits or because they are adjacent to a sank ship
type Target struct {
	SankShips *[]Ship
	Hits      map[Cell]bool
	Misses    map[Cell]bool
}

// Ship is an object containing a collection of cells
type Ship struct {
	Cells map[Cell]bool
}

// Cell is one item in a grid
type Cell struct {
	X, Y int
}

// InitializeGame sets up the game between 2 players. To create the player,
// call InitializePlayer method
func InitializeGame(playerA, playerB Player, turn int) Game {
	return Game{rand.Int(), playerA, playerB, &turn, nil}
}

// NextShipLength is a function used in the first phase of the game, while
// players are still creating the ships. It's used to instruct players how
// large the next ship should be
func (game Game) NextShipLength(playerId int) (int, error) {
	switch playerId {
	case game.PlayerA.Id:
		return game.PlayerA.nextShipLength()
	case game.PlayerB.Id:
		return game.PlayerB.nextShipLength()
	default:
		return -1, fmt.Errorf("Player with id %d not found", playerId)
	}
}

// Shoot method is a function used in the second phase of the game. Player whose turn it is
// tries to guess where the ships are. The function returns 4 values: 1) Weather the ship was hit 2)
// Weather this hit sank the ship 3) Weather this sank ship means that the player won the game 4)
// Error thrown if the shot is illegal. The shot is illegal if 1) The shooting phase of the game didn't
// start yet 2) It's not player's turn 3) It's not even player's game
func (game *Game) Shoot(playerId int, cell Cell) (hit bool, sank bool, won bool, err error) {
	if len(*game.PlayerA.Ships) != 5 {
		return false, false, false, fmt.Errorf("Shot attempted, but player %d does not have his board full yet", game.PlayerA.Id)
	}
	if len(*game.PlayerB.Ships) != 5 {
		return false, false, false, fmt.Errorf("Shot attempted, but player %d does not have his board full yet", game.PlayerB.Id)
	}
	if *game.Turn != playerId {
		return false, false, false, fmt.Errorf("Player %d tried to shoot, but it's not their turn", playerId)
	}

	var player Player
	var opponent Player

	switch playerId {
	case game.PlayerA.Id:
		{
			player = game.PlayerA
			opponent = game.PlayerB
			*game.Turn = opponent.Id
		}
	case game.PlayerB.Id:
		{
			player = game.PlayerB
			opponent = game.PlayerB
			*game.Turn = opponent.Id

		}
	default:
		return false, false, false, fmt.Errorf("Player %d not in game %d", playerId, game.Id)
	}

	hit = opponent.shoot(cell)
	if !hit {
		player.Target.Misses[cell] = true
		return hit, false, false, nil
	}
	player.Target.Hits[cell] = true

	ship := player.shipFromHits(cell)
	if !opponent.sank(ship) {
		return hit, false, false, nil
	}
	player.markAsSank(ship)

	if len(*player.Target.SankShips) != 5 {
		return hit, true, false, nil
	}
	game.Winner = &player.Id
	return hit, true, true, nil
}

// Initializes the contestant with the given name and return the player object
func InitializePlayer(name string) Player {
	target := Target{&[]Ship{}, map[Cell]bool{}, map[Cell]bool{}}
	return Player{rand.Int(), name, &[]Ship{}, &target}
}

// AvailableCells method gives a utility method that can be used to draw the board
// for the player.
func (player Player) AvailableCells() map[int]map[int]bool {
	cells := map[int]map[int]bool{}
	for x := 0; x < width; x++ {
		cells[x] = map[int]bool{}
		for y := 0; y < height; y++ {
			cells[x][y] = true
		}
	}

	for _, ship := range *player.Ships {
		for cell := range ship.Cells {
			cells[cell.X][cell.Y] = false
			filter := func(cell Cell) bool { return true }
			for _, neighbor := range neighborCells(cell, false, filter) {
				cells[neighbor.X][neighbor.Y] = false
			}
		}
	}
	return cells
}

// AddShip method is a method used in the initializing phase of the game. It allows
// the player to add the ship to their method. Returns error if the provided ship isn't
// of the correct length
func (player Player) AddShip(ship Ship) error {
	nextShipLength, err := player.nextShipLength()
	if err != nil {
		return err
	}
	if len(ship.Cells) != nextShipLength {
		return fmt.Errorf("Cannot add ship, expected length %d, was %d", nextShipLength, len(ship.Cells))
	}

	availableCells := player.AvailableCells()
	for cell := range ship.Cells {
		if !availableCells[cell.X][cell.Y] {
			return fmt.Errorf("Cannot add ship, cell not available %d - %d", cell.X, cell.Y)
		}
	}

	*player.Ships = append(*player.Ships, ship)

	return nil
}

func (player Player) markAsSank(ship Ship) {
	*player.Target.SankShips = append(*player.Target.SankShips, ship)
	notHits := func(cell Cell) bool {
		return !player.Target.Hits[cell]
	}
	for shipCell := range ship.Cells {
		for _, neighbor := range neighborCells(shipCell, true, notHits) {
			player.Target.Misses[neighbor] = true
		}
	}
	for cell := range ship.Cells {
		player.Target.Hits[cell] = false
	}
}

func (player Player) shoot(shootAtCell Cell) bool {
	for _, ship := range *player.Ships {
		for cell := range ship.Cells {
			if cell == shootAtCell {
				return true
			}
		}
	}

	return false
}

func (player Player) nextShipLength() (int, error) {
	switch len(*player.Ships) {
	case 0:
		return 5, nil
	case 1, 2:
		return 4, nil
	case 3, 4:
		return 3, nil
	default:
		return -1, errors.New("Player board is full, cannot create new ship!")
	}
}

func neighborCells(originCell Cell, includeDiagonal bool, filter func(Cell) bool) []Cell {
	cells := []Cell{}

	if originCell.X > 0 {
		neighbor := Cell{originCell.X - 1, originCell.Y}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	if originCell.X < width-1 {
		neighbor := Cell{originCell.X + 1, originCell.Y}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	if originCell.Y > 0 {
		neighbor := Cell{originCell.X, originCell.Y - 1}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	if originCell.Y < height-1 {
		neighbor := Cell{originCell.X, originCell.Y + 1}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	if !includeDiagonal {
		return cells
	}

	if originCell.X > 0 && originCell.Y > 0 {
		neighbor := Cell{originCell.X - 1, originCell.Y - 1}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}

	}

	if originCell.X < width-1 && originCell.Y > 0 {
		neighbor := Cell{originCell.X + 1, originCell.Y - 1}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	if originCell.X < width-1 && originCell.Y < height-1 {
		neighbor := Cell{originCell.X + 1, originCell.Y + 1}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	if originCell.X > 0 && originCell.Y < height-1 {
		neighbor := Cell{originCell.X - 1, originCell.Y + 1}
		if filter(neighbor) {
			cells = append(cells, neighbor)
		}
	}

	return cells
}

func (player Player) shipFromHits(cell Cell) Ship {
	target := player.Target
	shipCells := map[Cell]bool{}
	queue := []Cell{cell}

	for len(queue) > 0 {

		next := queue[0]
		queue = queue[1:]

		if shipCells[next] {
			continue
		}

		shipCells[next] = true

		filter := func(cell Cell) bool { return target.Hits[cell] }
		queue = append(queue, neighborCells(next, false, filter)...)
	}

	cells := map[Cell]bool{}
	for cell := range shipCells {
		cells[cell] = true
	}
	return Ship{cells}

}

func (player Player) sank(potentialShip Ship) bool {
	for _, ship := range *player.Ships {
		if maps.Equal(ship.Cells, potentialShip.Cells) {
			return true
		}
	}

	return false
}
