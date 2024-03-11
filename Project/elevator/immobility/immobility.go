package immobility

func Immobility(obstCh chan bool, motorActivityCh chan bool, immob chan<- bool) {
	// TODO
	for {
		select {
		case obst := <-obstCh:
			if obst{ // && door.open()
				// TODO
				immob <- true
			} else {
				immob <- false
			}
		case motorActivity := <-motorActivityCh:
			if !motorActivity {
				// TODO
				immob <- true
			} else {
				immob <- false
			}
		}
	}
}
