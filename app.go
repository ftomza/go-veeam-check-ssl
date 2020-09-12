package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type CheckResult struct {
	URL            string
	Hostname       string
	Port           string
	Result         string
	Desc           string
	ValidityExpire time.Time
}

const (
	ResultBAD = "BAD"
	ResultOK  = "OK"
)

func StartCheck(file *os.File, out *os.File) error {
	reader := bufio.NewReader(file)
	csvRes := csv.NewWriter(out)
	csvRes.Comma = ';'
	err := csvRes.Write([]string{"URL", "Hostname", "Port", "Result", "Desc", "ValidityExpire"})
	if err != nil {
		return err
	}
	csvRes.Flush()
	queue := make(chan *CheckResult, 4)
	wgQueue := sync.WaitGroup{}
	for {
		inLine, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		checkedUrl := strings.TrimSpace(inLine)
		wgQueue.Add(1)
		go func() {
			ctx, closer := context.WithTimeout(context.Background(), 5*time.Second)
			defer closer()
			log.Printf("Start check: %s\n", checkedUrl)
			queue <- CheckUrl(ctx, checkedUrl)
			log.Printf("Finish check: %s\n", checkedUrl)
			wgQueue.Done()
		}()
	}
	go func() {
		wgQueue.Wait()
		close(queue)
	}()
	for {
		res, ok := <-queue
		if !ok {
			return nil
		}
		err = csvRes.Write([]string{res.URL, res.Hostname, res.Port, res.Result, res.Desc, res.ValidityExpire.Format(time.RFC3339)})
		if err != nil {
			return err
		}
		csvRes.Flush()
	}
}

//CheckUrl check checkedUrl result into a CheckResult structure.
//
func CheckUrl(ctx context.Context, checkedUrl string) (res *CheckResult) {
	if checkedUrl == "" {
		return nil
	}
	res = &CheckResult{
		URL: checkedUrl,
	}
	if !strings.Contains(checkedUrl, "://") {
		checkedUrl = "https://" + checkedUrl
	}
	parseUrl, err := url.Parse(checkedUrl)
	if err != nil {
		res.Result = ResultBAD
		res.Desc = fmt.Sprintf("Parse URL: %s", err)
		return res
	}
	res.Hostname = parseUrl.Hostname()
	res.Port = parseUrl.Port()
	if res.Port == "" {
		res.Port = "443"
	}
	if res.Hostname == "" {
		res.Result = ResultBAD
		res.Desc = fmt.Sprintf("host not set")
		return res
	}
	dialer := new(net.Dialer)
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%s", res.Hostname, res.Port))
	if err != nil {
		res.Result = ResultBAD
		res.Desc = fmt.Sprintf("Dial error: %s", err)
		return res
	}
	c := tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
	err = c.Handshake()
	if err != nil {
		res.Result = ResultBAD
		res.Desc = fmt.Sprintf("Handshake error: %s", err)
		return res
	}
	certs := c.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		res.Result = ResultBAD
		res.Desc = fmt.Sprintf("Cert not found")
		return res
	}
	cert := certs[0]
	if cert.NotAfter.Before(time.Now()) {
		res.Result = ResultBAD
		res.Desc = fmt.Sprintf("validity expired")
		res.ValidityExpire = cert.NotAfter
		return res
	}
	res.Result = ResultOK
	res.ValidityExpire = cert.NotAfter
	return res
}

func main() {

	file, err := os.Open("input.txt")
	if err != nil {
		log.Fatalf("Cannot open input file: %s", err)
	}

	out, err := os.Create("output.csv")
	if err != nil {
		log.Fatalf("Cannot open output file: %s", err)
	}

	if err = StartCheck(file, out); err != nil {
		log.Fatalf("Check error: %s", err)
	}
}
