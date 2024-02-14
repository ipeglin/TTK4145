package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	count := 0
	filename := "dataStorage.txt"
	sleep_backup := 100
	sleep_primary := 151

	for {
		data, err := os.ReadFile(filename)

		if err != nil {
			fmt.Println(err)
		} else {

			dataStr := string(data)
			if dataStr == "" {
				break
			}
			split_data := strings.Split(dataStr, " ")
			write_time_str := split_data[0]
			count, _ = strconv.Atoi(split_data[1])
			write_time, _ := time.Parse(time.RFC3339Nano, write_time_str)
			delta_time := time.Since(write_time) - time.Duration(sleep_primary)*time.Millisecond
			limit := 100 * time.Millisecond
			if delta_time > limit {
				break
			}
			fmt.Println(count)
		}
		time.Sleep(time.Duration(sleep_backup) * time.Millisecond)
	}

	cmd := exec.Command("gnome-terminal", "--", "bash", "-c", "./main; exec bash")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting command: %s\n", err)
	}

	for {
		//Telle
		count++
		//data_str := time.Now().String() + " " + strconv.Itoa(count)
		data_str := fmt.Sprintf("%s %d", time.Now().Format(time.RFC3339Nano), count)

		//Skrive til fil
		data := []byte(data_str)
		os.WriteFile(filename, data, 0644)
		time.Sleep(time.Duration(sleep_primary) * time.Millisecond)

	}
}
