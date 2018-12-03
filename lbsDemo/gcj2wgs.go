package main

import (
	"fmt"
	trans "github.com/eviltransform"
)

func main() {
	gcjLat := 22.488375
	gcjLng := 113.952356
	wgsLat, wgsLng := trans.GCJtoWGS(gcjLat, gcjLng)
	fmt.Println(wgsLat, wgsLng)
}
