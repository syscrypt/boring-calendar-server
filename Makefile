##
# Boring Calender Server
#
# @file
# @version 0.1
EXEC      = server
BUILD_DIR = ./build
CC        = go
BLD       = build
CONFIG    = ./init/config.json
SRC       = ./cmds/server/main.go
DB		  = $(BUILD_DIR)/calendar.db
SCHEMA	  = ./init/schema.sql

all: init-db build run

build:
	@mkdir -p $(BUILD_DIR)
	$(CC) $(BLD) -o $(BUILD_DIR)/$(EXEC) $(SRC)

run:
	@$(BUILD_DIR)/$(EXEC) -config=$(CONFIG)

init-db:
	sqlite3 $(DB) < $(SCHEMA)

.PHONY: build init-db

# end
