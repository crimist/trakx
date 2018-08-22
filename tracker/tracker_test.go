package tracker_test

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestApp(t *testing.T) {
	fmt.Println("run the program")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
