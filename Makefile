BIN =	/Users/claude/bin/autobench
autobench: autobench.go
	go build

install: autobench
	chmod 755 /Users/claude/bin/autobench && cp autobench /Users/claude/bin/autobench && chmod 4555 /Users/claude/bin/autobench

