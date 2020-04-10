# Pullup Your Socks! :muscle: :socks: 
Basic golang pullup tracking webserver

This is a basic golang webserver that I built in order to get the hang of go webservers and basic mysql usage.
It also allows my climbing friends and I to compete with one another during the COVID19 lockdown :muscle:

The server is up and running and can be found at: http://52.16.76.131/pullups

## Step 1

The db directory contains the sql commands that need to be sourced before the server can be run. And your mysql username and 
password along with address and port need to be set in the 'pullup.go' file in the main function.

## Step 2

Compile the binary: <pre><code>$ go install pullup</code></pre>

### Step 3

Run the binary: <pre><code>$ ./bin/pullup &</code></pre>

### Step 4

Take note of the process number and disown it:

<pre><code>$ disown PID </code></pre>
