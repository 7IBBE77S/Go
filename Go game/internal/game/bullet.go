package game

// import (
// 	"fmt"
// 	"math"
// 	"math/rand"

// )

// type Bullet struct {
//     ID       string
//     OwnerID  string
//     Position Position
//     Velocity struct {
//         X float64
//         Y float64
//     }
//     Damage   int
//     Active   bool
// }

// func (b *Bullet) Update() {
//     b.Position.X += b.Velocity.X
//     b.Position.Y += b.Velocity.Y
// }
// func randomString(length int) string {
// 	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
// 	b := make([]byte, length)
// 	for i := range b {
// 		b[i] = letters[rand.Intn(len(letters))]
// 	}
// 	return string(b)
// }
// func NewBullet(ownerID string, startPos Position, angle float64, speed float64, damage int) *Bullet {
//     bullet := &Bullet{
//         ID:      fmt.Sprintf("bullet-%s", randomString(8)),
//         OwnerID: ownerID,
//         Damage:  damage,
//         Active:  true,
//     }
//     bullet.Position = startPos
//     bullet.Velocity.X = math.Cos(angle) * speed
//     bullet.Velocity.Y = math.Sin(angle) * speed
//     return bullet
// }