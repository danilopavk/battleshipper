package engine

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_InitializeGame(t *testing.T) {
	playerA, playerB, game := initialize()
	if playerA.Id == 0 {
		t.Error("No player a")
	}

	if playerB.Id == 0 {
		t.Error("No player b")
	}

	if game.Id == 0 {
		t.Error("No game")
	}
}

func Test_FillBoard(t *testing.T) {
	player, _, game := initialize()

	for i := 0; i < 5; i++ {
		nextShipLength, err := game.NextShipLength(player.Id)
		if err != nil {
			t.Error("Could not get next ship length")
		}

		cells := map[Cell]bool{}
		for j := 0; j < nextShipLength; j++ {
			cells[Cell{i * 2, j}] = true
		}
		error := player.AddShip(Ship{cells})
		if error != nil {
			t.Errorf("Cannot add ship, %v", error)
		}
	}

	nextShipLength, err := game.NextShipLength(player.Id)
	if err == nil {
		t.Errorf("Expected to throw error on getting next ship, but it did not happen, instead got: %d", nextShipLength)
	}
}

func Test_AvailableCells(t *testing.T) {
	player, _, _ := initializeAndStart()
	availableCells := player.AvailableCells()

	expectedAvailableCells := map[int]int{
		0: 4,
		1: 5,
		2: 5,
		3: 6,
		4: 5,
		5: 6,
		6: 6,
		7: 7,
		8: 6,
		9: 7,
	}

	for cellIndex, expectedAvailable := range expectedAvailableCells {
		actuallyAvailable := 0
		available := availableCells[cellIndex]
		for _, exists := range available {
			if exists {
				actuallyAvailable++
			}
		}
		if diff := cmp.Diff(actuallyAvailable, expectedAvailable); diff != "" {
			t.Errorf("unexpected diff: %v", diff)
		}
	}
}

func Test_ShootAndHit(t *testing.T) {
	player, _, game := initializeAndStart()
	hit, sank, won, err := game.Shoot(player.Id, Cell{0, 0})

	if !hit {
		t.Error("Expected to hit, but it didn't")
	}

	if won {
		t.Error("Expected not to win, but id did")
	}

	if err != nil {
		t.Error(fmt.Errorf("Expected to hit, but err returned: %v", err))
	}

	hits := player.Target.Hits

	if diff := cmp.Diff(map[Cell]bool{{0, 0}: true}, hits); diff != "" {
		t.Errorf("Unexpected diff %v", diff)
	}

	if sank {
		t.Error("Expected not to sink the ship, but it did")
	}
}

func Test_ShootAndMiss(t *testing.T) {
	player, _, game := initializeAndStart()
	hit, sank, won, err := game.Shoot(player.Id, Cell{9, 9})

	if hit {
		t.Error("Expected to miss, but it didn't")
	}

	if sank {
		t.Error("Expected not to sink the ship, but it did")
	}

	if won {
		t.Error("Expected not to win, but it did")
	}

	if err != nil {
		t.Error(fmt.Errorf("Expected to miss, but err returned: %v", err))
	}

	misses := player.Target.Misses
	if diff := cmp.Diff(map[Cell]bool{{9, 9}: true}, misses); diff != "" {
		t.Errorf("Unexpected diff %v", diff)
	}
}

func Test_ShootWithoutStart(t *testing.T) {
	playerA, _, game := initialize()
	_, _, _, err := game.Shoot(playerA.Id, Cell{0, 0})

	if err == nil {
		t.Error("Shooting without start expected to return error")
	}
}

func Test_ShootOutOfTurn(t *testing.T) {
	_, player, game := initializeAndStart()
	_, _, _, err := game.Shoot(player.Id, Cell{0, 0})

	if err == nil {
		t.Error("Player shot out of turn, but there was no error")
	}
}

