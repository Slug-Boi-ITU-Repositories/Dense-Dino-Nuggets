init:
	# This is not exactly one to one the same but the init is built into the context of the programming running
	go build -o minitwit ./src/main.go

build:
	gcc flag_tool.c -l sqlite3 -o flag_tool

clean:
	rm flag_tool
