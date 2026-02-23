#!/bin/bash
if [ "$1" = "init" ]; then

    if [ -f "/tmp/minitwit.db" ]; then 
        echo "Database already exists."
        exit 1
    fi
    echo "Putting a database to /tmp/minitwit.db..."
    # This is currently the best we can do as the db is auto inited when not found
    timeout 30 go run ./src/main.go 
elif [ "$1" = "startprod" ]; then
     echo "Starting minitwit with production webserver..."
     nohup "$HOME"/.local/bin/gunicorn --workers 4 --timeout 120 --bind 0.0.0.0:5000 minitwit:app > /tmp/out.log 2>&1 &
elif [ "$1" = "start" ]; then
    echo "Starting minitwit..."
    nohup go run ./src/main.go > /tmp/out.log 2>&1 &
elif [ "$1" = "stop" ]; then
    echo "Stopping minitwit..."
    pkill -f minitwit
elif [ "$1" = "inspectdb" ]; then
    ./flag_tool -i | less
elif [ "$1" = "flag" ]; then
    ./flag_tool "$@"
else
  echo "I do not know this command..."
fi
