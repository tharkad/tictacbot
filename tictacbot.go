
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"golang.org/x/net/websocket"
	"strconv"
)

type tictactoe struct {
	board [9]int
	turn string
	running bool
}


func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: tictacbot slack-bot-token\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("tictacbot ready, ^C exits")

	game := new(tictactoe)
	fmt.Println("New Game Ready")
	fmt.Printf("Board: %#v\n", game.board)

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			parseMessage(game, ws, m)
		}
	}
}

func parseMessage(gamePtr *tictactoe, ws *websocket.Conn, m Message) {
	// if so try to parse if
	fmt.Println("message")
	fmt.Printf("test %#v\n", m)
	parts := strings.Fields(m.Text)
	fmt.Printf("Fields: %#v\n", parts)
	if parts[1] == "board" {
		m.Text = printBoard(gamePtr)
	} else if parts[1] == "new" {
		if gamePtr.running {
			m.Text = "There is already a game in progress."
		} else {		
			m.Text = newGame(gamePtr)
		}
	} else if parts[1] == "new!" {
			m.Text = newGame(gamePtr)
	} else if parts[1] == "X" || parts[1] == "x" || parts[1] == "O" || parts[1] == "o" {
			m.Text = playerMove(gamePtr, parts)
	} else if parts[1] == "help" {
			m.Text = helpText()
	} else {
		m.Text = fmt.Sprintf("I don't understand. Try saying help to me.\n")
	}

	postMessage(ws, m)
}

func printBoard(gamePtr *tictactoe) string {
	runningStr := ""
	if gamePtr.running {
		runningStr = fmt.Sprintf("%s's Turn", gamePtr.turn)
	} else {
		runningStr = "The game over."
	}

	boardStr := fmt.Sprintf("```%s | %s | %s\n%s | %s | %s\n%s | %s | %s```\n%s", printPlayer(gamePtr.board[0]), printPlayer(gamePtr.board[1]), printPlayer(gamePtr.board[2]), printPlayer(gamePtr.board[3]), printPlayer(gamePtr.board[4]), printPlayer(gamePtr.board[5]), printPlayer(gamePtr.board[6]), printPlayer(gamePtr.board[7]), printPlayer(gamePtr.board[8]), runningStr)
	return boardStr
}

func printPlayer(player int) string {
	switch player {
		case 1: return "X"
		case 2: return "O"
		default: return " "
	}
}

func newGame(gamePtr *tictactoe) string {
	for i, _ := range gamePtr.board {
		gamePtr.board[i] = 0
	}
	gamePtr.running = true
	gamePtr.turn = "X"

	fmt.Println("New Game Ready")
	fmt.Printf("Board: %#v\n", gamePtr.board)

	return "New game ready. It is X's turn."
} 

func playerMove(gamePtr *tictactoe, parts []string) string {
	fmt.Println(parts)

	if !gamePtr.running {
		return ("The game is over. Say new to me to start a new game.")
	}

	listedPlayer := ""
	switch parts[1] {
		case "x": listedPlayer = "X"
		case "X": listedPlayer = "X"
		case "o": listedPlayer = "O"
		case "O": listedPlayer = "O"
		default: return fmt.Sprintf("It is player %s's Turn", gamePtr.turn)
	}

	if listedPlayer != gamePtr.turn {
		return fmt.Sprintf("It is player %s's Turn", gamePtr.turn)
	}

	movePosition, err := strconv.Atoi(parts[2])
	if err != nil || movePosition < 1 || movePosition > 9 {
		return fmt.Sprintf("I didn't understand your move. You must say '%s #' where # is the postion you want to move on the board:\n1|2|3\n4|5|6\n7|8|9", gamePtr.turn)
	}

	if gamePtr.board[movePosition - 1] != 0 {
		return fmt.Sprintf("You can't move there. Here is the board:\n%s", printBoard(gamePtr))
	} else {
		if gamePtr.turn == "X" {
			gamePtr.board[movePosition - 1] = 1
			gamePtr.turn = "O"
		} else {
			gamePtr.board[movePosition - 1] = 2
			gamePtr.turn = "X"
		}

		over, overText := isGameOver(gamePtr)
		if over {
			return overText
		} else {
			return fmt.Sprintf("%s", printBoard(gamePtr))
		}
	}

	return ""
}

func isGameOver(gamePtr *tictactoe) (bool, string) {
	winLines := [][]int {
		[]int{0,1,2},
		[]int{3,4,5},
		[]int{6,7,8},
		[]int{0,3,6},
		[]int{1,4,7},
		[]int{2,5,8},
		[]int{0,4,8},
		[]int{2,4,6},
	}

	for _, line := range winLines {
		if gamePtr.board[line[0]] != 0 && gamePtr.board[line[0]] == gamePtr.board[line[1]] && gamePtr.board[line[1]] == gamePtr.board[line[2]] {
			gamePtr.running = false
			return true, fmt.Sprintf("*%s Wins!*\n%s", printPlayer(gamePtr.board[line[0]]), printBoard(gamePtr))
		}
	}

	emptySpace := false
	for i := 0; i < 9; i++ {
		if gamePtr.board[i] == 0 {
			emptySpace = true
			break
		}
	}

	if !emptySpace {
			gamePtr.running = false
			return true, fmt.Sprintf("*It's a draw!*\n%s", printBoard(gamePtr))
	}

	return false, ""
}

func helpText() string {
	return ("I moderate a single Tic-Tac-Toe Game. Here's what you can say to me:\n*new* - Start a new game if there is not already one running.\n*new!* - Quit the current game a start a new one.\n*board* - Display the board and who's turn it is.\n*X #* - Place an X in board space #. The spaces are numbered across the board and down starting with 1 in the upper left corner.\n*O #* - Place an O in board space #.\n*help* - This help text.")
}
