package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const PaddleHeight = 4
const PaddleSymbol = 0x2588
const BallSymbol = 0x25CF
const InitialBallVelocityRow = 1
const InitialBallVelocityCol = 1

type GameObject struct {
	row, col, width, height int
	velRow, velColumn       int
	symbol                  rune
}

var isGamePaused bool = false
var player1Paddle *GameObject
var player2Paddle *GameObject
var ball *GameObject
var screen tcell.Screen

var gameObjects []*GameObject

func PrintStringCentered(row, col int, str string) {
	col = col - len(str)/2
	printString(row, col, str)
}

func printString(row, col int, str string) {
	for _, c := range str {
		screen.SetContent(col, row, c, nil, tcell.StyleDefault)
		col += 1
	}
}

func Print(row, col, width, height int, ch rune) {
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			screen.SetContent(col+c, row+r, ch, nil, tcell.StyleDefault)
		}
	}
}

func DrawState() {
	if isGamePaused {
		return
	}
	screen.Clear()
	for _, gameObject := range gameObjects {
		Print(gameObject.row, gameObject.col, gameObject.width, gameObject.height, gameObject.symbol)
	}
	screen.Show()
}

func CollidesWithWall(obj *GameObject) bool {
	_, screenHeight := screen.Size()
	return !(obj.row+obj.velRow >= 0 && obj.row+obj.velRow < screenHeight)
}

func CollidesWithPaddle(ball, paddle *GameObject) bool {
	var collidesCol bool
	if ball.col < paddle.col {
		collidesCol = ball.col+ball.velColumn >= paddle.col
	} else {
		collidesCol = ball.col+ball.velColumn <= paddle.col
	}
	return collidesCol &&
		ball.row >= paddle.row &&
		ball.row < paddle.row+paddle.height

}

func UpdateState() {
	if isGamePaused {
		return
	}
	for i := range gameObjects {
		gameObjects[i].row += gameObjects[i].velRow
		gameObjects[i].col += gameObjects[i].velColumn
	}
	if CollidesWithWall(ball) {
		ball.velRow = -ball.velRow
	}
	if CollidesWithPaddle(ball, player1Paddle) || CollidesWithPaddle(ball, player2Paddle) {
		ball.velColumn = -ball.velColumn
	}
}

func IsGameOver() bool {
	return getWinner() != ""
}

func InitScreen() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)

}
func InitUserInput() chan string {
	inputChan := make(chan string)
	go func() {
		for {
			switch ev := screen.PollEvent().(type) {
			case *tcell.EventKey:
				inputChan <- ev.Name()
			}
		}
	}()
	return inputChan
}

func HandleUserInput(key string) {
	_, screenHeight := screen.Size()
	if key == "Rune[q]" {
		screen.Fini()
		os.Exit(0)
	} else if key == "Rune[w]" && player1Paddle.row > 0 {
		player1Paddle.row--
	} else if key == "Rune[s]" && player1Paddle.row+player1Paddle.height < screenHeight {
		player1Paddle.row++
	} else if key == "Up" && player2Paddle.row > 0 {
		player2Paddle.row--
	} else if key == "Down" && player2Paddle.row+player2Paddle.height < screenHeight {
		player2Paddle.row++
	} else if key == "Rune[p]" {
		isGamePaused = !isGamePaused
	}
}

func InitGameState() {
	width, height := screen.Size()
	paddleStart := height/2 - PaddleHeight/2
	player1Paddle = &GameObject{
		row:       paddleStart,
		col:       0,
		width:     1,
		velRow:    0,
		velColumn: 0,
		height:    PaddleHeight,
		symbol:    PaddleSymbol,
	}
	player2Paddle = &GameObject{
		row:       paddleStart,
		col:       width - 1,
		velRow:    0,
		velColumn: 0,
		width:     1,
		height:    PaddleHeight,
		symbol:    PaddleSymbol,
	}
	ball = &GameObject{
		row:       height / 2,
		col:       width / 2,
		width:     1,
		height:    1,
		symbol:    BallSymbol,
		velRow:    InitialBallVelocityRow,
		velColumn: InitialBallVelocityCol,
	}
	gameObjects = []*GameObject{
		player1Paddle, player2Paddle, ball,
	}
}

func ReadInput(inputChan chan string) string {
	var key string
	select {
	case key = <-inputChan:
	default:
		key = ""
	}
	return key
}

func main() {
	InitScreen()
	InitGameState()
	inputChan := InitUserInput()

	for !IsGameOver() {
		HandleUserInput(ReadInput(inputChan))
		UpdateState()
		DrawState()
		time.Sleep(75 * time.Millisecond)

	}

	winner := getWinner()
	screenWidth, screenHeight := screen.Size()
	PrintStringCentered(screenHeight/2-1, screenWidth/2, "Game Over!")
	PrintStringCentered(screenHeight/2, screenWidth/2, fmt.Sprintf("%s win......", winner))
	screen.Show()

	time.Sleep(3 * time.Second)
	screen.Fini()
}

func getWinner() string {
	screenWidth, _ := screen.Size()
	if ball.col < 0 {
		return "Player 1"
	} else if ball.col >= screenWidth {
		return "Player 2"
	} else {
		return ""
	}
}
