# Battleshipper game

## Repository overview

This repository is intended to contain the code necessary to run battleshipper game. For now, it contains:

 - Engine package containing basic game logic.
 - Store package containing in-memory store for the game state. Naive implementation, set up to enable initial testing.
 - Initial web setup: home page with basic instructions and a "start game" button that initializes the game for the player.

## Game description

Battleship is a 2-player board game. Each player has 2 grids in front of themselves: one where they have their ships, and one where they keep track of the opposing ships. The game contains of 2 phases: initialization and shooting. In the initialization phase, players are setting up their boards by adding ships to the hidden locations on the grid. Two ships cannot touch each other horizontally, vertically or diagonally. Once both players finished the initialization phase, shooting phase starts. Shooting phase is turn based. Each player chooses one cell on the grid and guesses if the opposing ship is there. If it is, the ship is hit. Once all the cells of one ship are hit, the ship is sunk. Once all ships are sunk, game is over and a player that has surviving ship(s) wins.
