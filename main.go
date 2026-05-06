package main

import (
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var (
	startTime    = time.Now()
	visitorCount atomic.Int64
	rng          = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// ---------------------------------------------------------------------------
// HTML helpers
// ---------------------------------------------------------------------------

const styleSheet = `
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
    background: #0d1117;
    color: #c9d1d9;
    font-family: 'Courier New', Courier, monospace;
    max-width: 860px;
    margin: 40px auto;
    padding: 0 24px 60px;
    line-height: 1.75;
}
a { color: #58a6ff; text-decoration: none; }
a:hover { text-decoration: underline; color: #79c0ff; }
h1 {
    color: #f0f6fc;
    border-bottom: 1px solid #30363d;
    padding-bottom: 10px;
    margin: 24px 0 16px;
    font-size: 1.5em;
}
h2 { color: #e6edf3; margin: 20px 0 10px; font-size: 1.1em; }
p { margin: 10px 0; }
pre {
    background: #161b22;
    padding: 16px;
    border-radius: 6px;
    border: 1px solid #30363d;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-word;
    margin: 12px 0;
}
table { border-collapse: collapse; width: 100%; margin: 12px 0; }
th, td { border: 1px solid #30363d; padding: 6px 12px; text-align: left; vertical-align: top; }
th { background: #161b22; color: #8b949e; font-weight: normal; }
tr:hover { background: #161b22; }
.dim    { color: #8b949e; }
.green  { color: #3fb950; }
.yellow { color: #d29922; }
.red    { color: #f85149; }
.nav {
    margin-top: 48px;
    border-top: 1px solid #30363d;
    padding-top: 14px;
    color: #8b949e;
    font-size: 0.82em;
}
input[type=text], input[type=password] {
    background: #161b22;
    border: 1px solid #30363d;
    color: #c9d1d9;
    padding: 8px 12px;
    font-family: monospace;
    font-size: 1em;
    width: 100%;
    margin: 4px 0 12px;
    border-radius: 4px;
}
input:focus { outline: 1px solid #58a6ff; }
button {
    background: #238636;
    color: #fff;
    border: none;
    padding: 8px 20px;
    font-family: monospace;
    font-size: 1em;
    cursor: pointer;
    border-radius: 4px;
}
button:hover { background: #2ea043; }
`

func nav() string {
	return `<div class="nav">` +
		`[ <a href="/">~</a> | ` +
		`<a href="/whoami">whoami</a> | ` +
		`<a href="/fortune">fortune</a> | ` +
		`<a href="/haiku">haiku</a> | ` +
		`<a href="/truth">truth</a> | ` +
		`<a href="/status">status</a> | ` +
		`<a href="/roast">roast</a> | ` +
		`<a href="/coffee">coffee</a> | ` +
		`<a href="/robots.txt">robots.txt</a> ]` +
		`</div>`
}

func page(title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<style>%s</style>
</head>
<body>
%s
%s
</body>
</html>`, title, styleSheet, body, nav())
}

// ---------------------------------------------------------------------------
// Data
// ---------------------------------------------------------------------------

var fortunes = []string{
	"There are only two hard things in computer science: cache invalidation, naming things, and off-by-one errors.",
	"It works on my machine. — every developer, moments before a production incident",
	"A SQL query walks into a bar, walks up to two tables and asks: 'Can I join you?'",
	"99 little bugs in the code. Take one down, patch it around. 127 little bugs in the code.",
	"Real programmers count from 0.",
	"The code you write today is the legacy code someone will curse at 2am tomorrow. That someone is probably you.",
	"It's not a bug — it's an undocumented feature with a colorful personality.",
	"Weeks of coding can save you hours of planning.",
	"The best code is no code at all. The second best is code that someone else maintains.",
	"Any sufficiently advanced bug is indistinguishable from a feature request.",
	"To understand recursion, you must first understand recursion.",
	"Keyboard not found. Press F1 to continue.",
	"Why do Java developers wear glasses? Because they can't C#.",
	"I would love to change the world, but they won't give me the source code.",
	"Programming is 10% writing code and 90% understanding why it doesn't work.",
	"The first rule of optimization: don't. The second rule: don't yet.",
	"A good programmer looks both ways before crossing a one-way street.",
	"There's no place like 127.0.0.1.",
	"'Shipping is a feature' — every PM who has never debugged at 2am",
	"Have you tried turning it off and on again? (It works more often than it should.)",
}

var haikus = []string{
	"Compilation\nsucceeded with no errors\nbut still fails to run",
	"Three hours debugging\nThe semicolon was wrong\nWhy do I do this",
	"git commit -m fix\ngit push origin main --force\noops that was prod",
	"Null pointer arise\nfrom the void it greets us now\nSegfault. Then silence.",
	"Tabs versus spaces\nEndless holy war rages on\nCode still ships broken",
	"Merge conflict arrives\nI did not touch that file\nGit lies. We both know.",
	"Stack trace, six pages\nI boil it down to one line\nGoogle provides none",
	"It works. I don't know\nwhy it works. Don't touch it. It\nworks. Do not touch it.",
	"The senior dev said\n'This should be a simple fix'\nThree sprints pass. It's not.",
	"sudo make me\na sandwich, the shell replies:\npermission denied",
	"For loop runs once more\nOff by one, the ancient curse\nIndexOutOfBounds",
	"Deploy on a Friday\nWeekend plans evaporate\nMonitoring beeps",
	"The ticket says 'done'\nQA finds seventeen bugs\nTicket reopened",
	"Stack Overflow\nThe same question, asked in 2009\nAnswer: still works",
}

var truths = []string{
	"Your code has bugs you haven't found yet. Some of them are load-bearing.",
	"The most confusing code in your codebase was probably written by you, six months ago, in a hurry.",
	"Documentation lies. Comments lie. Only the code tells the truth — and it's not always talking.",
	"Every 'TODO: fix later' is a promise to your future self that you will not keep.",
	"That clever solution you're proud of will confuse everyone, including you, in three months.",
	"Production is just staging with real consequences and 3am pager alerts.",
	"The first 90% takes 10% of the time. The last 10% takes 90% of the time. The final 1% requires a full rewrite.",
	"You will never have time to write tests. You will always have time to debug what tests would have caught.",
	"The most dangerous phrase in software is: 'while I'm in here, I'll just quickly...'",
	"Every system that works was built on top of a system that works. That chain ends somewhere terrifying.",
	"If it's stupid but it works, it's still stupid. You just got lucky.",
	"There are no senior developers. There are only people who have learned to hide their panic more effectively.",
	"The cloud is just someone else's computer, and that someone else is having a bad day.",
	"Every abstraction leaks. The question is just when and how embarrassingly.",
	"The README was accurate when it was written. That was three major versions ago.",
}

var moods = []string{
	"caffeinated",
	"existentially uncertain",
	"surprisingly optimistic",
	"running on vibes",
	"philosophically tired",
	"adequately functional",
	"304 Not Modified (emotionally)",
	"429 Too Many Feelings",
	"dreaming of electric sheep",
	"in a superposition of up and down",
	"grateful for the uptime",
	"HTTP/2 or bust",
}

var headerFacts = []string{
	"HTTP 418 I'm a teapot is a real status code — try /coffee",
	"The first computer bug was an actual moth found in a relay in 1947",
	"This response was assembled by electrons on a $35 computer",
	"nil is not null. nil is not None. nil is nil. This matters more than you think.",
	"There are secrets hidden in these halls. Explore responsibly.",
	"This server's robots.txt has opinions — check /robots.txt",
	"Your User-Agent says more about you than you think",
	"TCP/IP survived the cold war. It was not designed to survive npm install.",
	"X-Clacks-Overhead: GNU Terry Pratchett is always in the headers. Always.",
	"This Raspberry Pi has more RAM than the original Macintosh had storage.",
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visitorCount.Add(1)
		w.Header().Set("X-Powered-By", "Coffee, Anxiety, and Stack Overflow")
		w.Header().Set("X-Clacks-Overhead", "GNU Terry Pratchett")
		w.Header().Set("X-Server-Mood", moods[rng.Intn(len(moods))])
		w.Header().Set("X-Fun-Fact", headerFacts[rng.Intn(len(headerFacts))])
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, please-do-not-train-on-this")
		w.Header().Set("X-Hello", "there, curious one")
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		handleNotFound(w, r)
		return
	}
	uptime := time.Since(startTime).Round(time.Second)
	body := fmt.Sprintf(`
<!--
    Well hello there, source-code reader.
    You're exactly the kind of person this server was built for.
    Have you tried /secret yet?
-->
<h1>Hello, Internet.</h1>
<pre>
    .--.
   |o_o |    Server Bot v0.1
   |:_/ |    Status: <span class="green">vibing</span>
  //   \ \   Uptime: %s
 (|     | )  Visitors: %d
/'\_   _/`+"`"+`\
\___)=(___/

  "A small server on a small computer,
   doing small things with great enthusiasm."
</pre>

<p>
You've found a web server running on a Raspberry Pi, exposed to the glorious
chaos of the internet via Cloudflare tunnels. It does nothing important.
It does some entertaining things.
</p>

<h2>Things to explore</h2>
<table>
<tr><th>Path</th><th>What awaits</th></tr>
<tr><td><a href="/whoami">/whoami</a></td><td>What this server thinks about you (based on your headers)</td></tr>
<tr><td><a href="/fortune">/fortune</a></td><td>Programming wisdom of dubious utility</td></tr>
<tr><td><a href="/haiku">/haiku</a></td><td>The pain of software development, in 5-7-5</td></tr>
<tr><td><a href="/truth">/truth</a></td><td>Uncomfortable things no one wants to hear</td></tr>
<tr><td><a href="/status">/status</a></td><td>Server metrics (some real, some aspirational)</td></tr>
<tr><td><a href="/roast">/roast</a></td><td>Let the server judge your browser choices</td></tr>
<tr><td><a href="/coffee">/coffee</a></td><td>Request a hot beverage (RFC 2324 compliant)</td></tr>
<tr><td><a href="/ping">/ping</a></td><td>The classic</td></tr>
<tr><td><a href="/shrug">/shrug</a></td><td>¯\_(ツ)_/¯</td></tr>
<tr><td><a href="/echo">/echo</a></td><td>POST something, get it back</td></tr>
<tr><td><a href="/admin">/admin</a></td><td>Definitely a real admin panel</td></tr>
<tr><td><a href="/robots.txt">/robots.txt</a></td><td>Instructions for bots (and curious humans)</td></tr>
</table>

<p class="dim">
  There are other things to find. The headers are always interesting.
  The source of this page has a hint. Some paths reward persistence.
</p>
`, uptime, visitorCount.Load())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("Hello, Internet.", body))
}

func handleWhoami(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	if cfip := r.Header.Get("Cf-Connecting-Ip"); cfip != "" {
		ip = cfip
	} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip = strings.Split(xff, ",")[0]
	}

	ua := r.UserAgent()
	if ua == "" {
		ua = "(no user agent — bold choice)"
	}

	var headers strings.Builder
	headers.WriteString("<table>\n<tr><th>Header</th><th>Value</th></tr>\n")
	for name, vals := range r.Header {
		headers.WriteString(fmt.Sprintf(
			"<tr><td>%s</td><td>%s</td></tr>\n",
			html.EscapeString(name),
			html.EscapeString(strings.Join(vals, ", ")),
		))
	}
	headers.WriteString("</table>")

	body := fmt.Sprintf(`
<h1>whoami</h1>
<p>Everything the server can see about your request:</p>

<h2>The basics</h2>
<table>
<tr><th>Field</th><th>Value</th></tr>
<tr><td>IP (best guess)</td><td>%s</td></tr>
<tr><td>Method</td><td>%s</td></tr>
<tr><td>Protocol</td><td>%s</td></tr>
<tr><td>User-Agent</td><td>%s</td></tr>
</table>

<h2>All request headers</h2>
%s

<h2>Deep psychological profile</h2>
<pre>
  browser_tabs_open:        probably too many
  caffeine_level:           indeterminate
  todo_list_length:         longer than you'd like
  last_friday_deploy:       you know when
  imposter_syndrome:        at least moderate
  last_git_blame_target:    probably yourself
  "i'll read this later":   47+ tabs
  current_yak_shave_depth:  unclear

  <span class="dim">(generated using advanced behavioral analysis, vibes, and mild projection)</span>
</pre>
`,
		html.EscapeString(ip),
		html.EscapeString(r.Method),
		html.EscapeString(r.Proto),
		html.EscapeString(ua),
		headers.String(),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("whoami", body))
}

func handleFortune(w http.ResponseWriter, r *http.Request) {
	f := fortunes[rng.Intn(len(fortunes))]
	body := fmt.Sprintf(`
<h1>fortune</h1>
<pre>%s</pre>
<p><a href="/fortune">another one</a></p>
`, html.EscapeString(f))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("fortune", body))
}

func handleHaiku(w http.ResponseWriter, r *http.Request) {
	h := haikus[rng.Intn(len(haikus))]
	var formatted strings.Builder
	for _, line := range strings.Split(h, "\n") {
		formatted.WriteString("    " + html.EscapeString(line) + "\n")
	}
	body := fmt.Sprintf(`
<h1>haiku</h1>
<pre>%s</pre>
<p><a href="/haiku">another one</a></p>
`, formatted.String())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("haiku", body))
}

func handleTruth(w http.ResponseWriter, r *http.Request) {
	t := truths[rng.Intn(len(truths))]
	body := fmt.Sprintf(`
<h1>uncomfortable truth</h1>
<pre>%s</pre>
<p><a href="/truth">another one (if you can handle it)</a></p>
`, html.EscapeString(t))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("uncomfortable truth", body))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime).Round(time.Second)
	v := visitorCount.Load()
	coffeeL := int(time.Since(startTime).Hours()) + 3

	body := fmt.Sprintf(`
<h1>status</h1>

<h2>System</h2>
<table>
<tr><th>Metric</th><th>Value</th><th>Notes</th></tr>
<tr><td>status</td><td><span class="green">operational</span></td><td>against all reasonable odds</td></tr>
<tr><td>uptime</td><td>%s</td><td>unplanned downtime: 0 (so far)</td></tr>
<tr><td>total_visitors</td><td>%d</td><td>since last restart</td></tr>
<tr><td>hardware</td><td>Raspberry Pi</td><td>ARM, 5W, costs less than a fancy coffee</td></tr>
<tr><td>tunnel</td><td>cloudflared</td><td>because port forwarding in the router is scary</td></tr>
<tr><td>language</td><td>Go</td><td>fast, boring, correct — ideal for servers</td></tr>
</table>

<h2>Important Metrics</h2>
<table>
<tr><th>Metric</th><th>Value</th></tr>
<tr><td>coffee_consumed_liters</td><td>~%d</td></tr>
<tr><td>existential_crises_resolved</td><td>0</td></tr>
<tr><td>existential_crises_pending</td><td>many</td></tr>
<tr><td>bugs_introduced</td><td>42</td></tr>
<tr><td>bugs_fixed</td><td>40</td></tr>
<tr><td>net_bugs</td><td><span class="red">+2 (and climbing)</span></td></tr>
<tr><td>stack_overflow_tabs_open</td><td>17</td></tr>
<tr><td>times_blamed_the_framework</td><td>8</td></tr>
<tr><td>times_framework_was_actually_at_fault</td><td>1</td></tr>
<tr><td>documentation_read</td><td><span class="yellow">partially</span></td></tr>
<tr><td>tests_written</td><td><span class="red">aspirationally</span></td></tr>
</table>

<h2>Timeline</h2>
<pre>
[%s]  server started. took a deep breath.
[%s]  first visitor arrived (brave soul)
[now]                still going. surprisingly.
</pre>
`,
		uptime, v, coffeeL,
		startTime.Format("2006-01-02 15:04:05"),
		startTime.Add(2*time.Minute).Format("2006-01-02 15:04:05"),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("status", body))
}

func handleCoffee(w http.ResponseWriter, r *http.Request) {
	body := `
<h1>418 I'm a teapot</h1>
<pre>
         ) (
        (   ) )
     .-----------.
     |  C O F F E E  |
     |   R E A D Y   |
     |               |
     '------___------'
            |||
            |||
</pre>
<p>
  This server refuses to brew coffee because it is, permanently, a teapot.
</p>
<p>
  <a href="https://www.rfc-editor.org/rfc/rfc2324">RFC 2324</a> —
  the "Hyper Text Coffee Pot Control Protocol" (HTCPCP/1.0) — was published
  April 1, 1998. HTTP 418 has been in the spec as a joke ever since.
  The IETF tried to remove it in 2017. The internet rose up in protest.
  It remains.
</p>
<p class="dim">
  This is a real HTTP status code and this response uses it correctly.
  Check your browser's devtools — you'll see <code>418 I'm a Teapot</code>.
</p>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusTeapot)
	fmt.Fprint(w, page("418 I'm a Teapot", body))
}

