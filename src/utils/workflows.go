package utils

import "strings"

// CartesianProduct takes map of lists and returns list of unique tuples
func CartesianProduct(mapOfLists map[string][]interface{}) []map[string]interface{} {
	listNames := make([]string, 0)
	lists := make([][]interface{}, 0)
	for k, v := range mapOfLists {
		listNames = append(listNames, k)
		lists = append(lists, v)
	}

	listCart := cartN(lists...)

	rtn := make([]map[string]interface{}, 0)
	for _, list := range listCart {
		vMap := make(map[string]interface{})
		for i, v := range list {
			vMap[listNames[i]] = v
		}
		rtn = append(rtn, vMap)
	}
	return rtn
}

func cartN(a ...[]interface{}) [][]interface{} {
	c := 1
	for _, a := range a {
		c *= len(a)
	}
	if c == 0 || len(a) == 0 {
		return nil
	}
	p := make([][]interface{}, c)
	b := make([]interface{}, c*len(a))
	n := make([]int, len(a))
	s := 0
	for i := range p {
		e := s + len(a)
		pi := b[s:e]
		p[i] = pi
		s = e
		for j, n := range n {
			pi[j] = a[j][n]
		}
		for j := len(n) - 1; j >= 0; j-- {
			n[j]++
			if n[j] < len(a[j]) {
				break
			}
			n[j] = 0
		}
	}
	return p
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TrimRunner(runners []string) []string {
	var finalRunners []string

	latestMapping := map[string]string{
		"ubuntu-latest":  "ubuntu-20.04",
		"macos-latest":   "macos-11",
		"windows-latest": "windows-2019",
	}

	for _, runner := range runners {
		// lowercase the macOS runner
		if runner == "macOS-11" || runner == "macOS-10.15" || runner == "macOS-latest" {
			runner = strings.ToLower(runner)
		}

		if val, ok := latestMapping[runner]; ok {
			runner = val
		}

		if !contains(finalRunners, runner) {
			finalRunners = append(finalRunners, runner)
		}
	}

	return finalRunners
}

func TrimEscape(v string) string {
	trimPrefix := strings.TrimPrefix(v, "${{")
	trimSuffix := strings.TrimSuffix(trimPrefix, "}}")
	return strings.TrimSpace(trimSuffix)
}
