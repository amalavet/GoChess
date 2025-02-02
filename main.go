package main

import "fmt"

type (
	Player bool
	Board  map[Player]*Side
)

type Side struct {
	Goal    uint64
	Buckets uint64 // Each bucket uses 6 bits (0-63), total 36 bits used
}

var StartBoard = Board{
	PlayerA: {
		Goal:    0,
		Buckets: 0b000100_000100_000100_000100_000100_000100, // 6 buckets, each set to 4 (000100)
	},
	PlayerB: {
		Goal:    0,
		Buckets: 0b000100_000100_000100_000100_000100_000100, // 6 buckets, each set to 4 (000100)
	},
}

const (
	PlayerA  Player = true
	PlayerB  Player = false
	MaxDepth        = 14 // How deep the AI looks ahead
)

// Helper functions to get/set bucket values
func (s *Side) getBucket(i int) uint64 {
	shift := i * 6
	return (s.Buckets >> shift) & 0x3F // 0x3F = 111111 (6 bits)
}

func (s *Side) setBucket(i int, value uint64) {
	shift := i * 6
	// Clear the bucket's bits and set the new value
	mask := uint64(0x3F) << uint(shift) // 0x3F = 111111 (6 bits)
	s.Buckets = (s.Buckets & ^mask) | (uint64(value) << uint(shift))
}

func (b Board) display() {
	scoreA, scoreB := b.scores()
	fmt.Printf("\nScores - Player A: %d, Player B: %d\n\n", scoreA, scoreB)

	// Player B's buckets (reversed)
	fmt.Print("     ")
	for i := 5; i >= 0; i-- {
		fmt.Printf("[%d] ", b[PlayerB].getBucket(i))
	}
	fmt.Println()

	// Goals on both sides
	fmt.Printf("  %d                           %d\n", b[PlayerB].Goal, b[PlayerA].Goal)

	// Player A's buckets
	fmt.Print("     ")
	for i := 0; i < 6; i++ {
		fmt.Printf("[%d] ", b[PlayerA].getBucket(i))
	}
	fmt.Println("\n")
}

func main() {
	board := StartBoard
	currPlayer := PlayerA

	fmt.Println("Play against AI? (y/n)")
	var response string
	fmt.Scanln(&response)
	playAgainstAI := response == "y" || response == "Y"

	board.display()
	for !board.isEnd() {
		var bucket int
		if playAgainstAI && currPlayer == PlayerB {
			bucket = getAIMove(board, PlayerB)
			fmt.Printf("AI chooses bucket %d\n", bucket+1)
		} else {
			bucket = getBucketInput(currPlayer)
		}

		playAgain, err := board.move(currPlayer, bucket, nil)
		if err != nil {
			fmt.Printf("Invalid move: %v\n", err)
			continue
		}

		// Always display the board after a move
		board.display()

		if !playAgain {
			currPlayer = !currPlayer
		}
	}

	// Game over section
	fmt.Println("Game Over!")
	scoreA, scoreB := board.scores()
	fmt.Printf("Player A: %d, Player B: %d\n", scoreA, scoreB)
	if scoreA > scoreB {
		fmt.Println("Player A wins!")
	} else if scoreB > scoreA {
		fmt.Println("Player B wins!")
	} else {
		fmt.Println("It's a tie!")
	}
}

func getBucketInput(player Player) int {
	playerName := "Player A"
	if !player {
		playerName = "Player B"
	}

	for {
		fmt.Printf("%s, choose bucket (1-6): ", playerName)
		var input int
		_, err := fmt.Scanf("%d", &input)
		if err != nil {
			fmt.Println("Please enter a number between 1 and 6")
			fmt.Scanln()
			continue
		}

		input--
		if input >= 0 && input < 6 {
			return input
		}
		fmt.Println("Please enter a number between 1 and 6")
	}
}

func (b Board) move(player Player, bucket int, opts *MoveOptions) (bool, error) {
	if bucket > 5 {
		return false, fmt.Errorf("invalid bucket")
	}

	count := b[player].getBucket(bucket)
	if count == 0 {
		return false, fmt.Errorf("bucket is empty")
	}

	if opts == nil || !opts.silent {
		playerName := "Player A"
		if player == PlayerB {
			playerName = "Player B"
		}
		fmt.Printf("\n%s picks up %d stones from bucket %d\n", playerName, count, bucket+1)
	}

	b[player].setBucket(bucket, 0)
	actualPlayer := player

	for count > 0 {
		bucket++
		if bucket == 6 {
			if actualPlayer == player {
				b[player].Goal++
				count--
			}
			player = !player
			bucket = -1
			continue
		}
		count--
		b[player].setBucket(bucket, b[player].getBucket(bucket)+1)
	}

	return bucket == -1, nil
}

