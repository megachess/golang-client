package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Connection settings
	URL        = "wss://mega-chess.herokuapp.com/service"
	AUTH_TOKEN = "MEGACHESS_AUTH_TOKEN"

	// Actions
	CONNECT          = "connect"
	ASK_CHALLENGE    = "ask_challenge"
	ACCEPT_CHALLENGE = "accept_challenge"
	CHALLENGE        = "challenge"
	YOUR_TURN        = "your_turn"
	MOVE             = "move"

	// Board offsets
	MAX_OFFSET = 9
	MIN_OFFSET = 1
)

var conn *websocket.Conn

type MessageData struct {
	Auth_token string `json:"auth_token,omitempty"`
	Turn_token string `json:"turn_token,omitempty"`
	Message    string `json:"message,omitempty"`
	Username   string `json:"username,omitempty"`
	Board_id   string `json:"board_id,omitempty"`
}
type Message struct {
	Action      string      `json:"action,omitempty"`
	MessageData MessageData `json:"data,omitempty"`
}
type MoveData struct {
	Turn_token string `json:"turn_token"`
	Board_id   string `json:"board_id,omitempty"`
	From_col   string `json:"from_col,omitempty"`
	To_col     string `json:"to_col,omitempty"`
	From_row   string `json:"from_row,omitempty"`
	To_row     string `json:"to_row,omitempty"`
}
type Move struct {
	Action   string   `json:"action,omitempty"`
	MoveData MoveData `json:"data,omitempty"`
}

func main() {
	// Connect to server
	err := connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	// We start & keep reading from the websocket
	readForever()
}

func connect() error {
	var dialer *websocket.Dialer
	var err error
	conn, _, err = dialer.Dial(URL, nil)
	if err != nil {
		return err
	}

	auth_token := os.Getenv(AUTH_TOKEN)
	if len(auth_token) == 0 {
		return errors.New("Error: MEGACHESS_AUTH_TOKEN env variable not set")
	}

	connect_request := Message{Action: CONNECT, MessageData: MessageData{Auth_token: auth_token}}
	json_connect, _ := json.Marshal(connect_request)
	conn.WriteMessage(websocket.TextMessage, []byte(json_connect))
	log.Println("Connected to MegaChess!")

	return nil
}

func readForever() {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error while reading:", err)
			log.Println("Attempting to re-connect...")
			err = connect()
			if err != nil {
				log.Println("Failed to re-connect to websocket...")
				return
			}
		}

		var m Message
		err = json.Unmarshal(message, &m)

		log.Println("Received action: ", m.Action)

		if m.Action == ASK_CHALLENGE {
			acceptChallenge(m.MessageData.Board_id)
		}

		if m.Action == YOUR_TURN {
			moveFigure(m.MessageData.Board_id, m.MessageData.Turn_token)
		}
	}
}

func challenge() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Username to challenge: ")
	username, _ := reader.ReadString('\n')

	challenge_request := Message{Action: CHALLENGE, MessageData: MessageData{Username: username}}
	json_challenge, _ := json.Marshal(challenge_request)
	conn.WriteMessage(websocket.TextMessage, []byte(json_challenge))

}

func acceptChallenge(board_id string) {
	accept_request := Message{Action: ACCEPT_CHALLENGE, MessageData: MessageData{Board_id: board_id}}
	json_accept, _ := json.Marshal(accept_request)
	conn.WriteMessage(websocket.TextMessage, []byte(json_accept))
}

func moveFigure(board_id, turn_token string) {
	move_request := Move{Action: MOVE, MoveData: MoveData{
		Turn_token: turn_token,
		Board_id:   board_id,
		From_col:   randomPosition(MIN_OFFSET, MAX_OFFSET),
		To_col:     randomPosition(MIN_OFFSET, MAX_OFFSET),
		From_row:   randomPosition(MIN_OFFSET, MAX_OFFSET),
		To_row:     randomPosition(MIN_OFFSET, MAX_OFFSET),
	}}

	json_move, _ := json.Marshal(move_request)
	log.Println(string(json_move))

	conn.WriteMessage(websocket.TextMessage, []byte(json_move))
}

func randomPosition(min, max int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	pos := rand.Intn(max-min) + min

	return strconv.Itoa(pos)
}
