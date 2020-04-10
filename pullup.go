package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/crypto/bcrypt"
)

var (
	db    *sql.DB
	GODIR string
)

func main() {
	database, err := sql.Open("mysql", "root:@(127.0.0.1:3306)/pullups")
	if err != nil {
		log.Fatal(err)
	}
	err = database.Ping()
	if err != nil {
		log.Fatal(err)
	}
	db = database
	GODIR = os.Getenv("GOPATH")

	// routes
	http.HandleFunc("/", viewStats)
	http.HandleFunc("/adduser", adduser)
	http.HandleFunc("/addpullup", addpullup)
	http.HandleFunc("/viewusers", viewusers)
	http.HandleFunc("/viewpullups", viewPullups)
	http.HandleFunc("/pullups", pullupHandler)
	http.HandleFunc("/view", viewStats)
	http.HandleFunc("/success", successHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome! Visit the '/pullups' endpoint to log your pull-ups.")
}

func viewusers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, username FROM users`)
	if err != nil {
		fmt.Fprintf(w, "Error querying from DB")
		return
	}
	defer rows.Close()

	var users []user
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.username)
		if err != nil {
			fmt.Fprintf(w, "Error scanning DB row")
			return
		}
		users = append(users, u)
	}

	fmt.Fprintln(w, users)
}

func viewPullups(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select * from pullups`)
	if err != nil {
		fmt.Fprintf(w, "Error querying from DB")
		return
	}
	defer rows.Close()

	var pullups []pullup
	for rows.Next() {
		var p pullup
		err := rows.Scan(&p.id, &p.userID, &p.date, &p.pullups)
		if err != nil {
			fmt.Fprintf(w, "Error scanning DB row")
			fmt.Fprintf(w, err.Error())
			return
		}
		pullups = append(pullups, p)
	}

	fmt.Fprintln(w, pullups)
}

func viewStats(w http.ResponseWriter, r *http.Request) {
	date, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	users, err := getAllUsers()
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	totals, err := getTotals(users)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	totalsPerDay, err := getTotalsPerDay(users, date)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	graphData, err := getDailyTotals(users, 7)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	totalsCumulative, err := getTotalsCumulative(users, 7)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	pageData := ViewPage{
		Totals:       totals,
		DailyTotals:  totalsPerDay,
		Day:          date.Format("2006-01-02"),
		Graph1Points: graphData,
		Graph2Points: totalsCumulative,
	}

	if r.Method == "POST" {

		r.ParseForm()

		date, err = time.Parse("2006-01-02", r.Form.Get("date"))
		if err != nil {
			fmt.Fprintf(w, "The date you entered is invalid. Please try again dodo brain.")
			fmt.Fprintf(w, err.Error())
			return
		}

		totalsPerDay, err = getTotalsPerDay(users, date)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		pageData = ViewPage{
			Totals:       totals,
			DailyTotals:  totalsPerDay,
			Day:          date.Format("2006-01-02"),
			Graph1Points: graphData,
			Graph2Points: totalsCumulative,
		}
	}

	tmpl, err := template.ParseFiles(GODIR + "/src/pullup/static/viewstats.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		fmt.Println(err)
	}
}

