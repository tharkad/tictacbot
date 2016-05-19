
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
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
		postMessage(ws, m)
	} else if parts[1] == "new" {
		if gamePtr.running {
			m.Text = "There is already a game in progress."
			postMessage(ws, m)
		} else {		
			m.Text = newGame(gamePtr)
			postMessage(ws, m)
		}
	} else if parts[1] == "new!" {
			m.Text = newGame(gamePtr)
			postMessage(ws, m)
	} else if parts[1] == "X" || parts[1] == "x" || parts[1] == "O" || parts[1] == "o" {
			m.Text = playerMove(gamePtr, parts)
			postMessage(ws, m)
	} else {
		m.Text = fmt.Sprintf("I don't understand. Try saying help to me.\n")
		postMessage(ws, m)
	}
}

func printBoard(gamePtr *tictactoe) string {
	runningStr := ""
	if gamePtr.running {
		runningStr = fmt.Sprintf("%s's Turn", gamePtr.turn)
	} else {
		runningStr = "The game is not in progress."
	}

	boardStr := fmt.Sprintf("%s | %s | %s\n%s | %s | %s\n%s | %s | %s\n%s", printPlayer(gamePtr.board[0]), printPlayer(gamePtr.board[1]), printPlayer(gamePtr.board[2]), printPlayer(gamePtr.board[3]), printPlayer(gamePtr.board[4]), printPlayer(gamePtr.board[5]), printPlayer(gamePtr.board[6]), printPlayer(gamePtr.board[7]), printPlayer(gamePtr.board[8]), runningStr)
	return boardStr
}

func printPlayer(player int) string {
	switch player {
		case 1: return "X"
		case 2: return "O"
		default: return ""
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
	if err != nil {
		return fmt.Sprintf("I didn't understand your move. You must say '%s #' where # is the postion you want to move on the board:\n1|2|3\n4|5|6\n7|8|9", gamePtr.turn)
	}

	fmt.Println(movePosition)
	return ""
}

// Get the quote via Yahoo. You should replace this method to something
// relevant to your team!
func getQuote(sym string) string {
	sym = strings.ToUpper(sym)
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1op&e=.csv", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if len(rows) >= 1 && len(rows[0]) == 5 {
		return fmt.Sprintf("%s (%s) is trading at $%s", rows[0][0], rows[0][1], rows[0][2])
	}
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}
