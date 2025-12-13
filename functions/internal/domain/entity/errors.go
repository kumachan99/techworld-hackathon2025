package entity

import "errors"

var (
	// Room errors
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomFull           = errors.New("room is full")
	ErrGameAlreadyStarted = errors.New("game has already started")
	ErrCannotStartGame    = errors.New("cannot start game")
	ErrInvalidPhase       = errors.New("invalid game phase")
	ErrNotEnoughPlayers   = errors.New("not enough players to start")
	ErrNotAllVoted        = errors.New("not all players have voted")
	ErrNotAllReady        = errors.New("not all players are ready")

	// Player errors
	ErrPlayerNotFound      = errors.New("player not found")
	ErrPlayerNotInRoom     = errors.New("player is not in this room")
	ErrPlayerAlreadyInRoom = errors.New("player is already in this room")
	ErrAlreadyVoted        = errors.New("player has already voted")
	ErrPetitionUsed        = errors.New("petition has already been used")

	// Policy errors
	ErrPolicyNotFound = errors.New("policy not found")
	ErrInvalidPolicy  = errors.New("invalid policy")

	// Ideology errors
	ErrNoIdeologyAvailable = errors.New("no ideology available")

	// Authorization errors
	ErrNotHost = errors.New("only host can perform this action")

	// AI errors
	ErrPetitionRejected = errors.New("petition was rejected by AI")
)
