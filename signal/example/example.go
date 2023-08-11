package main

import (
	"fmt"
	"time"

	"github.com/snowzach/golib/signal"
)

func main() {

	// Setup the stop signal to handle interrupts.
	signal.Stop.OnSignal(signal.DefaultStopSignals...)

	fmt.Println("Press Ctrl-c to begin shutdown.")

	// Channel signal
	signal.Stop.Add(1)
	go func() {
		// Make everything wait until we exit
		defer signal.Stop.Done()

		// Loop and use the stop channel
		for {
			select {
			case <-signal.Stop.Chan():
				fmt.Println("Got stop channel signal. Cleaning up.")
				time.Sleep(2 * time.Second)
				fmt.Println("Stop channel exiting.")
				return
			case <-time.After(time.Second):
				fmt.Println("Stop routine spin.")
			}
		}
	}()

	// Bool signal
	signal.Stop.Add(1)
	go func() {
		// Make everything wait until we exit
		defer signal.Stop.Done()

		// Loop and use the stop boolean
		for !signal.Stop.Bool() {
			time.Sleep(time.Second)
			fmt.Println("Bool routine spin.")
		}
		fmt.Println("Got stop bool signal. Cleaning up.")
		time.Sleep(2 * time.Second)
		fmt.Println("Stop bool exiting.")
	}()

	// Context signal
	signal.Stop.Add(1)
	go func() {
		// Make everything wait until we exit
		defer signal.Stop.Done()

		ctx := signal.Stop.Context()

		// Loop and use the stop context (Err() == nil means it's not cancelled yet)
		for ctx.Err() == nil {
			time.Sleep(time.Second)
			fmt.Println("Context routine spin.")
		}
		fmt.Println("Got stop context signal. Cleaning up.")
		time.Sleep(2 * time.Second)
		fmt.Println("Stop context exiting.")
	}()

	// Wait here until we get the signal
	<-signal.Stop.Chan()
	fmt.Println("Main thread waiting for everyone to be ready to exit.")

	// Wait for everyone to cleanup
	signal.Stop.Wait()
	fmt.Println("Exiting.")

}
