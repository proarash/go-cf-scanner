package report

import (
	"bufio"
	"log"
	"os"
)

func SaveHits(done chan<- struct{}, hits <-chan string) {
	file, err := os.OpenFile("hits.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		close(done)
		return
	}

	writer := bufio.NewWriter(file)
	defer close(done)
	defer file.Close()
	defer writer.Flush()

	for ip := range hits {
		if _, err := writer.WriteString(ip + "\n"); err != nil {
			log.Println(err)
			return
		}
	}
}
