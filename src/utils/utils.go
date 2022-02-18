package utils

import (
	"CIHunter/src/config"
	"fmt"
	"github.com/gocolly/colly"
	"math/rand"
	"os"
	"time"
)

func Init() {
	os.RemoveAll(config.REPOS_PATH)
}

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

// CommonCollector return a base collector
func CommonCollector() *colly.Collector {
	c := colly.NewCollector()

	// set random `User-Agent`
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", RandomString())
	})

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 404 {
			return
		}
		fmt.Println("[debug]", r.StatusCode, r.Request.URL)
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
		time.Sleep(time.Duration(50) * time.Millisecond)
		r.Retry()
	} else {
		fmt.Println("❗️ cannot fetch", r.URL)
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