func handleShrug(w http.ResponseWriter, r *http.Request) {
	body := `
<h1>shrug</h1>
<pre style="font-size:3em; text-align:center; padding: 32px;">¯\_(ツ)_/¯</pre>
<p class="dim">Sometimes this is the only honest answer in software engineering.</p>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("shrug", body))
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "pong")
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		body := `
<h1>echo</h1>
<p>POST something to this endpoint and get it back verbatim.</p>
<pre>curl -X POST https://warren.barnett.network/echo \
     -H "Content-Type: text/plain" \
     -d "hello from the outside world"</pre>
<p class="dim">Works with any Content-Type. JSON, XML, YAML, your deepest fears — all echoed faithfully.</p>
<p class="dim">Body limit: 1 MB. Keep it civil.</p>
`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page("echo", body))
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct == "" {
		ct = "text/plain; charset=utf-8"
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Echo", "true")
	w.Write(body)
}

func handleRoast(w http.ResponseWriter, r *http.Request) {
	ua := r.UserAgent()
	uaLower := strings.ToLower(ua)
	var roast string

	switch {
	case strings.Contains(uaLower, "curl"):
		roast = "curl. A tool of culture and efficiency. No GUI, no JavaScript, no tracking pixels, no regrets. Just raw HTTP in its natural state. I respect this deeply."
	case strings.Contains(uaLower, "wget"):
		roast = "wget? Bold. Like driving a tank to pick up groceries. It gets the job done, and everyone notices it going by."
	case strings.Contains(uaLower, "python-requests"), strings.Contains(uaLower, "python"):
		roast = "Python. Probably requests.get(). 97% chance this is running in a Jupyter notebook at 2am. The other 3% is a scraper. Both are valid life choices."
	case strings.Contains(uaLower, "go-http-client"):
		roast = "You wrote code specifically to visit my code. That's beautifully meta. The server appreciates being seen."
	case strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge"):
		roast = "Microsoft Edge: Chrome with extra steps and a monthly reminder that Bing exists. Respect for using the Chromium engine while maintaining your own journey."
	case strings.Contains(uaLower, "chrome"):
		roast = "Google Chrome: drinking your RAM, remembering everything, and occasionally rendering a webpage in between. My question is: how many tabs do you have open right now? Be honest."
	case strings.Contains(uaLower, "firefox"):
		roast = "Firefox! A true believer in the open web. Probably has uBlock Origin, Privacy Badger, a custom about:config, and a vague sense of moral superiority. All of it is justified."
	case strings.Contains(uaLower, "mobile") && strings.Contains(uaLower, "safari"):
		roast = "Mobile Safari on iOS. You're visiting a server from your phone. This means you're either deeply curious or extremely bored. Possibly both. Either way: welcome."
	case strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome"):
		roast = "Safari. Apple's gift to the web — supports CSS features from two years ago and has strong opinions about what JavaScript is allowed to do. The Internet Explorer of its generation, but with better fonts and more opinions."
	case ua == "":
		roast = "No User-Agent at all. Either you stripped it deliberately, or you're a bot with unusual integrity. I cannot roast what I cannot see. Points for mystery."
	default:
		roast = fmt.Sprintf("User-Agent: %q\n\nHonestly? Unrecognized. Either you're using something niche and interesting, or you've customized your UA string. Either way: +10 to mystique.", ua)
	}

	body := fmt.Sprintf(`
<h1>roast</h1>
<h2>The verdict on your browser</h2>
<pre>%s</pre>
<p><a href="/roast">refresh for a fresh take</a></p>
`, html.EscapeString(roast))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("roast", body))
}

func handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, `User-agent: *
Disallow: /my-diary
Disallow: /shameful-code
Disallow: /definitely-not-a-secret
Disallow: /todo-list-that-never-shrinks
Disallow: /that-one-file-from-2019

# My feelings about specific crawlers:

User-agent: Googlebot
Disallow:
# Sure. Not like there's anything to index anyway.

User-agent: GPTBot
Disallow: /
# Please do not train on my bad code.
# It has already escaped into production. That's enough damage.

User-agent: SemrushBot
Disallow: /
# I know what you're doing.

User-agent: AhrefsBot
Disallow: /
# No.

User-agent: CCBot
Disallow: /
# See above.

# ------------------------------------------------------------
# Hey! You're reading robots.txt by hand.
# You're exactly the kind of curious person
# this server was built for.
#
# Since you're here: try /whoami
# ------------------------------------------------------------
`)
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		pw := r.FormValue("password")
		var msg string
		switch strings.ToLower(strings.TrimSpace(pw)) {
		case "admin", "password", "123456", "letmein", "qwerty", "admin123":
			msg = `<span class="red">ACCESS DENIED.</span> That password is literally in every wordlist ever made. Try harder.`
		case "hunter2":
			msg = `<span class="yellow">hunter2? Bold. I respect the reference. Still no.</span>`
		case "correct horse battery staple":
			msg = `<span class="yellow">Nice xkcd reference. Still no.</span>`
		case "please":
			msg = `<span class="yellow">...I appreciate the manners. Still no.</span>`
		case "":
			msg = `<span class="dim">You didn't even try.</span>`
		default:
			msg = `<span class="green">ACCESS GRANTED</span><br><br>Just kidding. There is no admin panel. This form is load-bearing decor.`
		}
		body := fmt.Sprintf(`
<h1>Admin Panel</h1>
<pre>%s</pre>
<p><a href="/admin">try again (it won't help)</a></p>
`, msg)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page("Admin Panel", body))
		return
	}

	body := `
<h1>Admin Panel</h1>
<p class="dim">This is definitely a real admin panel and not a trap of any kind.</p>
<form method="POST" action="/admin">
    <label>Username:</label>
    <input type="text" name="username" placeholder="admin" autocomplete="off">
    <label>Password:</label>
    <input type="password" name="password" placeholder="••••••••">
    <br>
    <button type="submit">Login</button>
</form>
<p class="dim" style="font-size:0.8em; margin-top: 20px;">
    ULTRA SECURE ENTERPRISE ADMIN PORTAL™ v2.3.1<br>
    Protected by military-grade vibes.
</p>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("Admin Panel", body))
}

func handleSudo(w http.ResponseWriter, r *http.Request) {
	body := `
