package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	socketio "github.com/googollee/go-socket.io"
)

var (
	currentGame  *GameState
	gameMutex    sync.RWMutex
	level        = 1
	socketServer *socketio.Server
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateRandNo(max int) int {
	return rand.Intn(max) + 1
}

type GameState struct {
	currentGuess int
	secretNumber int
	noOfTrials   int
	record       []int
}

func NewGameState(difficulty int) *GameState {
	rangeMax, attempts := getDifficultySettings(difficulty)
	return &GameState{
		secretNumber: generateRandNo(rangeMax),
		noOfTrials:   attempts,
		record:       make([]int, 0),
	}
}

func getDifficultySettings(difficulty int) (rangeMax, attempts int) {
	switch difficulty {
	case 1:
		fmt.Println("1. Easy (1-50, 10 attempts)")
		return 50, 10
	case 3:
		fmt.Println("3. Hard (1-200, 3 attempts)")
		return 200, 3
	default:
		fmt.Printf("%d. Medium (1-100, 5 attempts)\n", difficulty)
		return 100, 5
	}
}

func (g *GameState) MakeGuess(guess int) string {
	// Check for termination signal
	if guess == -1 {
		fmt.Printf("\n%s🎮 Game terminated by user%s\n", colorPurple, colorReset)
		return "Game terminated by user"
	}

	g.currentGuess = guess
	g.noOfTrials--
	g.record = append(g.record, guess)

	if currentGame.noOfTrials < 1 {
		currentGame.EndGame()
		currentGame = nil
		message := "You have used up the available no of trials.\n\033[32mYou can start the game to play again."
		message1 := "You have used up the available no of trials.\nYou can start the game to play again."
		fmt.Println(message)
		socketServer.BroadcastToNamespace("/", "board", message1)
		return message1
	}

	if guess < g.secretNumber {
		fmt.Println("Too low!")
		socketServer.BroadcastToNamespace("/", "board", "Too low!")
		return "Too low!"
	} else if guess > g.secretNumber {
		fmt.Println("Too high!")
		socketServer.BroadcastToNamespace("/", "board", "Too high!")
		return "Too high!"
	}
	var msg string
	if level >= 3 {
		msg = "Maximum level reached."
	} else {
		currentGame.EndGame(true)
		msg = "To move to the next level, Click Next"
	}
	fmt.Println("Correct! " + msg)
	return "Correct! " + msg
}

func (g *GameState) IsGameOver() bool {
	return g.noOfTrials <= 0 || g.currentGuess == g.secretNumber
}

func (g *GameState) EndGame(nextLevel ...bool) GameResponse {
	if g == nil {
		fmt.Printf("\n%s❌ No active game to end%s\n", colorRed, colorReset)
		socketServer.BroadcastToNamespace("/", "notice", "No active game to end")
		return GameResponse{
			Message:  "No active game to end",
			Success:  false,
			GameOver: true,
		}
	}

	// Set default value for nextLevel
	isNextLevel := false
	if len(nextLevel) > 0 {
		isNextLevel = nextLevel[0]
	}

	if isNextLevel {
		return GameResponse{
			Message:  fmt.Sprintln("Correct! To move to the next level, Click Next"),
			Success:  true,
			GameOver: true,
		}
	}

	fmt.Printf("\n%s🎮 Game ended. The number was: %d%s\n", colorPurple, g.secretNumber, colorReset)
	socketServer.BroadcastToNamespace("/", "notice", "🎮 Game ended. The number was: "+strconv.Itoa(g.secretNumber))
	return GameResponse{
		Message:  fmt.Sprintf("Game ended. The number was: %d", g.secretNumber),
		Success:  true,
		GameOver: true,
	}
}

type GameResponse struct {
	Message  string `json:"message"`
	Success  bool   `json:"success"`
	GameOver bool   `json:"gameOver"`
	Room     string `json:"room"`
}

func StartGameMessage(w http.ResponseWriter) {
	currentGame = NewGameState(level) // Default level is 1
	fmt.Printf(`
		%s╔═══════════════════════════════════════════════╗%s
		%s║             N U M B E R Q U E S T             ║%s
		%s║           The Ultimate Number Game            ║%s
		%s║───────────────────────────────────────────────║%s
		%s║  Progress through levels of increasing        ║%s
		%s║  difficulty to become the Number Master!      ║%s
		%s╚═══════════════════════════════════════════════╝%s

		`, colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset,
		colorCyan, colorReset)
	fmt.Println("\033[35m•Welcome to the NumberQuest!!")
	fmt.Println("•\033[32m•I chose a number between 1 and 100.")
	fmt.Println("•Can you guess what it is? (If you don't, enter -1 to quit)")
	fmt.Println("-----------------------------------------------------------")
	socketServer.BroadcastToNamespace("/", "notice", "NEW GAME STARTED! 🎮\n Welcome to the Number Guessing Game! \nI chose a number between 1 and 100. \n Can you guess what it is ?")
	json.NewEncoder(w).Encode(GameResponse{
		Message: "New game started",
		Success: true,
	})
}

func gameController(w http.ResponseWriter, r *http.Request) {
	gameMutex.Lock()
	defer gameMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	options := r.URL.Query()["option"]
	if len(options) == 0 {
		http.Error(w, "missing option parameter", http.StatusBadRequest)
		return
	}

	switch options[0] {
	case "start":
		StartGameMessage(w)
	case "next":
		if level <= 3 {
			level++
			StartGameMessage(w)
			return
		}
		if currentGame == nil {
			message := "No active game to proceed to next level"
			socketServer.BroadcastToNamespace("/", "notice", message)
			http.Error(w, message, http.StatusBadRequest)
			return
		}
		currentGame.EndGame()
		currentGame = nil
		message := "Maximum level reached. You can start the game to play again."
		fmt.Printf("\033[31m%s\n", message)
		socketServer.BroadcastToNamespace("/", "notice", message)
		http.Error(w, message, http.StatusBadRequest)
	case "end":
		if currentGame == nil {
			message := "No active game to end"
			socketServer.BroadcastToNamespace("/", "notice", message)
			json.NewEncoder(w).Encode(GameResponse{
				Message:  message,
				Success:  false,
				GameOver: true,
			})
			return
		}
		response := currentGame.EndGame()
		currentGame = nil
		json.NewEncoder(w).Encode(response)
	default:
		if currentGame == nil {
			socketServer.BroadcastToNamespace("/", "notice", "no active game")
			http.Error(w, "no active game", http.StatusBadRequest)
			return
		}

		guess, err := strconv.Atoi(options[0])
		if err != nil {
			socketServer.BroadcastToNamespace("/", "notice", "invalid number format")
			http.Error(w, "invalid number format", http.StatusBadRequest)
			return
		}

		message := currentGame.MakeGuess(guess)

		if currentGame.noOfTrials > 0 {
			fmt.Printf("\n%s📝 Guess: %d%s\n", colorYellow, guess, colorReset)
			fmt.Printf("%s %s (%d attempts left)%s\n",
				getMessageColor(message),
				message,
				currentGame.noOfTrials,
				colorReset,
			)
			fmt.Println("------------------------")

			json.NewEncoder(w).Encode(GameResponse{
				Message:  message,
				Success:  true,
				GameOver: currentGame.IsGameOver(),
			})
		}
	}
}

func getMessageColor(message string) string {
	switch {
	case strings.Contains(message, "low"):
		return colorBlue
	case strings.Contains(message, "high"):
		return colorRed
	case strings.Contains(message, "Correct"):
		return colorGreen
	case strings.Contains(message, "Congratulations"):
		return colorPurple
	default:
		return colorReset
	}
}

func main() {
	// Start a goroutine to watch for terminal input
	go func() {
		var input string
		for {
			fmt.Scan(&input)
			if input == "-1" {
				if currentGame != nil {
					fmt.Printf("\n%s🎮 Game terminated via terminal%s\n", colorPurple, colorReset)
					currentGame = nil
				}
			}
		}
	}()

	socketServer = socketio.NewServer(nil)

	socketServer.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("Connected:", s.ID())
		return nil
	})

	socketServer.OnEvent("/", "register", func(s socketio.Conn, id string) {
		fmt.Println("Registered:", id)
		s.Join(id)
		socketServer.BroadcastToRoom("/", id, "notice", "You successfully registered.")
	})

	socketServer.OnEvent("/", "message", func(s socketio.Conn, msg GameResponse) {
		fmt.Println("Received message:", msg)
		socketServer.BroadcastToRoom("/", msg.Room, "reply", msg)
	})

	socketServer.OnEvent("/", "play", func(s socketio.Conn, msg GameResponse) {
		fmt.Println("Received message:", msg)
		socketServer.BroadcastToRoom("/", msg.Room, "reply", msg)
	})

	socketServer.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("Disconnected:", s.ID(), "Reason:", reason)
		// s.Leave("room1")
	})

	socketServer.OnError("/", func(s socketio.Conn, e error) {
		log.Printf("Socket.IO error: %v\n", e)
	})

	go func() {
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("Socket.IO server error: %v\n", err)
		}
	}()
	defer socketServer.Close()

	// Your existing server code
	mux := http.NewServeMux()
	mux.Handle("/api", http.HandlerFunc(gameController))
	mux.Handle("/socket.io/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			return
		}

		socketServer.ServeHTTP(w, r)
	}))
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		// Create shutdown context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v\n", err)
		}

		if err := socketServer.Close(); err != nil {
			log.Printf("Socket server close error: %v\n", err)
		}
	}()

	log.Println("Serving at localhost:8080...")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
