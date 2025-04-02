/************************************************************
*  Author:         Nicholas Tibbetts
*  Date:           01/27/2025 T15:46:47
*  Description:    _
*  Resources:      Distributed services with go by travis jeffery
*				   Build an orchestrator in go by tim boring
*				   Shipping go by joel holmes
*				   select books by john arundel
*				   https://github.com/anthdm/hollywood
*  :               _
***********************************************************/

package main

import (
	"arena-tactics/internal/database"
	"arena-tactics/internal/game"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// websocket setup for real time communication.
// upgrader that takes care of turning http connections into websockets.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// allow all origins for development - obviously we will lock this down in production!
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 60 updates per second - same as most modern games.
// smooth movement without killing the cpu. maybe make it based on hardware instead of a magic number
const tickRate = time.Second / 60

// bullet type represents a projectile in the game.
// these are made super simple - just position, direction, and lifetime.
// No need for complex physics here just basic trigonometry.
type Bullet struct {
	ID       string
	PlayerID string // need track who fired it for damage attribution
	Position game.Position
	Rotation float64 // direction in radians easier for trig calculations
	Speed    float64
	Lifetime float64 // seconds remaining before disappearing
}

// weapons that players can pick up. should be modular so we can easily add new types later.
type Weapon struct {
	ID        string
	Type      string // "pistol", "shotgun", etc
	Position  game.Position
	SpawnTime time.Time // just used for despawning old weapons
}

// power add some strategic depth to the gameplay.
// players can choose between mobility, defense, or healing.
type PowerUp struct {
	ID        string
	Type      string // "teleportation", "force_field", or "health_regen"
	Position  game.Position
	SpawnTime time.Time
}

// The damage/speed/fire rate balance is crucial for gameplay (some stuff needs to be reworked though)
// Shotgun does more total damage but has spread, machine gun is rapid but weak, etc.
var WeaponProperties = map[string]struct {
	FireRate float64
	Damage   int
	Speed    float64
	Lifetime float64
}{
	"pistol": {
		FireRate: 500, // milliseconds between shots
		Damage:   21,  // Decent mid range damage
		Speed:    550, // medium bullet speed
		Lifetime: 1.2, // how long bullets exist (in seconds)
	},
	"shotgun": {
		FireRate: 1000, // Slow fire rate to balance the multiple pellets
		Damage:   14,   // per pellet! Total damage is 14*8 = 112 if all hit
		Speed:    500,
		Lifetime: 0.5, // short range is key for shotgun balance
	},
	"machine_gun": {
		FireRate: 100, // super fast firing 10 shots per second!
		Damage:   7,   // low damage per bullet but high dps
		Speed:    700, // fastest projectiles
		Lifetime: 0.725,
	},
	"rocket_launcher": {
		FireRate: 1000, // slow firing this thing is powerful (still need to add it)
		Damage:   40,   // high damage
		Speed:    400,  // slower projectiles for balance
		Lifetime: 1.5,  // longer range
	},
	"laser": {
		FireRate: 20,   // super fast firing 10 shots per second!
		Damage:   2,    // low damage per bullet but high dps
		Speed:    2500, // fastest projectiles
		Lifetime: 2.0,
	},
}

// single message struct with optional fields to handle
// all different message types keeps the protocol simpler than having
// dozens of different message structs.

type Message struct {
	Type              string         `json:"type"`
	Position          game.Position  `json:"position,omitempty"`
	PlayerID          string         `json:"player_id,omitempty"`
	Color             int            `json:"color,omitempty"`
	SessionID         string         `json:"session_id,omitempty"`
	Weapon            string         `json:"weapon,omitempty"`
	WeaponID          string         `json:"weapon_id,omitempty"`
	Rotation          float64        `json:"rotation,omitempty"`
	Health            int            `json:"health,omitempty"`
	IsDead            bool           `json:"is_dead,omitempty"`
	DeathTime         int64          `json:"death_time,omitempty"`
	PowerUp           string         `json:"powerup,omitempty"`
	PowerUpID         string         `json:"powerup_id,omitempty"`
	CursorPos         *game.Position `json:"cursorPos,omitempty"`
	TeleportAvailable bool           `json:"teleportAvailable,omitempty"`
	ForceFieldActive  bool           `json:"forceFieldActive,omitempty"`
	HealthRegenActive bool           `json:"healthRegenActive,omitempty"`
	LobbyUpdate       []string       `json:"lobby_update,omitempty"`
	MatchStarted      bool           `json:"match_started,omitempty"`
}

// Essentially the core of the game its the state container that holds everything.
// Using maps for o(1) lookup by id, which matters when we've got dozens of
// entities and need to check collisions every frame.
type GameState struct {
	players     map[string]*game.Player
	bullets     map[string]*Bullet
	weapons     map[string]*Weapon
	powerups    map[string]*PowerUp
	mutex       sync.RWMutex // super crucial mutex since we're accessing state from multiple goroutines
	matchActive bool
}

// broadcasts a message to all connected players.
// using this pattern a lot it's cleaner than repeating the loop everywhere.
func (gs *GameState) broadcast(message Message) {
	for _, p := range gs.players {
		// sends message to all connected players
		if p.Conn != nil {
			p.WriteMu.Lock()
			p.Conn.WriteJSON(message)
			p.WriteMu.Unlock()
		}
	}
}

// creates a new weapon at a random position.
// might seem simple but this little function prevents a lot of code duplication.
func spawnWeapon() *Weapon {
	weapons := []string{"pistol", "shotgun", "machine_gun"}
	return &Weapon{
		ID:   generateID(),
		Type: weapons[rand.Intn(len(weapons))], // random weapon type for variety
		Position: game.Position{
			X: rand.Float64() * 1000, // and random position within game bounds
			Y: rand.Float64() * 600,
		},
		SpawnTime: time.Now(),
	}
}

// creates a power up at a random position.
// they should be rarer than weapons and add strategic depth to gameplay.
func spawnPowerUp() *PowerUp {
	types := []string{"teleportation", "force_field", "health_regen"}
	powerup := &PowerUp{
		ID:        generateID(),
		Type:      types[rand.Intn(len(types))],
		Position:  game.Position{X: rand.Float64() * 1000, Y: rand.Float64() * 600},
		SpawnTime: time.Now(),
	}
	log.Printf("Spawning power-up: %+v", powerup)
	return powerup
}

// factory function for creating a fresh game state.
// they make the initialization intent very clear
// and ensure i don't forget to initialize any maps.
func newGameState() *GameState {
	return &GameState{
		players:  make(map[string]*game.Player),
		weapons:  make(map[string]*Weapon),
		bullets:  make(map[string]*Bullet),
		powerups: make(map[string]*PowerUp),
	}
}

// many bugs were caused by nan positions...
func isValidPosition(pos game.Position) bool {
	return !math.IsNaN(pos.X) && !math.IsNaN(pos.Y) &&
		math.Abs(pos.X) < 1e6 && math.Abs(pos.Y) < 1e6
}

// concept function used for the lobby UI to show who's waiting to play.
// func getConnectedPlayers(gs *GameState) []string {
// 	var players []string
// 	for _, p := range gs.players {
// 		players = append(players, p.SessionID)
// 	}
// 	return players
// }

// This is the heart of the server handles websockets connections and
// manages the entire lifecycle of a player's connection.
// it's a big function but it handles a lot: connection setup, message processing,
// and eventually disconnection.
func handleWebSocket(gameState *GameState, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	// tried detecting incognito mode but its unreliable:
	// isIncognito := strings.Contains(r.Header.Get("User-Agent"), "Incognito")
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	// Create a done channel for cleanup
	done := make(chan struct{})
	defer close(done)

	// Start physics loop in goroutine

	// Run weapon spawning in a separate goroutine

	// Read initial session message
	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Error reading initial message: %v", err)
		return
	}

	var initMessage Message
	if err := json.Unmarshal(messageBytes, &initMessage); err != nil {
		log.Printf("Error parsing initial message: %v", err)
		return
	}

	// (thread safety) locking before touching shared state.
	gameState.mutex.Lock()
	var player *game.Player

	// check if this is a returning player by looking for their session id.
	// will let players refresh the page without losing their character.
	// there character is technically ephemeral in the sense they are removed from the database on disconnect.
	// but for continuity if they refresh the page we want them to have the same character as before (color, position, weapon, etc).
	for _, p := range gameState.players {
		if p.SessionID == initMessage.SessionID {
			// found existing player thus handle reconnection
			// close any existing connection can't have two connections for one player!
			if p.Conn != nil {
				p.Conn.Close()
			}

			// update connection
			p.Conn = conn
			player = p
			// Sanity check position
			if !isValidPosition(p.Position) {
				p.Position = game.Position{
					X: 500,
					Y: 300,
				}
			}
			// handle reconnection for dead players specially
			if p.IsDead {
				if !isValidPosition(p.Position) {
					p.Position = game.Position{
						X: 500,
						Y: 300,
					}
				}
				if math.IsNaN(p.Position.X) || math.IsNaN(p.Position.Y) {
					p.Position = game.Position{X: 500, Y: 300}
				}

				// let the player know they're still dead
				p.WriteMu.Lock()
				conn.WriteJSON(Message{
					Type:      "player_death",
					PlayerID:  p.ID,
					Position:  p.Position,
					DeathTime: p.DeathTime.Unix(),
				})
				p.WriteMu.Unlock()

				// manages the death timer for reconnects.
				// needs to calculate the remaining time on the death timer.
				if p.DeathTimer != nil {
					p.DeathTimer.Stop()
					p.DeathTimer = nil
				}

				// calculate time since death
				timeSinceDeath := time.Since(p.DeathTime)
				if timeSinceDeath < 3*time.Second {

					// resume the timer with remaining time
					remaining := 3*time.Second - timeSinceDeath
					p.DeathTimer = time.AfterFunc(remaining, func() {
						respawnPlayer(gameState, p)
					})
				} else {
					// respawn immediately if timer expired
					respawnPlayer(gameState, p)
				}
			}

			// sends the current state to reconnected player
			p.WriteMu.Lock()
			conn.WriteJSON(Message{
				Type:              "player_init",
				PlayerID:          p.ID,
				Color:             p.Color,
				Position:          p.Position,
				Health:            p.Health,
				Weapon:            p.Weapon,
				IsDead:            p.IsDead,
				DeathTime:         p.DeathTime.Unix(),
				TeleportAvailable: p.TeleportAvailable,
				ForceFieldActive:  p.ForceFieldActive,
				HealthRegenActive: p.HealthRegenActive,
			})
			p.WriteMu.Unlock()
			log.Printf("Player %s reconnected with session %s at position %v",
				p.ID, p.SessionID, p.Position)
			// the code is duplicated but its ok because the
			// logic is slightly different for reconnection vs. initial setup
			if p.IsDead {
				timeSinceDeath := time.Since(p.DeathTime)
				if timeSinceDeath >= 3*time.Second {
					// force respawn if timer expired during reload
					respawnPlayer(gameState, p)
				} else {
					// send current state before creating new timer
					p.SendMessage(Message{
						Type:      "player_death",
						PlayerID:  p.ID,
						Position:  p.Position,
						DeathTime: p.DeathTime.Unix(),
					})

					// create new timer with remaining time
					remaining := 3*time.Second - timeSinceDeath
					p.DeathTimer = time.AfterFunc(remaining, func() {
						respawnPlayer(gameState, p)
					})
				}
			}
			break

		}

	}
	// no existing player found (so create a new one)
	// this is where first time players enter the game.
	if player == nil {
		// get or create player from database this lets us persist stats*
		dbPlayer, err := database.GetOrCreatePlayer(initMessage.SessionID)
		if err != nil {
			log.Printf("Error getting or creating player: %v", err)
			gameState.mutex.Unlock()
			return
		}
		// in game fields that arent persisted
		dbPlayer.Conn = conn
		dbPlayer.Health = 100
		dbPlayer.Weapon = "pistol" // everyone starts with the basic pistol
		dbPlayer.IsDead = false
		dbPlayer.Position = game.Position{X: 500, Y: 300} // center of the map
		dbPlayer.Color = rand.Intn(0xFFFFFF)              // random color to distinguish players
		dbPlayer.WriteMu = sync.Mutex{}

		// adds the player to our game state
		gameState.players[dbPlayer.ID] = dbPlayer
		player = dbPlayer
		// Get last known position if it exists
		lastPos, err := database.GetLastKnownPosition(initMessage.SessionID)
		if err != nil {
			log.Printf("Error retrieving last known position: %v", err)
		}
		if lastPos != nil {
			player.Position = *lastPos
		} else {
			player.Position = game.Position{X: 500, Y: 300} // Default spawn
		}
	}

	// Send initialization message.
	player.WriteMu.Lock()
	initResp := Message{
		Type:      "player_init",
		PlayerID:  player.ID,
		Color:     player.Color,
		Position:  player.Position,
		Health:    player.Health,
		Weapon:    player.Weapon,
		IsDead:    player.IsDead,
		DeathTime: player.DeathTime.Unix(),
	}
	err = conn.WriteJSON(initResp)
	player.WriteMu.Unlock()
	// if gameState.matchActive {
	// 	player.SendMessage(Message{Type: "match_started", MatchStarted: true})
	// }

	if err != nil {
		log.Printf("Error sending initial state: %v", err)
		gameState.mutex.Unlock()
		return
	}

	// send all existing weapons to the player they need to see what is already on the map
	for id, weapon := range gameState.weapons {
		player.SendMessage(Message{
			Type:     "weapon_spawn",
			WeaponID: id,
			Position: weapon.Position,
			Weapon:   weapon.Type,
		})
	}
	// and same for powerups
	for id, powerup := range gameState.powerups {
		player.SendMessage(Message{
			Type:      "powerup_spawn",
			PowerUpID: id,
			Position:  powerup.Position,
			PowerUp:   powerup.Type,
		})
	}

	//set up done release the lock
	gameState.mutex.Unlock()

	// main message loop this runs until the player disconnect
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection error for player %s: %v", player.ID, err)

			gameState.mutex.Lock()
			if player.Conn == conn {
				player.Conn = nil
				// wait 5 seconds before actually removing the player
				// this gives them a chance to reconnect without losing their character
				go gameState.removePlayer(player.ID)
			}
			gameState.mutex.Unlock()
			return
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			continue
		}
		gameState.mutex.Lock()
		// Process the different message types
		switch message.Type {
		case "position":
			// Only log if player is alive and position is valid
			if !player.IsDead && isValidPosition(message.Position) {
				player.PendingPosition = &message.Position
				player.LastKnownPosition = message.Position

				// Log movement to DB
				go func(sessionID, playerID string, pos game.Position) {
					if err := database.InsertPlayerPosition(sessionID, playerID, pos); err != nil {
						log.Printf("Error inserting player position: %v", err)
					}
					if err := database.UpdateLastKnownPosition(sessionID, playerID, pos); err != nil {
						log.Printf("Error updating last known position: %v", err)
					}
				}(player.SessionID, player.ID, message.Position)
			}

		case "shoot":
			// dead players can't shoot
			if player.IsDead {
				break
			}
			props := WeaponProperties[player.Weapon]
			// different weapons create different bullet patterns
			switch player.Weapon {
			case "shotgun":
				// shotgun shoots multiple pellets in a spread pattern
				for i := -3; i <= 4; i++ {
					bullet := &Bullet{
						ID:       generateID(),
						PlayerID: player.ID,
						Position: player.Position,
						Rotation: message.Rotation + (float64(i)*5.0)*(math.Pi/180),
						Speed:    props.Speed,
						Lifetime: props.Lifetime,
					}
					gameState.bullets[bullet.ID] = bullet
				}
			case "machine_gun":
				// machine gun just shoots one bullet but very rapidly
				bullet := &Bullet{
					ID:       generateID(),
					PlayerID: player.ID,
					Position: player.Position,
					Rotation: message.Rotation,
					Speed:    props.Speed,
					Lifetime: props.Lifetime,
				}
				gameState.bullets[bullet.ID] = bullet
			default:
				// default single bullet for pistol and other weapons
				bullet := &Bullet{
					ID:       generateID(),
					PlayerID: player.ID,
					Position: player.Position,
					Rotation: message.Rotation,
					Speed:    props.Speed,
					Lifetime: props.Lifetime,
				}
				gameState.bullets[bullet.ID] = bullet
			}
		case "teleport":
			// teleportation power up usage
			if !player.IsDead && player.TeleportAvailable && message.CursorPos != nil {
				// calculates direction vector to cursor
				dx := message.CursorPos.X - player.Position.X
				dy := message.CursorPos.Y - player.Position.Y
				angle := math.Atan2(dy, dx)

				// teleports a fixed distance in said direction
				offset := 100.0
				newPos := game.Position{X: player.Position.X + math.Cos(angle)*offset, Y: player.Position.Y + math.Sin(angle)*offset}
				if isValidPosition(newPos) {
					player.Position = newPos
					gameState.broadcast(Message{
						Type:     "teleport",
						PlayerID: player.ID,
						Position: player.Position,
					})
				}
			}

			// case "join_match":
			// 	if gameState.matchActive {
			// 		player.SendMessage(Message{Type: "match_started", MatchStarted: true})
			// 		log.Printf("player %s joined the active match", player.ID)
			// 	} else {
			// 		player.SendMessage(Message{Type: "lobby_update", LobbyUpdate: getConnectedPlayers(gameState)})
			// 		log.Printf("player %s requested to join, but no match is active", player.ID)
			// 	}
			// case "start_match":
			// 	if !gameState.matchActive {
			// 		gameState.matchActive = true
			// 		gameState.broadcast(Message{Type: "match_started", MatchStarted: true})
			// 		log.Println("Match is now active!")
			// 	}
		}
		gameState.mutex.Unlock()

	}
}


