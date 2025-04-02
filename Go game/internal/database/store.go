package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"arena-tactics/internal/game"
	_ "github.com/go-sql-driver/mysql" //u sing the blank import pattern imports the driver but doesn't use it directly
)

// global database connection
var db *sql.DB

// sets up our database connection and runs the migrations.
// env variables to keep credentials out of the codebase.
func InitDB() error {
	// now portable across different environments (ie for containerization)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_DATABASE"))
	var err error
	db, err = sql.Open("mysql", dsn) // doesnt actually connect until the first query

	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	// always ping the db to verify connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// simple migration system so the tables exist first
	if err := migrateDB(); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	return nil
}

// migrateDB creates the necessary tables if they don't already exist.
func migrateDB() error {
	queries := []string{
		`create table if not exists players (
			id int auto_increment primary key,
			session_id varchar(255) unique not null,
			username varchar(255),
			kills int default 0,
			deaths int default 0,
			score int default 0,
			last_active datetime default current_timestamp,
			created_at datetime default current_timestamp,
			updated_at datetime default current_timestamp on update current_timestamp
		);`,
		`
		create table if not exists game_sessions (
			id int auto_increment primary key,
			session_id varchar(255) unique not null,
			start_time datetime not null,
			end_time datetime,
			winner_session_id varchar(255),
			created_at datetime default current_timestamp
		);
		`,
		`
		create table if not exists pickups (
			id int auto_increment primary key,
			pickup_type varchar(50),
			spawned_at datetime default current_timestamp,
			picked_by_session_id varchar(255),
			picked_at datetime
		);`,
		`create table if not exists player_events (
			id int auto_increment primary key,
			session_id varchar(255) not null,
			event_type varchar(50) not null,
			event_time datetime default current_timestamp,
			details TEXT
		);`,
		`create table if not exists player_positions (
			id bigint auto_increment primary key,
			session_id varchar(255) not null,
			player_id varchar(255) not null,
			x float not null,
			y float not null,
			timestamp timestamp default current_timestamp,
			index (player_id, timestamp)
		);`,
		`create table if not exists player_last_positions (
			session_id varchar(255) primary key,
			player_id varchar(255) not null,
			x float not null,
			y float not null,
			updated_at timestamp default current_timestamp on update current_timestamp
		);`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error executing query: %v; query: %s", err, query)
		}
	}
	// clears out old player data on startup.
	if _, err := db.Exec("delete from players"); err != nil {
		return fmt.Errorf("failed to clear players: %v", err)
	}
	return nil
}
// direct access to the underlying database connection
func GetDB() *sql.DB {
	return db
}

// deletes a player from the database using their unique session id.
// for cleaning up when players disconnect or sessions expire. by design players are ephemeral.
func RemovePlayer(sessionID string) error {
	queries := []string{
		"delete from players where session_id = ?",
		"delete from player_positions where session_id = ?",
		"delete from player_last_positions where session_id = ?",
	}
	for _, query := range queries {
		if _, err := db.Exec(query, sessionID); err != nil {
			return fmt.Errorf("error executing query: %v; query: %s", err, query)
		}
	}
	return nil
}

// Retrieves an existing player or it will create a new one if not found.
// first try to find the player
// only create a new record if they dont exist yet. Handles both
// the database operations and mapping between db data the player struct.
func GetOrCreatePlayer(sessionID string) (*game.Player, error) {
	var player game.Player
	query := "select id, session_id, username, kills, deaths, score from players where session_id = ?"
	err := db.QueryRow(query, sessionID).Scan(&player.ID, &player.SessionID, &player.Username, &player.Kills, &player.Deaths, &player.Score)
	if err == sql.ErrNoRows {
		// player doesn't exist yet so create one
		username := "Player_" + sessionID[:8]
		insertQuery := "insert into players (session_id, username) values (?, ?)"
		res, err := db.Exec(insertQuery, sessionID, username)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		player = game.Player{
			ID:        strconv.FormatInt(id, 10),
			SessionID: sessionID,
			Username:  username,
		}
	} else if err != nil {
		return nil, err
	}
	return &player, nil
}

// increments the kills and deaths counters for a player.
func UpdatePlayerStats(sessionID string, kills, deaths int) error {
	query := "update players set kills = kills + ?, deaths = deaths + ? where session_id = ?"
	_, err := db.Exec(query, kills, deaths, sessionID)
	return err
}


// logs every movement into the player positions table
func InsertPlayerPosition(sessionID, playerID string, pos game.Position) error {
	query := `
		insert into player_positions (session_id, player_id, x, y)
		values (?, ?, ?, ?)
	`
	_, err := db.Exec(query, sessionID, playerID, pos.X, pos.Y)
	return err
}

// updates players last position
func UpdateLastKnownPosition(sessionID, playerID string, pos game.Position) error {
	query := `
		insert into player_last_positions (session_id, player_id, x, y)
		values (?, ?, ?, ?)
		on duplicate key update
		x = values(x), y = values(y), updated_at = current_timestamp
	`
	_, err := db.Exec(query, sessionID, playerID, pos.X, pos.Y)
	return err
}

//Retrieves last known position on reconnections
func GetLastKnownPosition(sessionID string) (*game.Position, error) {
	query := `
		select x, y from player_last_positions where session_id = ?
	`
	var pos game.Position
	err := db.QueryRow(query, sessionID).Scan(&pos.X, &pos.Y)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &pos, nil
}
