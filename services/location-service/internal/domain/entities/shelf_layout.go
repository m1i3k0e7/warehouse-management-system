package entities

type ShelfLayout struct {
    ShelfID     string      `json:"shelf_id" bson:"shelf_id"`
    Rows        int         `json:"rows" bson:"rows"`
    Columns     int         `json:"columns" bson:"columns"`
    Position    Position    `json:"position" bson:"position"`
    Zones       []Zone      `json:"zones" bson:"zones"`
    UpdatedAt   time.Time   `json:"updated_at" bson:"updated_at"`
}

type Position struct {
    X float64 `json:"x" bson:"x"`
    Y float64 `json:"y" bson:"y"`
    Z float64 `json:"z" bson:"z"`
}

type Zone struct {
    ID       string `json:"id" bson:"id"`
    Name     string `json:"name" bson:"name"`
    StartRow int    `json:"start_row" bson:"start_row"`
    EndRow   int    `json:"end_row" bson:"end_row"`
    StartCol int    `json:"start_col" bson:"start_col"`
    EndCol   int    `json:"end_col" bson:"end_col"`
}