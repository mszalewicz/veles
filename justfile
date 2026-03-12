clean:
    rm $HOME/Library/Application\ Support/veles/application.sqlite
    rm $HOME/Library/Application\ Support/veles/log

run:
    wails3 build
    ./bin/veles

db:
    sqlite3 -box $HOME/Library/Application\ Support/veles/application.sqlite
