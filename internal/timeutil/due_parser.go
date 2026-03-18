package timeutil

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reAbsoluteDate = regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})`)
	reChineseDate  = regexp.MustCompile(`(\d{1,2})月(\d{1,2})[日号]?`)
	reDaysLater    = regexp.MustCompile(`(\d+)\s*天后`)
	reWeekday      = regexp.MustCompile(`(本|下)周([一二三四五六日天])`)
)

func NormalizeDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func ExtractDueDate(text string, baseDate time.Time) (*time.Time, bool) {
	t := strings.TrimSpace(text)
	if t == "" {
		return nil, false
	}

	base := NormalizeDate(baseDate)

	if m := reAbsoluteDate.FindStringSubmatch(t); len(m) == 4 {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, base.Location())
		return &date, true
	}

	if m := reChineseDate.FindStringSubmatch(t); len(m) == 3 {
		month, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		date := time.Date(base.Year(), time.Month(month), day, 0, 0, 0, 0, base.Location())
		return &date, true
	}

	if strings.Contains(t, "今天") {
		date := base
		return &date, true
	}

	if strings.Contains(t, "明天") {
		date := base.AddDate(0, 0, 1)
		return &date, true
	}

	if strings.Contains(t, "后天") {
		date := base.AddDate(0, 0, 2)
		return &date, true
	}

	if m := reDaysLater.FindStringSubmatch(t); len(m) == 2 {
		days, _ := strconv.Atoi(m[1])
		date := base.AddDate(0, 0, days)
		return &date, true
	}

	if m := reWeekday.FindStringSubmatch(t); len(m) == 3 {
		weekFlag := m[1]
		targetWeekday := chineseWeekdayToInt(m[2])
		currentWeekday := currentWeekdayIndex(base.Weekday())
		startOfWeek := base.AddDate(0, 0, -(currentWeekday - 1))
		candidate := startOfWeek.AddDate(0, 0, targetWeekday-1)
		if weekFlag == "本" && candidate.Before(base) {
			candidate = candidate.AddDate(0, 0, 7)
		}
		if weekFlag == "下" {
			candidate = startOfWeek.AddDate(0, 0, 7+(targetWeekday-1))
		}
		date := NormalizeDate(candidate)
		return &date, true
	}

	if strings.Contains(t, "下月底") {
		firstOfNextMonth := time.Date(base.Year(), base.Month()+1, 1, 0, 0, 0, 0, base.Location())
		lastOfNextMonth := firstOfNextMonth.AddDate(0, 1, -1)
		date := NormalizeDate(lastOfNextMonth)
		return &date, true
	}

	if strings.Contains(t, "本月底") || strings.Contains(t, "月底") {
		firstOfThisMonth := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, base.Location())
		lastOfThisMonth := firstOfThisMonth.AddDate(0, 1, -1)
		date := NormalizeDate(lastOfThisMonth)
		return &date, true
	}

	return nil, false
}

func chineseWeekdayToInt(s string) int {
	switch s {
	case "一":
		return 1
	case "二":
		return 2
	case "三":
		return 3
	case "四":
		return 4
	case "五":
		return 5
	case "六":
		return 6
	default:
		return 7
	}
}

func currentWeekdayIndex(w time.Weekday) int {
	switch w {
	case time.Sunday:
		return 7
	default:
		return int(w)
	}
}
