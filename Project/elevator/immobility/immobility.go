package immobility

func Immobility(obstCh chan bool, motorActivityCh chan bool, immob chan<- bool) {
	for {
		select {
		case obst := <-obstCh:
			if obst {
				immob <- true
			} else {
				immob <- false
			}
		case motorActivity := <-motorActivityCh:
			if !motorActivity {
				immob <- true
			} else {
				immob <- false
			}
		}
	}
}
