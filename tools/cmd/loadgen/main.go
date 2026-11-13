// Command loadgen replays customer uploads against the Invoicely API.
//
//	loadgen replay <customer>   upload one customer's file (desert-rose, oasis, al-noor, sahara, ghaf)
//	loadgen peak [-n 150]       morning-peak uploads from several customers
//	loadgen load                replay oasis, al-noor, sahara and ghaf, then the morning peak
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"invoicely/tools/internal/dataset"
	"invoicely/tools/internal/workbook"
)

var (
	apiURL = env("API_URL", "http://api:8080")
	client = &http.Client{Timeout: 10 * time.Minute}
	runTag = strings.ToUpper(strconv.FormatInt(time.Now().Unix()%1679616, 36))
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func logf(format string, args ...any) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

type importResp struct {
	ID           string         `json:"id"`
	Status       string         `json:"status"`
	Error        string         `json:"error"`
	InvoiceCount int            `json:"invoiceCount"`
	CreatedAt    time.Time      `json:"createdAt"`
	StatusCounts map[string]int `json:"statusCounts"`
}

func upload(sellerID int, filename string, data []byte) (*importResp, int, time.Duration, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("seller_id", strconv.Itoa(sellerID))
	fw, _ := w.CreateFormFile("file", filename)
	_, _ = fw.Write(data)
	_ = w.Close()

	start := time.Now()
	req, _ := http.NewRequest(http.MethodPost, apiURL+"/v1/imports", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return nil, 0, elapsed, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, resp.StatusCode, elapsed, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var imp importResp
	if err := json.Unmarshal(raw, &imp); err != nil {
		return nil, resp.StatusCode, elapsed, err
	}
	return &imp, resp.StatusCode, elapsed, nil
}

func getImport(id string) (*importResp, error) {
	resp, err := client.Get(apiURL + "/v1/imports/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var imp importResp
	return &imp, json.NewDecoder(resp.Body).Decode(&imp)
}

func formatCounts(c map[string]int) string {
	order := []string{"Parsed", "Validated", "Submitted", "Accepted", "Rejected", "Failed"}
	var parts []string
	for _, k := range order {
		if c[k] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", k, c[k]))
		}
	}
	if len(parts) == 0 {
		return "no invoices"
	}
	return strings.Join(parts, ", ")
}

// settle polls an import until its invoice statuses stop changing.
func settle(id string, maxWait time.Duration) *importResp {
	deadline := time.Now().Add(maxWait)
	var last string
	stable := 0
	var imp *importResp
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		cur, err := getImport(id)
		if err != nil {
			continue
		}
		imp = cur
		key := formatCounts(cur.StatusCounts)
		if key == last {
			stable++
			if stable >= 3 {
				return imp
			}
		} else {
			stable = 0
			last = key
		}
	}
	return imp
}

func findSample(key string) (dataset.Sample, bool) {
	for _, s := range dataset.Samples {
		if s.Key == key {
			return s, true
		}
	}
	return dataset.Sample{}, false
}

func replay(key string) error {
	s, ok := findSample(key)
	if !ok {
		var keys []string
		for _, s := range dataset.Samples {
			keys = append(keys, s.Key)
		}
		return fmt.Errorf("unknown customer %q (choose one of: %s)", key, strings.Join(keys, ", "))
	}
	data, err := workbook.Build(s.Build("-" + runTag))
	if err != nil {
		return err
	}
	logf("%-12s uploading %s (%d KB)", key, s.Filename, len(data)/1024)
	imp, code, elapsed, err := upload(s.SellerID, s.Filename, data)
	if err != nil {
		logf("%-12s upload failed after %.1fs: %v", key, elapsed.Seconds(), err)
		return nil
	}
	logf("%-12s HTTP %d in %.1fs  import=%s  status=%s  invoices=%d", key, code, elapsed.Seconds(), imp.ID, imp.Status, imp.InvoiceCount)
	if final := settle(imp.ID, 90*time.Second); final != nil {
		logf("%-12s invoices: %s", key, formatCounts(final.StatusCounts))
	}
	return nil
}

func peak(n, concurrency int) {
	logf("peak       %d uploads from %d customers, %d at a time", n, len(dataset.PeakSellers), concurrency)
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 7))
	type job struct {
		sellerID int
		name     string
		data     []byte
	}
	jobs := make(chan job)
	var ok, failed atomic.Int64
	var mu sync.Mutex
	failures := map[string]int{}
	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				_, _, _, err := upload(j.sellerID, j.name, j.data)
				if err != nil {
					failed.Add(1)
					msg := err.Error()
					if len(msg) > 80 {
						msg = msg[:80]
					}
					mu.Lock()
					failures[msg]++
					mu.Unlock()
				} else {
					ok.Add(1)
				}
				if done := ok.Load() + failed.Load(); done%25 == 0 {
					logf("peak       %d/%d uploads done (%d ok, %d failed)", done, n, ok.Load(), failed.Load())
				}
			}
		}()
	}
	for i := range n {
		seller := dataset.PeakSellers[i%len(dataset.PeakSellers)]
		prefix := fmt.Sprintf("%s-1116-%s-%03d", seller.Prefix, runTag, i+1)
		data, err := workbook.Build(dataset.PeakFile(seller.Party, prefix, r))
		if err != nil {
			logf("build failed: %v", err)
			continue
		}
		jobs <- job{seller.ID, strings.ToLower(seller.Prefix) + "-morning-" + strconv.Itoa(i+1) + ".xlsx", data}
	}
	close(jobs)
	wg.Wait()
	logf("peak       finished: %d ok, %d failed", ok.Load(), failed.Load())
	keys := make([]string, 0, len(failures))
	for k := range failures {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		logf("peak       %4d x %s", failures[k], k)
	}
}

