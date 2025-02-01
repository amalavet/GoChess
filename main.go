package main

import "fmt"

type (
	Player bool
	Board  map[Player]*Side
)

type Side struct {
	Goal    int
	Buckets [6]int
}

var StartBoard = Board{
	PlayerA: {
		Goal:    0,
		Buckets: [6]int{4, 4, 4, 4, 4, 4},
	},
	PlayerB: {
		Goal:    0,
		Buckets: [6]int{4, 4, 4, 4, 4, 4},
	},
}

const (
	PlayerA Player = true
	PlayerB Player = false
)

func (b Board) display() {
	// Clear screen (ANSI escape code)
	fmt.Print("\033[H\033[2J")

	scores := b.scores()
	fmt.Printf("\nScores - Player A: %d, Player B: %d\n\n", scores[PlayerA], scores[PlayerB])

	// Player B's buckets (reversed)
	fmt.Print("     ")
	for i := 5; i >= 0; i-- {
		fmt.Printf("[%d] ", b[PlayerB].Buckets[i])
	}
	fmt.Println()

	// Goals on both sides
	fmt.Printf("  %d                           %d\n", b[PlayerB].Goal, b[PlayerA].Goal)

	// Player A's buckets
	fmt.Print("     ")
	for i := 0; i < 6; i++ {
		fmt.Printf("[%d] ", b[PlayerA].Buckets[i])
	}
	fmt.Println("\n")
}

func main() {
	board := StartBoard
	currPlayer := PlayerA

	for !board.isEnd() {
		board.display()
		bucket := getBucketInput(currPlayer)
		playAgain, err := board.move(currPlayer, bucket)
		if err != nil {
			fmt.Printf("Invalid move: %v\n", err)
			continue
		}
		if !playAgain {
			currPlayer = !currPlayer
		}
	}

	// Show final state
	board.display()
	fmt.Println("Game Over!")
	fmt.Printf("Player A: %d, Player B: %d\n", board[PlayerA].Goal, board[PlayerB].Goal)
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
	if bucket > 5 {
		return false, fmt.Errorf("invalid bucket")
	}

	count := b[player].Buckets[bucket]
	if count == 0 {
		return false, fmt.Errorf("bucket is empty")
	}
	b[player].Buckets[bucket] = 0

	for count > 0 {
		count--
		bucket++
		if bucket == 6 {
			b[player].Goal++
			player = !player
			bucket = -1
			continue
		}
		b[player].Buckets[bucket]++
	}

	return bucket == -1, nil
}

func (b Board) scores() map[Player]int {
	scores := map[Player]int{
		PlayerA: b[PlayerA].Goal,
		PlayerB: b[PlayerB].Goal,
	}
	for _, bucket := range b[PlayerA].Buckets {
		scores[PlayerA] += bucket
	}
	for _, bucket := range b[PlayerB].Buckets {
		scores[PlayerB] += bucket
	}
	return scores
}

func (b Board) isEnd() bool {
	for _, bucket := range b[PlayerA].Buckets {
		if bucket > 0 {
			return false
		}
	}
	for _, bucket := range b[PlayerA].Buckets {
		if bucket > 0 {
			return false
		}
	}
	return true
}