// Batch insert function
func batchInsertPlayerPositions(gs *GameState, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gs.mutex.Lock()
			var values []string
			for _, p := range gs.players {
				values = append(values, fmt.Sprintf(
					"('%s', '%s', %.2f, %.2f, NOW())",
					p.SessionID, p.ID, p.Position.X, p.Position.Y,
				))
			}
			if len(values) > 0 {
				query := fmt.Sprintf("INSERT INTO player_positions (session_id, player_id, x, y, timestamp) VALUES %s", strings.Join(values, ","))
				_, err := database.GetDB().Exec(query)
				if err != nil {
					log.Printf("Error batch inserting player positions: %v", err)
				}
			}
			gs.mutex.Unlock()
		}
	}
}


// Resets their state and starts the respawn timer
func handlePlayerDeath(gameState *GameState, p *game.Player, position game.Position) {
	p.IsDead = true
	p.Health = 0
	p.DeathTime = time.Now()
	p.Position = position
	// take away all power ups on death (no keeping your goodies)
	p.TeleportAvailable = false
	p.ForceFieldActive = false
	p.Shield = 0
	p.HealthRegenActive = false
	p.RegenExpiry = time.Time{}
	log.Printf("Player %s died, setting death state", p.ID)

	// Cancel any existing death timer
	if p.DeathTimer != nil {
		p.DeathTimer.Stop() // stop the existing timer
		p.DeathTimer = nil
	}

	p.DeathTimer = time.AfterFunc(3*time.Second, func() {
		respawnPlayer(gameState, p)
	})

	// broadcast death
	gameState.broadcast(Message{
		Type:      "player_death",
		PlayerID:  p.ID,
		Position:  p.Position,
		DeathTime: p.DeathTime.Unix(),
	})

}

