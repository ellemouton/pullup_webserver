package main

import (
	"log"
	"text/template"
)

var (
	successTmpl = parseTmpl("success", successTemplate)
)

func parseTmpl(name, tmpl string) *template.Template {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		log.Fatal(err)
	}
	t, err = t.Parse(baseTemplate)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

const baseTemplate = `{{define "base"}}
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Pull-Up Your Socks!</title>
<style>
	h1 {
		color: #08746f;
		font-size: 24px;
  	}
	.btn {
		margin: 5px;
		border: 1ps;
		padding: 10px;
		background-color: #08746f;
		color:white;
		overflow-wrap: break-word;
		font-size: 14px;
		font-family: "Lucida Console";
		border-radius: 3px;
		text-decoration: none;
	}
</style>
</head>
<body>
{{template "body" .}}
</body>
</html>
{{end}}
`
const successTemplate = `
{{template "base" .}}
{{define "body"}}
	<h1> Success! </h1>
	<a class="btn" href="/pullups">Log a Sesh!</a>
	<a class="btn" href="/view">View Stats</a>
{{end}}
`
