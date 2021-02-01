package checks

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"testing"
	"time"
)

func isPrimeSqrt(value int) bool {
	for i := 2; i <= int(math.Floor(math.Sqrt(float64(value)))); i++ {
		if value%i == 0 {
			return false
		}
	}
	return value > 1
}

func burn(d time.Time) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		for i := 0; i < math.MaxInt64; i++ {
			if time.Now().After(d) {
				return
			}
			isPrimeSqrt(i)
		}
	}()

	return done
}

func TestChecksCheckLoad(t *testing.T) {

	check := &CheckLoad{}

	_, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	d := burn(time.Now().Add(time.Second))
	time.Sleep(time.Microsecond * 100)
	//Run it twice for windows
	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	<-d

	result, ok := cr.(*resultLoad)
	if !ok {
		t.Fatal("False type")
	}

	if runtime.GOOS != "windows" && result.Load1 <= 0 && result.Load5 <= 0 && result.Load15 <= 0 {
		//CPU load of 0 is impossible...
		t.Fatal("CPU load of 0 is impossible")
	}

	fmt.Println(result)

}