func (b Board) scores() (uint64, uint64) {
	// Start with goals
	scoreA := b[PlayerA].Goal
	scoreB := b[PlayerB].Goal

	// Add all bucket values using bit manipulation
	const mask uint64 = 0x3F // 111111 in binary

	// For each player's buckets:
	// 1. Extract each 6-bit group using shifts and masks
	// 2. Add them to the total
	for shift := 0; shift < 36; shift += 6 {
		scoreA += (b[PlayerA].Buckets >> shift) & mask
		scoreB += (b[PlayerB].Buckets >> shift) & mask
	}

	return scoreA, scoreB
}

func (b Board) isEnd() bool {
	return b[PlayerA].Buckets == 0 || b[PlayerB].Buckets == 0
}

// Add a helper function to count total stones in buckets
func (s *Side) totalStones() uint64 {
	var total uint64 = 0
	for i := 0; i < 6; i++ {
		total += s.getBucket(i)
	}
	return total
}

// Update MoveRecord to store delta operations
type MoveRecord struct {
	player        Player
	deltaGoalA    int64  // Change in goal A (can be negative)
	deltaGoalB    int64  // Change in goal B
	deltaBucketsA uint64 // XOR mask for player A's buckets
	deltaBucketsB uint64 // XOR mask for player B's buckets
}

func (b Board) recordMove(player Player) MoveRecord {
	return MoveRecord{
		player:        player,
		deltaGoalA:    -int64(b[PlayerA].Goal),
		deltaGoalB:    -int64(b[PlayerB].Goal),
		deltaBucketsA: b[PlayerA].Buckets,
		deltaBucketsB: b[PlayerB].Buckets,
	}
}

func (b Board) undoMove(record MoveRecord) {
	// Complete the delta calculations
	record.deltaGoalA += int64(b[PlayerA].Goal)
	record.deltaGoalB += int64(b[PlayerB].Goal)
	record.deltaBucketsA ^= b[PlayerA].Buckets
	record.deltaBucketsB ^= b[PlayerB].Buckets

	// Apply the deltas
	b[PlayerA].Goal = uint64(int64(b[PlayerA].Goal) - record.deltaGoalA)
	b[PlayerB].Goal = uint64(int64(b[PlayerB].Goal) - record.deltaGoalB)
	b[PlayerA].Buckets ^= record.deltaBucketsA
	b[PlayerB].Buckets ^= record.deltaBucketsB
}

// Add these functions for the AI
func (b Board) getValidMoves(player Player) []int {
	moves := []int{}
	for i := 0; i < 6; i++ {
		if b[player].getBucket(i) > 0 {
			moves = append(moves, i)
		}
	}
	return moves
}

func (b Board) evaluate(player Player) int {
	scoreA, scoreB := b.scores()
	if player == PlayerA {
		return int(scoreA) - int(scoreB)
	}
	return int(scoreB) - int(scoreA)
}

func minMax(board Board, depth int, maximizing bool, player Player, alpha, beta int) (int, int) {
	if depth == 0 || board.isEnd() {
		return board.evaluate(player), -1
	}

	validMoves := board.getValidMoves(player)
	if len(validMoves) == 0 {
		return board.evaluate(player), -1
	}

	opts := &MoveOptions{silent: true}
	var bestMove int
	if maximizing {
		maxEval := -1000000
		for _, move := range validMoves {
			record := board.recordMove(player)
			playAgain, _ := board.move(player, move, opts)
			var eval int
			if playAgain {
				eval, _ = minMax(board, depth-1, true, player, alpha, beta)
			} else {
				eval, _ = minMax(board, depth-1, false, !player, alpha, beta)
			}
			board.undoMove(record)

			if eval > maxEval {
				maxEval = eval
				bestMove = move
			}
			alpha = max(alpha, eval)
			if beta <= alpha {
				break
			}
		}
		return maxEval, bestMove
	} else {
		minEval := 1000000
		for _, move := range validMoves {
			record := board.recordMove(player)
			playAgain, _ := board.move(player, move, opts)
			var eval int
			if playAgain {
				eval, _ = minMax(board, depth-1, false, player, alpha, beta)
			} else {
				eval, _ = minMax(board, depth-1, true, !player, alpha, beta)
			}
			board.undoMove(record)

			if eval < minEval {
				minEval = eval
				bestMove = move
			}
			beta = min(beta, eval)
			if beta <= alpha {
				break
			}
		}
		return minEval, bestMove
	}
}

func getAIMove(board Board, player Player) int {
	_, move := minMax(board, MaxDepth, true, player, -1000000, 1000000)
	return move
}

// Helper functions for the AI
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Add MoveOptions type for silent moves during AI evaluation
type MoveOptions struct {
	silent bool
}
