package main

type user struct {
	id       int
	username string
}

type pullup struct {
	id      int
	userID  int
	date    string
	pullups int
}

type Total struct {
	Username string
	Pullups  int
}

type ViewPage struct {
	Totals      []Total
	DailyTotals []Total
	WeeklyTotals []Total
	Day         string
	WinnerDay		string
	WinnerWeek		string
	WinnerAllTime	string

	Graph1Points []DayData
	Graph2Points []DayData
}

type DayData struct {
	Day        string
	UserPoints []int
}