func Test_ShotInWrongGame(t *testing.T) {
	_, _, game := initializeAndStart()
	player, _, _ := initializeAndStart()

	_, _, _, err := game.Shoot(player.Id, Cell{0, 0})

	if err == nil {
		t.Error("Player shot in the wrong game, but there was no error")
	}
}

func Test_ShootAndSink(t *testing.T) {
	player, _, game := initializeAndStart()
	player.Target.Hits = map[Cell]bool{
		{0, 0}: true,
		{0, 1}: true,
		{0, 2}: true,
		{0, 3}: true,
	}

	hit, sank, won, err := game.Shoot(player.Id, Cell{0, 4})

	if !hit {
		t.Error("Expected to hit with 0, 4 but didn't")
	}

	if !sank {
		t.Error("Expected to sink the ship, but it did not")
	}

	if won {
		t.Error("Expected not to win, but it did")
	}

	if err != nil {
		t.Error(fmt.Errorf("Unexpected error, %v", err))
	}

	expectedSankShips := []Ship{
		{
			Cells: map[Cell]bool{
				{0, 0}: true,
				{0, 1}: true,
				{0, 2}: true,
				{0, 3}: true,
				{0, 4}: true,
			},
		},
	}
	if diff := cmp.Diff(expectedSankShips, *player.Target.SankShips); diff != "" {
		t.Errorf("Unexpected diff: %v", diff)
	}

	if len(player.Target.Hits) == 0 {
		t.Error("Expected to cleanup Hits object, but it didn't")
	}

	expectedMisses := map[Cell]bool{
		{1, 0}: true,
		{1, 1}: true,
		{1, 2}: true,
		{1, 3}: true,
		{1, 4}: true,
		{1, 5}: true,
		{0, 5}: true,
	}
	if diff := cmp.Diff(expectedMisses, player.Target.Misses); diff != "" {
		t.Error(fmt.Errorf("Expected ship sinking to bring 7 misses in neighbor cells, but there are %v", player.Target.Misses))
	}
}

func Test_ShootAndWin(t *testing.T) {
	player, _, game := initializeAndStart()
	player.Target.Hits = map[Cell]bool{
		{0, 0}: true,
		{0, 1}: true,
		{0, 2}: true,
		{0, 3}: true,
	}
	player.Target.SankShips = &[]Ship{
		{map[Cell]bool{}},
		{map[Cell]bool{}},
		{map[Cell]bool{}},
		{map[Cell]bool{}},
	}

	hit, sank, won, err := game.Shoot(player.Id, Cell{0, 4})

	if !hit {
		t.Error("Expected to hit with 0, 4 but didn't")
	}
	if !sank {
		t.Error("Expected to sink the ship, but it did not")
	}
	if !won {
		t.Error("expected to win, but it did not!")
	}
	if err != nil {
		t.Error(fmt.Errorf("Unexpected error, %v", err))
	}
	if len(*player.Target.SankShips) != 5 {
		t.Error("Expected to have a sank ship, but there isn't")
	}

	if *game.Winner != player.Id {
		t.Error(fmt.Errorf("Expected player %d to be the winner, but it was %d", player.Id, *game.Winner))
	}
}

func initializeAndStart() (Player, Player, Game) {
	playerA, playerB, game := initialize()
	playerA = fill(playerA, game)
	playerB = fill(playerB, game)

	return playerA, playerB, game
}

func fill(player Player, game Game) Player {
	for i := 0; i < 5; i++ {
		nextShipLength, _ := game.NextShipLength(player.Id)

		cells := map[Cell]bool{}
		for j := 0; j < nextShipLength; j++ {
			cells[Cell{i * 2, j}] = true
		}
		_ = player.AddShip(Ship{cells})
	}
	return player
}

func initialize() (Player, Player, Game) {
	playerA := InitializePlayer("Anomander")
	playerB := InitializePlayer("Whiskeyjack")
	game := InitializeGame(playerA, playerB, playerA.Id)
	return playerA, playerB, game
}
