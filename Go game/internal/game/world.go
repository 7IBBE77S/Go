package game

// type World struct {
    // Players map[string]*Player
// }

// func (w *World) UpdateCollisions() {
//     const pushForce = 1.5

//     for id1, player1 := range w.Players {
//         for id2, player2 := range w.Players {
//             if id1 == id2 {
//                 continue
//             }

//             collider1 := &CircleCollider{
//                 X: player1.Position.X,
//                 Y: player1.Position.Y,
//                 Radius: 15,
//             }
//             collider2 := &CircleCollider{
//                 X: player2.Position.X,
//                 Y: player2.Position.Y,
//                 Radius: 15,
//             }

//             px1, py1, px2, py2 := ResolveCollision(collider1, collider2, pushForce)

//             player1.Position.X += px1
//             player1.Position.Y += py1
//             player2.Position.X += px2
//             player2.Position.Y += py2
//         }
//     }
// }

