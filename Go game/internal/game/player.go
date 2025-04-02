package game

import (
	"sync"
	"time"
	"github.com/gorilla/websocket"
)

type Player struct {
	Conn                  *websocket.Conn
	ConnMu                sync.Mutex
	ID                    string
	GameID                string
	DBID                  int    
	SessionID             string 
	Username              string 
	Kills                 int   
	Deaths                int   
	Score                 int   
	IsIncognito           bool
	Position              Position
	LastKnownPosition     Position
	PendingPosition       *Position
	Color                 int
	WriteMu               sync.Mutex
	Health                int
	Weapon                string
	IsDead                bool
	DeathTime             time.Time
	DeathTimer            *time.Timer
	Velocity              Velocity
	Rotation              float64
	ForceFieldActive      bool
	Shield                int
	HealthRegenActive     bool
	RegenExpiry           time.Time
	TeleportAvailable     bool
	HealthRegenAccumulator float64
}

func (p *Player) SendMessage(msg interface{}) error {
	p.ConnMu.Lock()
	defer p.ConnMu.Unlock()

	if p.Conn != nil {
		return p.Conn.WriteJSON(msg)
	}
	return nil
}