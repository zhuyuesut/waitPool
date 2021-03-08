package waitPool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFunc(t *testing.T) {
	pools := New(2, 10)
	t1 := time.Now()
	for i := 0; i < 10; i++ {
		pools.Add(context.Background(), task)
	}
	pools.Close()
	fmt.Printf("%vms\n", time.Since(t1).Milliseconds())
}

func task() {
	time.Sleep(time.Second)
}
