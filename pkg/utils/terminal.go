package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
	"k8s.io/client-go/tools/remotecommand"
)

type termSizeQueue chan remotecommand.TerminalSize

func (this termSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-this
	if !ok {
		return nil
	}
	return &size
}

type Terminal struct {
	oldState  *term.State
	fd        int
	SizeQueue termSizeQueue
	cancel    context.CancelFunc
}

func NewTerminal() (*Terminal, error) {
	terminal := Terminal{
		fd:        int(os.Stdin.Fd()),
		SizeQueue: make(termSizeQueue, 1),
	}

	var err error

	terminal.oldState, err = term.MakeRaw(terminal.fd)
	if err != nil {
		return nil, err
	}

	return &terminal, nil
}

func (o *Terminal) MonitorSize() {
	ctx, cancel := context.WithCancel(context.Background())

	o.cancel = cancel

	go func() {
		for {
			termWidth, termHeight, err := term.GetSize(o.fd)
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
			}

			termSize := remotecommand.TerminalSize{Width: uint16(termWidth), Height: uint16(termHeight)}
			o.SizeQueue <- termSize

			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

func (o *Terminal) Close() error {
	if o.cancel != nil {
		o.cancel()
	}

	err := term.Restore(o.fd, o.oldState)
	if err != nil {
		return err
	}

	return nil
}
