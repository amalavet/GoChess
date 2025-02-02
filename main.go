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
	PlayerA        Player = true
	PlayerB        Player = false
	MaxDepth              = 16 // How deep the AI looks ahead
	goalMultiplier        = 3

	// Bit masks for each 6-stone bucket position (000000 through 111111)
	BucketMask0 uint64 = 0b111111                               // First bucket:  000000-000101
	BucketMask1 uint64 = 0b111111000000                         // Second bucket: 000006-000011
	BucketMask2 uint64 = 0b111111000000000000                   // Third bucket:  000012-000017
	BucketMask3 uint64 = 0b111111000000000000000000             // Fourth bucket: 000018-000023
	BucketMask4 uint64 = 0b111111000000000000000000000000       // Fifth bucket:  000024-000029
	BucketMask5 uint64 = 0b111111000000000000000000000000000000 // Sixth bucket:  000030-000035
)

// Helper functions to get/set bucket values
func (s *Side) getBucket(i int) uint64 {
	shift := i * 6
	return (s.Buckets >> shift) & 0b111111 // Single 6-bit mask
}

func (b Board) display() {
	scoreA, scoreB := b.scores()
	fmt.Printf("\nScores - Player A: %d, Player B: %d\n\n", scoreA, scoreB)
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

	var playerIsA bool
	if playAgainstAI {
		fmt.Println("Do you want to go first? (y/n)")
		fmt.Scanln(&response)
		playerIsA = response == "y" || response == "Y"
	}

	board.display()
	for !board.isEnd() {
		var bucket int
		if playAgainstAI && ((playerIsA && currPlayer == PlayerB) || (!playerIsA && currPlayer == PlayerA)) {
			bucket = getAIMove(board, currPlayer)
			fmt.Printf("AI chooses bucket %d\n", bucket+1)
		} else {
			bucket = getBucketInput(currPlayer)
		}

		// Add validation checks
		if bucket > 5 {
			fmt.Println("Invalid move: bucket number too high")
			continue
		}
		stones := board[currPlayer].getBucket(bucket)
		if stones == 0 {
			fmt.Println("Invalid move: bucket is empty")
			continue
		}

		// Print move information
		playerName := "Player A"
		if currPlayer == PlayerB {
			playerName = "Player B"
		}
		fmt.Printf("\n%s picks up %d stones from bucket %d\n", playerName, stones, bucket+1)

		playAgain, _ := board.move(currPlayer, bucket)

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

func (b Board) move(player Player, bucket int) (bool, error) {
	// Get stones using bit shift and mask
	shift := bucket * 6
	stones := (b[player].Buckets >> shift) & 0b111111 // Single 6-bit mask

	// Clear the source bucket
	b[player].Buckets &= ^(uint64(0b111111) << shift)
	actualPlayer := player

	for stones > 0 {
		bucket++
		if bucket == 6 {
			if actualPlayer == player {
				b[player].Goal++
				stones--
			}
			player = !player
			bucket = -1
			continue
		}
		stones--

		// Increment bucket in one operation by adding 1 shifted to the bucket position
		b[player].Buckets += 1 << (bucket * 6)
	}

	return bucket == -1, nil
}

func (b Board) scores() (uint64, uint64) {
	return b[PlayerA].Goal*goalMultiplier +
			(b[PlayerA].Buckets & BucketMask0) +
			((b[PlayerA].Buckets & BucketMask1) >> 6) +
			((b[PlayerA].Buckets & BucketMask2) >> 12) +
			((b[PlayerA].Buckets & BucketMask3) >> 18) +
			((b[PlayerA].Buckets & BucketMask4) >> 24) +
			((b[PlayerA].Buckets & BucketMask5) >> 30),
		b[PlayerB].Goal*goalMultiplier +
			(b[PlayerB].Buckets & BucketMask0) +
			((b[PlayerB].Buckets & BucketMask1) >> 6) +
			((b[PlayerB].Buckets & BucketMask2) >> 12) +
			((b[PlayerB].Buckets & BucketMask3) >> 18) +
			((b[PlayerB].Buckets & BucketMask4) >> 24) +
			((b[PlayerB].Buckets & BucketMask5) >> 30)
}

func (b Board) isEnd() bool {
	return b[PlayerA].Buckets == 0 || b[PlayerB].Buckets == 0
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
	// Create a mask that has 1 in the lowest bit of each 6-bit group
	const validBitMask uint64 = 0b000001_000001_000001_000001_000001_000001

	// Get a value where each 1 bit represents a non-empty bucket
	nonEmptyBuckets := (b[player].Buckets & validBitMask) |
		((b[player].Buckets >> 1) & validBitMask) |
		((b[player].Buckets >> 2) & validBitMask) |
		((b[player].Buckets >> 3) & validBitMask) |
		((b[player].Buckets >> 4) & validBitMask) |
		((b[player].Buckets >> 5) & validBitMask)

	moves := make([]int, 0, 6)

	if nonEmptyBuckets&1 != 0 {
		moves = append(moves, 0)
	}
	if nonEmptyBuckets&(1<<6) != 0 {
		moves = append(moves, 1)
	}
	if nonEmptyBuckets&(1<<12) != 0 {
		moves = append(moves, 2)
	}
	if nonEmptyBuckets&(1<<18) != 0 {
		moves = append(moves, 3)
	}
	if nonEmptyBuckets&(1<<24) != 0 {
		moves = append(moves, 4)
	}
	if nonEmptyBuckets&(1<<30) != 0 {
		moves = append(moves, 5)
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

	var bestMove int
	if maximizing {
		maxEval := -1000000
		for _, move := range validMoves {
			record := board.recordMove(player)
			playAgain, _ := board.move(player, move)
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
			playAgain, _ := board.move(player, move)
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
