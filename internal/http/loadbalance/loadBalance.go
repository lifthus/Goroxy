package loadbalance

import (
	"fmt"
	"froxy/pkg/helper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func LoadBalanceRoundRobinHTTP(portNum string, listPath string) error {
	listStr, err := helper.OpenAndReadFile(listPath)
	if err != nil {
		return err
	}

	listStr = strings.Trim(listStr, "\n")
	targetList := strings.Split(listStr, "\n")
	urlList, err := helper.ParseStringsToHttpUrls(targetList)
	if err != nil {
		return err
	}

	host := helper.HttpLocalHostFromPort(portNum)
	proxy := httpRoundRobinloadBalancingReverseProxy(urlList)
	log.Printf("http round robin load balancer listening on: %s", portNum)
	logLoadBalanceTargets(urlList)
	if err := http.ListenAndServe(host, proxy); err != nil {
		return fmt.Errorf("http round robin load balancer ListenAndServe failed: %v", err)
	}
	return nil
}

func logLoadBalanceTargets(targets []*url.URL) {
	for i, target := range targets {
		fmt.Printf("T %d : %s\n", i+1, target)
	}
}

func httpRoundRobinloadBalancingReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	numTargets := len(targets)
	targetCnt := 0
	director := func(req *http.Request) {
		target := targets[targetCnt]
		targetCnt++
		targetCnt %= numTargets
		// targetCnt is captured but it won't be a complete round robin:
		// http server will spawn a new goroutine for each request,
		// so that the value of targetCnt may not always be added by 1 for each request.

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		// For simplicity, we don't handle RawQuery or the User-Agent header here:
		// see the full code of NewSingleHostReverseProxy for an example of doing
		// that.
	}
	return &httputil.ReverseProxy{Director: director}
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
