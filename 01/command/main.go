package main

import "fmt"

type Command interface {
	Execute()
}

type Player interface {
	Play()
	Pause()
}

type PlayCommand struct {
	Device Player
}

func (p *PlayCommand) Execute() {
	p.Device.Play()
}

type PauseCommand struct {
	Device Player
}

func (p *PauseCommand) Execute() {
	p.Device.Pause()
}

type VideoPlayer struct {
	playing bool
}

func (p *VideoPlayer) Play() {
	p.playing = true
}

func (p *VideoPlayer) Pause() {
	p.playing = false
}

type Button struct {
	Action Command
}

func (b *Button) Press() {
	b.Action.Execute()
}

func main() {
	player := &VideoPlayer{}

	play := &PlayCommand{Device: player}
	pause := &PauseCommand{Device: player}


    playButton := &Button{Action: play}
    pauseButton := &Button{Action: pause}

    playButton.Press()

    fmt.Printf("Playing: %v\n", player.playing)

    pauseButton.Press()
    fmt.Printf("Playing: %v\n", player.playing)

}
