package utils

import (
	"CIAnalyser/config"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

var DIGITS = "0123456789"
var ALPHAS = "abcdefghijklmnopqrstuvwxyz"
var ALLCHARS = ALPHAS + DIGITS

func RandomString() string {
	const bytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, rand.Intn(10)+10)
	for i := range b {
		b[i] = bytes[rand.Intn(len(bytes))]
	}
	return string(b)
}

// RandomIntArray return shuffled int array
func RandomIntArray(min, max int) []int {
	intArray := make([]int, max-min+1)
	for i := range intArray {
		intArray[i] = min + i
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(intArray), func(i, j int) {
		intArray[i], intArray[j] = intArray[j], intArray[i]
	})

	return intArray
}

func ShuffleStringArray(src []string) []string {
	final := make([]string, len(src))
	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(src))

	for i, v := range perm {
		final[v] = src[i]
	}
	return final
}

// CommonCollector return a base collector
func CommonCollector() *colly.Collector {
	c := colly.NewCollector()

	// set random `User-Agent`
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", RandomString())
	})

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 404 || r.StatusCode == 500 {
			return
		}
		if config.DEBUG {
			fmt.Println("[debug]", r.StatusCode, r.Request.URL)
		}
		retryRequest(r.Request, config.TRYOUT)
	})

	return c
}

func retryRequest(r *colly.Request, maxRetries int) int {
	retriesLeft := maxRetries
	if x, ok := r.Ctx.GetAny("retriesLeft").(int); ok {
		retriesLeft = x
	}
	if retriesLeft > 0 {
		r.Ctx.Put("retriesLeft", retriesLeft-1)
		time.Sleep(time.Duration(config.TRYOUT-retriesLeft) * time.Second)
		r.Retry()
	} else {
		fmt.Println("! cannot fetch", r.URL)
	}
	return retriesLeft
}

func DirExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		return false
	}
}

var count = 0

// RequestGitHubToken return one token from the list and loop the pointer
func RequestGitHubToken() string {
	l := len(config.GITHUB_TOKEN)
	var mutex sync.Mutex
	mutex.Lock()

	token := config.GITHUB_TOKEN[count%l] // select one token
	count++                               // auto increment

	if count > l { // prevent overflow
		count -= l
	}

	mutex.Unlock()
	return token
}

func RandomDelay(ms int) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(ms)
	time.Sleep(time.Duration(n) * time.Millisecond)
}