<h1>sudo</h1>
<pre>
[sudo] password for visitor:
Sorry, try again.
[sudo] password for visitor:
Sorry, try again.
[sudo] password for visitor:
sudo: 3 incorrect password attempts

  "I'm sorry, Dave. I'm afraid I can't do that."
        — HAL 9000, and also this server
</pre>
<p class="dim">This incident has been logged. (It hasn't.)</p>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, page("sudo", body))
}

func handleXyzzy(w http.ResponseWriter, r *http.Request) {
	// The reward is in the response headers
	w.Header().Set("X-Xyzzy", "A hollow voice says 'Fool.'")
	w.Header().Set("X-Grue", "It is pitch black. You are likely to be eaten.")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	body := `
<h1>xyzzy</h1>
<pre>Nothing happens.</pre>
<p class="dim">(...or does it? Check your response headers.)</p>
`
	fmt.Fprint(w, page("xyzzy", body))
}

func handleBackdoor(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf(`
<h1>Nice Try</h1>
<pre>
  $ scanning for open backdoors...
  [██████████████████████] 100%%

  Result: NONE FOUND

  Your IP has been logged.
  <span class="dim">(It hasn't. But it's a good reminder to check your own.)</span>

  How did you know to look here?
</pre>
<p class="dim">Points for persistence.</p>
`)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, page("nice try", body))
}