// Respawns the player after they've been dead for the required time
func respawnPlayer(gameState *GameState, p *game.Player) {
	gameState.mutex.Lock()
	defer gameState.mutex.Unlock()

	// respawnDuration := 3 * time.Second
	// if p.isIncognito {
	//     respawnDuration = 2 * time.Second
	// }

	// cleans up all death related state first
	if p.DeathTimer != nil {
		p.DeathTimer.Stop()
		p.DeathTimer = nil
	}

	// Check connection before respawn ie if they disconnected while dead don't respawn them
	if p.Conn == nil {
		delete(gameState.players, p.ID)
		gameState.broadcast(Message{
			Type:     "player_disconnect",
			PlayerID: p.ID,
		})
		return
	}
	// Reset power ups
	p.TeleportAvailable = false
	p.ForceFieldActive = false
	p.Shield = 0
	p.HealthRegenActive = false
	// spawnPoint := getRandomSpawnPoint(gameState)

	// Reset player state for new character
	p.IsDead = false
	p.Health = 100
	p.Weapon = "pistol"
	// p.PendingPosition = &spawnPoint
	// p.deathTimer = nil
	p.DeathTime = time.Time{}
	p.Position = getRandomSpawnPoint(gameState)

	// Pre-calculate spawn point

	// broadcast respawn with position
	gameState.broadcast(Message{
		Type:     "player_respawn",
		PlayerID: p.ID,
		Position: p.Position,
		Health:   p.Health,
		Weapon:   p.Weapon,
	})

	p.SendMessage(Message{
		Type:     "player_respawn",
		PlayerID: p.ID,
		Position: p.Position,
		Health:   p.Health,
		Weapon:   p.Weapon,
	})
}

