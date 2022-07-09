package router

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var DIGITS = "0123456789"
var ALPHAS = "abcdefghijklmnopqrstuvwxyz"
var ALLCHARS = ALPHAS + DIGITS

func url_to_identifier(url string) string {
	url = strings.TrimPrefix(url, "https://github.com")
	url = strings.TrimPrefix(url, "http://github.com")
	url = strings.TrimPrefix(url, "/")
	first_slice_index := strings.Index(url, "/")
	second_slice_index := strings.Index(url[first_slice_index+1:], "/")
	if second_slice_index >= 0 {
		return url[:second_slice_index+first_slice_index+1]
	} else {
		return url
	}
}

func random_delay_ms(ms int) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(ms)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func is_direct_repo_link(candidate string) bool {
	if DEBUG {
		fmt.Println("determining candidate: " + candidate)
	}
	if strings.HasPrefix(candidate, "https://github.com/") && strings.Count(candidate, "/") == 4 {
		return true
	} else {
		return false
	}
}

func identifier_to_maintainer(identifier string) string {
	slash_pos := strings.Index(identifier, "/")
	return identifier[:slash_pos]
}