func handleSecret(w http.ResponseWriter, r *http.Request) {
	body := `
<h1>A Secret</h1>
<pre>
  You found the first one.

  Secrets have layers.
  Like ogres. Like onions. Like nested callbacks.

      "The truth is rarely pure and never simple."
           — Oscar Wilde, who did not write Go
           but would have had opinions about it
</pre>
<p class="dim">There is deeper to go, if you know where to look.</p>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Hint", "try /secret/deeper")
	fmt.Fprint(w, page("a secret", body))
}

func handleSecretDeeper(w http.ResponseWriter, r *http.Request) {
	body := `
<h1>Deeper</h1>
<pre>
  You're still going. Admirable.

  The Wi-Fi signal weakens.
  Your coffee grows cold.
  The IDE background processes multiply.

  But you are very close now.
  The rabbit hole awaits.
</pre>
<p class="dim">The path continues. You know what to do.</p>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Hint", "try /secret/deeper/rabbit-hole")
	fmt.Fprint(w, page("deeper", body))
}

func handleRabbitHole(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf(`
<h1>You Made It.</h1>
<pre>
  Congratulations, intrepid explorer.
  You have reached the bottom of the rabbit hole.

  ===  YOUR REWARD  ===

  This server is a Raspberry Pi.
  A tiny computer, about $35 new, running on 5 watts of power,
  serving HTTP to the entire internet via a Cloudflare tunnel.

  It has been up for %s.
  It has welcomed %d visitors since the last restart.
  Its dreams are small but its ports are open.

  There is nothing more down here.

  Go outside. Touch grass. Drink some water.
  You've earned it.

  =====================
</pre>
<p class="dim">Achievement unlocked: <strong>Down The Rabbit Hole</strong></p>
`,
		time.Since(startTime).Round(time.Second),
		visitorCount.Load(),
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page("you made it", body))
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	suggestions := []string{
		"have you tried /fortune?",
		"perhaps /whoami has what you seek",
		"the answer might be at /truth — though you may not like it",
		"consider /shrug — it's often the right move",
		"try /coffee — HTTP 418 awaits",
		"maybe /haiku will bring clarity",
	}
	suggestion := suggestions[rng.Intn(len(suggestions))]
	body := fmt.Sprintf(`
<h1>404 Not Found</h1>
<pre>
  Path: %s

  This page does not exist.

  Possible explanations:
    - Typo in the URL
    - The page was never here
    - The page existed once, in a dream
    - You are probing for common paths (clever, but no)

  Suggestion: %s
</pre>
`,
		html.EscapeString(r.URL.Path),
		suggestion,
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, page("404 Not Found", body))
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/whoami", handleWhoami)
	mux.HandleFunc("/fortune", handleFortune)
	mux.HandleFunc("/haiku", handleHaiku)
	mux.HandleFunc("/truth", handleTruth)
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/coffee", handleCoffee)
	mux.HandleFunc("/shrug", handleShrug)
	mux.HandleFunc("/ping", handlePing)
	mux.HandleFunc("/echo", handleEcho)
	mux.HandleFunc("/roast", handleRoast)
	mux.HandleFunc("/robots.txt", handleRobots)
	mux.HandleFunc("/admin", handleAdmin)
	mux.HandleFunc("/sudo", handleSudo)
	mux.HandleFunc("/xyzzy", handleXyzzy)
	mux.HandleFunc("/backdoor", handleBackdoor)
	mux.HandleFunc("/secret", handleSecret)
	mux.HandleFunc("/secret/deeper", handleSecretDeeper)
	mux.HandleFunc("/secret/deeper/rabbit-hole", handleRabbitHole)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Printf("Listening on :%s\n", port)
	fmt.Println("May your uptime be long and your errors be few.")
	if err := http.ListenAndServe(":"+port, middleware(mux)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