// Finds a safe spawn point away from other players
// prevents spawning on top of each other
func getRandomSpawnPoint(gameState *GameState) game.Position {
	spawnPoints := []game.Position{
		{X: 100, Y: 100},
		{X: 800, Y: 500},
		{X: 400, Y: 300},
	}

	// keep trying until we find a safe spawn point
	for {
		point := spawnPoints[rand.Intn(len(spawnPoints))]
		safe := true

		// check if any living players are too close
		for _, p := range gameState.players {
			if p.IsDead {
				continue
			}
			dx := point.X - p.Position.X
			dy := point.Y - p.Position.Y
			if math.Sqrt(dx*dx+dy*dy) < 100 { // minimum safe distance
				safe = false
				break
			}
		}

		if safe {
			return point
		}
		// if not safe then loop will try another point
	}
}

// generates a unique id for game entities
func generateID() string {
	// Just a simple prefix + random string for now
	// in a production game we'd use UUIDs
	return "player-" + randomString(8)
}

// generate random string of specified length
func randomString(length int) string {
	//just random letters
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// remove a player after they've been disconnected for a some time
func (gs *GameState) removePlayer(id string) {
	// this gives them a chance to reconnect
	time.Sleep(5 * time.Second)

	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	player, exists := gs.players[id]
	if !exists || player.Conn != nil {
		// player either doesnt exist or has reconnected
		return
	}

	// removes player from database to free up the session ID
	if err := database.RemovePlayer(player.SessionID); err != nil {
		log.Printf("Error removing player %s from DB: %v", player.SessionID, err)
	}

	// clean up any active timers
	if player.DeathTimer != nil {
		player.DeathTimer.Stop()
	}

	// and lastly remove from game state and notify
	delete(gs.players, id)
	gs.broadcast(Message{
		Type:     "player_disconnect",
		PlayerID: id,
	})
}

// initialize environment variables and configuration
func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
	}
	// Go up two levels to reach the project root
	envPath := filepath.Join(wd, "..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		// wont be found in production thus k8 default variables will be injected instead
		log.Printf("No .env file found at %s, using kubernetes system environment variables", envPath)
	} else {
		log.Printf(".env loaded successfully from %s", envPath)
	}
	log.Printf("DB_USERNAME: %s", os.Getenv("DB_USERNAME"))
}

