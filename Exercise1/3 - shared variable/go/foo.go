// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	//"time"
)

var i = 0

func incrementing(c chan<- int, done1 chan<- struct{}) {
	//TODO: increment i 1000000 times
	for n := 0; n < 10000; n++ {
		c <- 1
	}
    close(done1)
}

func decrementing(c chan<- int, done2 chan<- struct{}){
	//TODO: decrement i 1000000 times
	for n := 0; n < 100000; n++ {
		c <- -1
	}
    close(done2)
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(3)
    c := make(chan int)
    done1 := make(chan struct{})
    done2 := make(chan struct{})
	// TODO: Spawn both functions as goroutines
	go incrementing(c, done1)
	go decrementing(c, done2)
    go func(){
        <-done1
        <-done2
        close(c)
    }()

    for d:=range c{
        i += d
    }
	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.

    Println("The magic number is:", i)
}