func getAllUsers() ([]user, error) {
	rows, err := db.Query(`select id, username from users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []user
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.username)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func getTotalsCumulative(urs []user, n int) ([]DayData, error) {
	var data []DayData

	for i := n; i >= 0; i-- {

		day := time.Now().AddDate(0, 0, -1*i).Format("2006-01-02")
		date, err := time.Parse("2006-01-02", day)
		if err != nil {
			return nil, err
		}

		totals, err := getTotalsCumulativePerDay(urs, date)
		if err != nil {
			return nil, err
		}

		var up []int
		for _, t := range totals {
			up = append(up, t.Pullups)
		}

		data = append(data, DayData{
			Day:        day,
			UserPoints: up,
		})
	}

	return data, nil
}

func getDailyTotals(urs []user, n int) ([]DayData, error) {
	var data []DayData

	for i := n; i >= 0; i-- {

		day := time.Now().AddDate(0, 0, -1*i).Format("2006-01-02")
		date, err := time.Parse("2006-01-02", day)
		if err != nil {
			return nil, err
		}

		totals, err := getTotalsPerDay(urs, date)
		if err != nil {
			return nil, err
		}

		var up []int
		for _, t := range totals {
			up = append(up, t.Pullups)
		}

		data = append(data, DayData{
			Day:        day,
			UserPoints: up,
		})
	}
	return data, nil
}

func getTotals(ur []user) ([]Total, error) {
	var totals []Total

	for _, u := range ur {
		row, err := db.Query(`select COALESCE(sum(pullups), 0) from pullups where user_id = ? `, u.id)
		if err != nil {
			return nil, err
		}
		defer row.Close()

		var totPullups int
		if row.Next() {
			err = row.Scan(&totPullups)
			if err != nil {
				return nil, err
			}
		}

		totals = append(totals, Total{Username: u.username, Pullups: totPullups})
	}
	return totals, nil
}

func getTotalsPerDay(ur []user, d time.Time) ([]Total, error) {
	var totals []Total

	for _, u := range ur {
		row, err := db.Query(`select COALESCE(sum(pullups), 0) from pullups where (user_id = ? && day = ?)`, u.id, d)
		if err != nil {
			return nil, err
		}
		defer row.Close()

		var totPullups int
		if row.Next() {
			err = row.Scan(&totPullups)
			if err != nil {
				return nil, err
			}
		}

		totals = append(totals, Total{Username: u.username, Pullups: totPullups})
	}
	return totals, nil
}

func getTotalsCumulativePerDay(ur []user, d time.Time) ([]Total, error) {
	var totals []Total

	for _, u := range ur {
		row, err := db.Query(`select COALESCE(sum(pullups), 0) from pullups where (user_id = ? && day <= ?)`, u.id, d)
		if err != nil {
			return nil, err
		}
		defer row.Close()

		var totPullups int
		if row.Next() {
			err = row.Scan(&totPullups)
			if err != nil {
				return nil, err
			}
		}

		totals = append(totals, Total{Username: u.username, Pullups: totPullups})
	}
	return totals, nil
}

func addpullup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles(GODIR + "/src/pullup/static/addpullup.html")
		if err != nil {
			log.Fatal(err)
		}
		tmpl.Execute(w, tmpl)
		return
	}

	r.ParseForm()

	userID := template.HTMLEscapeString(r.Form.Get("userID"))
	date, err := time.Parse("2006-01-02", r.Form.Get("date"))
	if err != nil {
		fmt.Fprintf(w, "The date you entered is invalid. Please try again dodo brain.")
		return
	}

	pullups, err := strconv.Atoi(r.Form.Get("number"))
	if err != nil || pullups < 0 {
		fmt.Fprintf(w, "Isn't entering a valid pullup number the point of this? Try again.")
		return
	}

	_, err = db.Exec(`insert into pullups (user_id, day, pullups) VALUES (?, ?, ?)`, userID, date, pullups)
	if err != nil {
		fmt.Fprintf(w, "error inserting new user into DB")
		fmt.Fprintf(w, err.Error())
		return
	}
}

func adduser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles(GODIR + "/src/pullup/static/adduser.html")
		if err != nil {
			log.Fatal(err)
		}
		tmpl.Execute(w, tmpl)
		return
	}

	r.ParseForm()

	username := template.HTMLEscapeString(r.Form.Get("username"))
	psswd := template.HTMLEscapeString(r.Form.Get("password"))

	if len(username) == 0 || len(psswd) == 0 {
		fmt.Fprintf(w, "Empty Username or Password? Try again.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(psswd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query(`select * from users where username = ?`, username)
	if err != nil {
		fmt.Fprintf(w, "Error querying from db.")
		fmt.Fprintf(w, err.Error())
		return
	}

	if rows.Next() {
		fmt.Fprintf(w, "Username taken.")
		return
	}

	_, err = db.Exec(`insert into users (username, password) VALUES (?, ?)`, username, hash)
	if err != nil {
		fmt.Fprintf(w, "error inserting new user into DB")
		return
	}
}

func pullupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles(GODIR + "/src/pullup/static/pullupform.html")
		if err != nil {
			log.Fatal(err)
		}
		tmpl.Execute(w, tmpl)
		return
	}

	r.ParseForm()

	username := template.HTMLEscapeString(r.Form.Get("username"))
	psswd := template.HTMLEscapeString(r.Form.Get("password"))

	if len(username) == 0 || len(psswd) == 0 {
		fmt.Fprintf(w, "Empty Username or Password? Try again.")
		return
	}

	date, err := time.Parse("2006-01-02", r.Form.Get("date"))
	if err != nil {
		fmt.Fprintf(w, "The date you entered is invalid. Please try again dodo brain.")
		return
	}

	pullups, err := strconv.Atoi(r.Form.Get("number"))
	if err != nil || pullups < 0 {
		fmt.Fprintf(w, "Isn't entering a valid pull-up number the point of this? Try again.")
		return
	}

	var (
		id       int
		passHash []byte
	)

	err = db.QueryRow(`select id, password from users where username = ?`, username).Scan(&id, &passHash)
	if err != nil {
		fmt.Fprintf(w, "User does not exist.")
		return
	}

	err = bcrypt.CompareHashAndPassword(passHash, []byte(psswd))
	if err != nil {
		fmt.Fprintln(w, "Incorrect Password.")
		return
	}

	_, err = db.Exec(`insert into pullups (user_id, day, pullups) VALUES (?, ?, ?)`, id, date, pullups)
	if err != nil {
		fmt.Fprintf(w, "error inserting new user into DB")
		fmt.Fprintf(w, err.Error())
		return
	}

	http.Redirect(w, r, "/success", http.StatusFound)
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	err := successTmpl.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}