// orchestrates the game server
func main() {

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	gameState := newGameState()
	gameState.matchActive = true

	  // Start the batch insertion goroutine for player positions.
    // This will flush player position data every 500ms.
    go batchInsertPlayerPositions(gameState, 500*time.Millisecond)

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	// channel for clean up
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-ticker.C:
				gameState.mutex.Lock()
				if !gameState.matchActive {
					// skips the entire update if match not active
					gameState.mutex.Unlock()
					continue
				}

				// updates the bullets
				for id, bullet := range gameState.bullets {
					// moves them
					bullet.Position.X += math.Cos(bullet.Rotation) * bullet.Speed * float64(tickRate) / float64(time.Second)
					bullet.Position.Y += math.Sin(bullet.Rotation) * bullet.Speed * float64(tickRate) / float64(time.Second)

					// bullet collisions with players
					for _, p := range gameState.players {
						// will skip if this is the shooter or if the target is already dead
						if p.ID == bullet.PlayerID || p.IsDead {
							continue
						}
						if p.ID != bullet.PlayerID {
							// super simple circle collision check
							dx := bullet.Position.X - p.Position.X
							dy := bullet.Position.Y - p.Position.Y
							distance := math.Sqrt(dx*dx + dy*dy)
							// In bullet collision check
							if distance < 20 { // collision detected (20 is the radius of the player)
								// then get the shooter (the player who fired the bullet)
								shooter, exists := gameState.players[bullet.PlayerID]
								if !exists {
									// if for some reason the shooter is not found just skip damage calculation.
									continue
								}
								// gets the weapon properties for the shooter’s current weapon.
								props := WeaponProperties[shooter.Weapon]

								if p.ForceFieldActive && p.Shield > 0 {
									if p.Shield >= props.Damage {
										p.Shield -= props.Damage
									} else {
										// Not enough shield to block full damage.
										leftover := props.Damage - p.Shield
										p.Shield = 0
										p.ForceFieldActive = false
										newHealth := p.Health - leftover
										if newHealth < 0 {
											newHealth = 0
										}
										p.Health = newHealth
										if p.Health <= 0 {
											handlePlayerDeath(gameState, p, p.LastKnownPosition)
											// finds the shooter (the player who fired the bullet)
											shooter, exists := gameState.players[bullet.PlayerID]
											if exists {
												// update killer stats: increment kills by 1.
												if err := database.UpdatePlayerStats(shooter.SessionID, 1, 0); err != nil {
													log.Printf("Error updating killer stats: %v", err)
												}
											}
											// update victim stats: increment deaths by 1.
											if err := database.UpdatePlayerStats(p.SessionID, 0, 1); err != nil {
												log.Printf("Error updating victim stats: %v", err)
											}
										}
									}
								} else {
									// no active shield thus apply damage directly.
									newHealth := p.Health - props.Damage
									if newHealth < 0 {
										newHealth = 0
									}
									p.Health = newHealth
									if p.Health <= 0 {
										handlePlayerDeath(gameState, p, p.LastKnownPosition)
										shooter, exists := gameState.players[bullet.PlayerID]
										if exists {
											if err := database.UpdatePlayerStats(shooter.SessionID, 1, 0); err != nil {
												log.Printf("Error updating killer stats: %v", err)
											}
										}
										if err := database.UpdatePlayerStats(p.SessionID, 0, 1); err != nil {
											log.Printf("Error updating victim stats: %v", err)
										}
									}
								}
								gameState.broadcast(Message{
									Type:     "health_update",
									PlayerID: p.ID,
									Health:   p.Health,
								})

								bulletDirX := math.Cos(bullet.Rotation)
								bulletDirY := math.Sin(bullet.Rotation)

								// Apply impulse
								impulse := 100.0 // Adjust as needed
								p.Velocity.X += bulletDirX * impulse
								p.Velocity.Y += bulletDirY * impulse

								delete(gameState.bullets, id)
								break
							}
						}

						// Applies velocity
						p.Position.X += p.Velocity.X * float64(tickRate) / float64(time.Second)
						p.Position.Y += p.Velocity.Y * float64(tickRate) / float64(time.Second)

						// and damping
						p.Velocity.X *= 0.9
						p.Velocity.Y *= 0.9

						// Clamp position if needed
						if !isValidPosition(p.Position) {
							p.Position = p.LastKnownPosition
						}
					}

					bullet.Lifetime -= float64(tickRate) / float64(time.Second)
					if bullet.Lifetime <= 0 {
						delete(gameState.bullets, id)
					}
				} // Right after bullet updates
				for _, bullet := range gameState.bullets {
					gameState.broadcast(Message{
						Type:     "bullet_update",
						PlayerID: bullet.PlayerID,
						Position: bullet.Position,
					})
				}

				// after the bullet updates adds the weapon pickup check
				for id, weapon := range gameState.weapons {

					if time.Since(weapon.SpawnTime) > 30*time.Second { // 30 second despawn timer
						delete(gameState.weapons, id)
						gameState.broadcast(Message{
							Type:     "weapon_despawn",
							WeaponID: id,
						})
						continue
					}
					for playerID, p := range gameState.players {
						dx := weapon.Position.X - p.Position.X
						dy := weapon.Position.Y - p.Position.Y
						if !isValidPosition(weapon.Position) {
							delete(gameState.weapons, id) // removes invalid weapons
							continue
						}
						distance := math.Sqrt(dx*dx + dy*dy)

						if distance < 20 { // player radius + weapon radius
							// don't pick up if it's the same weapon type
							if p.Weapon == weapon.Type {
								continue
							}

							// store old weapon ID if player has one
							oldWeaponType := p.Weapon

							// assign new weapon to player
							p.Weapon = weapon.Type
							delete(gameState.weapons, id)

							gameState.broadcast(Message{
								Type:     "weapon_pickup",
								PlayerID: playerID,
								WeaponID: id,
								Weapon:   weapon.Type,
							})

							// only spawns old weapon if it was different
							if oldWeaponType != "" && oldWeaponType != weapon.Type {
								newWeapon := &Weapon{
									ID:   generateID(),
									Type: oldWeaponType,
									Position: game.Position{
										X: p.Position.X + 30,
										Y: p.Position.Y,
									},
								}
								gameState.weapons[newWeapon.ID] = newWeapon
								gameState.broadcast(Message{
									Type:     "weapon_spawn",
									WeaponID: newWeapon.ID,
									Position: newWeapon.Position,
									Weapon:   newWeapon.Type,
								})
							}
							break
						}
					}
				}

				for id, powerup := range gameState.powerups {
					// removes expired powerups (after 30 seconds)
					if time.Since(powerup.SpawnTime) > 30*time.Second {
						delete(gameState.powerups, id)
						gameState.broadcast(Message{
							Type:      "powerup_despawn",
							PowerUpID: id,
						})
						continue
					}
					for playerID, p := range gameState.players {
						dx := powerup.Position.X - p.Position.X
						dy := powerup.Position.Y - p.Position.Y
						if !isValidPosition(powerup.Position) {
							delete(gameState.powerups, id)
							continue
						}
						distance := math.Sqrt(dx*dx + dy*dy)
						if distance < 20 { // collision detected
							p.TeleportAvailable = false
							p.ForceFieldActive = false
							p.HealthRegenActive = false
							switch powerup.Type {
							case "teleportation":
								if !p.TeleportAvailable { // prevents stacking
									p.TeleportAvailable = true
									gameState.broadcast(Message{
										Type:      "powerup_pickup",
										PlayerID:  playerID,
										PowerUpID: id,
										PowerUp:   powerup.Type,
									})
								}
							case "force_field":
								// activate force field.
								p.ForceFieldActive = true
								p.Shield = 100 // maximum shield value
								gameState.broadcast(Message{
									Type:      "powerup_pickup",
									PlayerID:  playerID,
									PowerUpID: id,
									PowerUp:   powerup.Type,
								})
								//not working
							case "health_regen":
								p.HealthRegenActive = true
								p.RegenExpiry = time.Now().Add(10 * time.Second)
								p.HealthRegenAccumulator = 0 // reset accumulator
								gameState.broadcast(Message{
									Type:      "powerup_pickup",
									PlayerID:  playerID,
									PowerUpID: id,
									PowerUp:   powerup.Type,
								})
							}
							delete(gameState.powerups, id)
							break
						}
					}
				}

				changed := false

				for _, p := range gameState.players {
					if p.PendingPosition != nil {
						if isValidPosition(*p.PendingPosition) {
							p.Position = *p.PendingPosition
						} else {
							p.Position = game.Position{X: 500, Y: 300} // reset to default
						}
						p.PendingPosition = nil
						changed = true
					}
				}
				for id1, player1 := range gameState.players {
					for id2, player2 := range gameState.players {
						if id1 == id2 {
							continue
						}
						const pushForce = 1.5

						collider1 := &game.CircleCollider{
							X:      player1.Position.X,
							Y:      player1.Position.Y,
							Radius: 15,
						}
						collider2 := &game.CircleCollider{
							X:      player2.Position.X,
							Y:      player2.Position.Y,
							Radius: 15,
						}
						// collision check
						px1, py1, px2, py2 := game.ResolveCollision(collider1, collider2, pushForce)
						if !math.IsNaN(px1) && !math.IsNaN(py1) && !math.IsNaN(px2) && !math.IsNaN(py2) {
							player1.Position.X += px1
							player1.Position.Y += py1
							player2.Position.X += px2
							player2.Position.Y += py2
						}
						// check for if players are too close
						if px1 != 0 || py1 != 0 {
							player1.Position.X += px1
							player1.Position.Y += py1
							player2.Position.X += px2
							player2.Position.Y += py2
							log.Printf("Collision resolved: Player %s at %+v, Player %s at %+v",
								id1, player1.Position, id2, player2.Position)

							// new positions
							gameState.broadcast(Message{
								Type:     "position_update",
								PlayerID: id1,
								Position: player1.Position,
								Color:    player1.Color,
							})
							gameState.broadcast(Message{
								Type:     "position_update",
								PlayerID: id2,
								Position: player2.Position,
								Color:    player2.Color,
							})
						}
					}
				}
				if changed {
					for id, p := range gameState.players {
						gameState.broadcast(Message{
							Type:     "position_update",
							PlayerID: id,
							Position: p.Position,
							Color:    p.Color,
						})
					}
				}
				// health regeneration logic
				for _, p := range gameState.players {
					if p.HealthRegenActive && p.Health < 100 {
						if time.Now().After(p.RegenExpiry) {
							p.HealthRegenActive = false
							p.HealthRegenAccumulator = 0 // reset accumulator when a power up expires
						} else {
							regenPerTick := 10 * (float64(tickRate) / float64(time.Second)) // ≈ 0.16
							p.HealthRegenAccumulator += regenPerTick
							if p.HealthRegenAccumulator >= 1.0 {
								increment := int(p.HealthRegenAccumulator) // whole number part
								p.Health += increment
								p.HealthRegenAccumulator -= float64(increment) // keeping the remainder
								if p.Health > 100 {
									p.Health = 100
									p.HealthRegenAccumulator = 0 // resets if max health reached
								}
								log.Printf("Player %s health regenerated to %d", p.ID, p.Health)
								gameState.broadcast(Message{
									Type:     "health_update",
									PlayerID: p.ID,
									Health:   p.Health,
								})
							}
							// logs every tick for debugging
							log.Printf("health regen tick for player %s: health=%d, regenActive=%t, accumulator=%f",
								p.ID, p.Health, p.HealthRegenActive, p.HealthRegenAccumulator)
						}
					}
				}

				gameState.mutex.Unlock()
			case <-done:
				return
			}
		}
	}()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			gameState.mutex.Lock()
			if !gameState.matchActive {
				// If the match isn’t active, skip spawning
				gameState.mutex.Unlock()
				continue
			}

			if len(gameState.weapons) < 5 { // Max 5 weapons at once
				weapon := spawnWeapon()
				gameState.weapons[weapon.ID] = weapon
				log.Printf("Spawning weapon: %+v", weapon)
				gameState.broadcast(Message{
					Type:     "weapon_spawn",
					WeaponID: weapon.ID,
					Position: weapon.Position,
					Weapon:   weapon.Type,
				})
			}
			gameState.mutex.Unlock()
		}
	}()

	log.Println("Power-up spawning goroutine started")
	// separate goroutine for power up spawning
	go func() {
		for {
			delay := time.Duration(15+rand.Intn(30)) * time.Second
			time.Sleep(delay)
			gameState.mutex.Lock()
			if !gameState.matchActive {
				// if the match hasnt started yet skip spawning
				gameState.mutex.Unlock()
				continue
			}
			// spawn only if there are less than 3 on the map
			if len(gameState.powerups) < 3 {
				powerup := spawnPowerUp()
				gameState.powerups[powerup.ID] = powerup
				log.Printf("Spawning powerup: %+v", powerup)
				gameState.broadcast(Message{
					Type:      "powerup_spawn",
					PowerUpID: powerup.ID,
					Position:  powerup.Position,
					PowerUp:   powerup.Type,
				})
			}
			gameState.mutex.Unlock()
		}
	}()

	// serve static files
	http.Handle("/", http.FileServer(http.Dir("web")))

	// and webSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(gameState, w, r)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
