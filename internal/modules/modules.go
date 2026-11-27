package modules

import "fmt"

func ListModules() {
	fmt.Println("Available modules:")
	fmt.Println(" - scanner/portscan")
	fmt.Println(" - post/windows/gather/enum-shares")
	// add more modules as we implement it
}
