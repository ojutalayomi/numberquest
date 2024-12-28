# NumberQuest - The Ultimate Number Game

A multiplayer number guessing game built with Go and Socket.IO where players progress through increasingly difficult levels.

## Features

- Multiple difficulty levels with increasing complexity
- Real-time multiplayer gameplay using Socket.IO
- Interactive numpad interface
- Graceful server shutdown
- Cross-platform compatibility
- Mobile-friendly responsive design

## Getting Started

1. Clone the repository
2. Navigate to the server-game directory
3. Run `go run main.go`
4. Open `http://localhost:8080/numberquest` in your browser

## How to Play

### Desktop
1. Click "New Game" to start
2. Guess the secret number between 1-100
3. Use the numpad or keyboard to enter your guess
4. Get feedback if your guess is too high or too low
5. Progress through levels by correctly guessing numbers
6. Try to become the Number Master!

### Mobile Devices
1. Open the game URL on your mobile browser
2. Tap "New Game" to begin
3. Use the on-screen numpad to enter numbers
4. Tap "Send" to submit your guess
5. The interface automatically adjusts to your screen size
6. Play in portrait or landscape mode
7. Enjoy smooth touch interactions

## Game Controls

- **New Game** - Start a new game session
- **End Game** - End the current game
- **Next Level** - Progress to next level after winning
- **Numpad** - Enter guesses using on-screen buttons
- **Send** - Submit your guess
- **Clear (⌫)** - Delete entered numbers

## Mobile-Specific Features

- Touch-optimized buttons
- Responsive layout that adapts to screen size
- No keyboard required - full on-screen controls
- Portrait and landscape support
- Smooth animations and transitions
- Clear visual feedback on touch

## Technical Details

- Backend: Go with Gorilla WebSocket
- Frontend: HTML, CSS, JavaScript
- Real-time Communication: Socket.IO
- State Management: Server-side game state
- Concurrency: Mutex-protected shared resources
- Mobile Support: Responsive design, touch events

## License

This project is licensed under the MIT License - see the LICENSE file for details.