func summary(since time.Time) {
	time.Sleep(15 * time.Second)
	var imports []importResp
	for range 30 {
		resp, err := client.Get(apiURL + "/v1/imports?limit=500")
		if err == nil {
			err = json.NewDecoder(resp.Body).Decode(&imports)
			resp.Body.Close()
			if err == nil {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	total := map[string]int{}
	byStatus := map[string]int{}
	for _, imp := range imports {
		if imp.CreatedAt.Before(since) {
			continue
		}
		byStatus[imp.Status]++
		for k, v := range imp.StatusCounts {
			total[k] += v
		}
	}
	logf("summary    imports since start: %v", byStatus)
	logf("summary    invoices: %s", formatCounts(total))
}

// waitForAPI waits until the API answers its health check.
func waitForAPI() {
	deadline := time.Now().Add(90 * time.Second)
	for {
		resp, err := client.Get(apiURL + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "the API at %s is not reachable. Is the stack running? Try `make up`.\n", apiURL)
			os.Exit(1)
		}
		time.Sleep(2 * time.Second)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: loadgen replay <customer> | peak [-n N] | load")
		os.Exit(2)
	}
	waitForAPI()
	start := time.Now()
	switch os.Args[1] {
	case "replay":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: loadgen replay <customer>")
			os.Exit(2)
		}
		if err := replay(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	case "peak":
		fs := flag.NewFlagSet("peak", flag.ExitOnError)
		n := fs.Int("n", 150, "number of uploads")
		c := fs.Int("c", 8, "concurrent uploads")
		_ = fs.Parse(os.Args[2:])
		peak(*n, *c)
		summary(start)
	case "load":
		logf("replaying customer traffic against %s (run %s)", apiURL, runTag)
		var wg sync.WaitGroup
		for _, key := range []string{"oasis", "al-noor", "sahara", "ghaf"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = replay(key)
			}()
		}
		wg.Wait()
		peak(240, 8)
		summary(start)
	default:
		fmt.Fprintln(os.Stderr, "unknown command", os.Args[1])
		os.Exit(2)
	}
}
