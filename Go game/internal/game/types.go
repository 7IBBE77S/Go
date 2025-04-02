package game


type Position struct {
    X        float64 `json:"x"`
    Y        float64 `json:"y"`
    Rotation float64 `json:"rotation"`
}

type Velocity struct {
    X float64
    Y float64
}