package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Replaces a string template, where the variable is {{%var_name}}
func ParseTemplateString(inp string, vars map[string]string) string {
	items := []string{}
	for varName, val := range vars {
		items = append(items, "{{%"+varName+"}}")
		items = append(items, val)
	}
	r := strings.NewReplacer(items...)

	last := ""
	for last != inp {
		last = inp
		inp = r.Replace(inp)
	}

	return inp
}

// Creates a text progress bar. Goal is the target number, progress is the number out of goal.
//
// 70% progress bar of 200 characters is TextProgressBar(70, 100, "", "", 200), or TextProgressBar(0.7, 1, "", "", 200) or TextProgressBar(14, 20, "", "", 200)
func TextProgressBar(goal float64, progress float64, leftStr, rightStr string, size int) string {
	str := ""

	p := (progress / goal)
	if p > 1 {
		p = 1
	}
	if p < 0 {
		p = 0
	}

	epic := float64(size) * p

	for i := 0.0; i < epic; i++ {
		str += "█"
	}

	for utf8.RuneCountInString(str) != size {
		str += "─"
	}

	return fmt.Sprintf("`%v ├%v┤ %v`", leftStr, str, rightStr)
}
