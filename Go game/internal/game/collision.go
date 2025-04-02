package game

import "math"

// circular collision boundary
type CircleCollider struct {
    X, Y   float64
    Radius float64
}

// Checks if two circles are colliding
func (c *CircleCollider) CheckCollision(other *CircleCollider) (bool, float64, float64) {
    dx := c.X - other.X
    dy := c.Y - other.Y
    distance := math.Sqrt(dx*dx + dy*dy)
    minDistance := c.Radius + other.Radius
    
    if distance < minDistance {
        //direction to push
        normalX := dx / distance
        normalY := dy / distance
        return true, normalX, normalY
    }
    
    return false, 0, 0
}

// Forces for two colliding entities
func ResolveCollision(a, b *CircleCollider, pushForce float64) (ax, ay, bx, by float64) {
    isColliding, normalX, normalY := a.CheckCollision(b)
    if !isColliding {
        return 0, 0, 0, 0
    }

    // calculate push distances
    overlap := (a.Radius + b.Radius) - math.Sqrt(math.Pow(b.X-a.X, 2) + math.Pow(b.Y-a.Y, 2))
    pushX := normalX * overlap * pushForce
    pushY := normalY * overlap * pushForce

    // returns push vectors for both entities
    return pushX, pushY, -pushX, -pushY
}
