clean:
    rm $HOME/Library/Application\ Support/veles/application.sqlite
    rm $HOME/Library/Application\ Support/veles/application.sqlite-shm
    rm $HOME/Library/Application\ Support/veles/application.sqlite-wal
    rm $HOME/Library/Application\ Support/veles/log

run:
    wails3 build
    ./bin/veles

dev:
    wails3 dev

db:
    sqlite3 -box $HOME/Library/Application\ Support/veles/application.sqlite
