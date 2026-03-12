run:
    wails3 build
    ./bin/veles

db:
    sqlite3 -box $HOME/Library/Application\ Support/veles/application.sqlite

remove_db:
    rm $HOME/Library/Application\ Support/veles/application.sqlite