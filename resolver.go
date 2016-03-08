package main
import (
	"fmt"
	"net"
	"flag"
	"os"
	"bufio"
	"strings"
	"sync"
)

var tlds = []string{"com", "jp"}

func read_tld_list(filename string) (tlds []string, err error) {
	var fp *os.File
	tlds = make([]string, 0, 100)

	fp, err = os.Open(filename)
	if err != nil {
		return
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.Index(line, "#") == 0 {
			continue
		}
		tld := strings.ToLower(line)
		tlds = append(tlds, tld)
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}

func check_tld(tld string) {
	var domain = "example." + tld
	lookup_map, err := net.LookupHost(domain)
	if err != nil {
		fmt.Printf("fail to resolve %s: %s\n", domain, err)
		return
	}
	fmt.Println("resolve " + domain + " to " + lookup_map[0])
}

func test_tld_worker(tasksCh <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		tld, ok := <-tasksCh
		if !ok {
			return
		}
		check_tld(tld)
	}
}

func test_tlds_pool(wg *sync.WaitGroup, workers int, tlds []string) {
	tasksCh := make(chan string)

	for i := 0; i < workers; i++ {
		go test_tld_worker(tasksCh, wg)
	}

	for _, tld := range tlds {
		tasksCh <- tld
	}
	close(tasksCh)
}

func main() {
	var (
		tld_file = flag.String("tld-file", "", "tld list")
	)
	flag.Parse()

	if *tld_file != "" {
		var err error
		tlds, err = read_tld_list(*tld_file)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
	}

	var wg sync.WaitGroup
	workers := 5
	wg.Add(workers)
	go test_tlds_pool(&wg, workers, tlds)
	wg.Wait()
}
