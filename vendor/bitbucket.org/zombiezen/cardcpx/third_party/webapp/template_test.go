package webapp

import (
	"fmt"
)

func ExampleCycle() {
	for i := 0; i < 3; i++ {
		fmt.Println(Cycle(i, "even", "odd"))
	}

	// Output:
	// even
	// odd
	// even
}
