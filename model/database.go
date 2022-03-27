package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CardList is a struct that represents the default database table
type Cards struct {
	ID            int           `gorm:"type:integer; primary key; not null" json:"id"`
	Name          string        `gorm:"type: text; not null" json:"name"`
	NamePt        string        `gorm:"type: text; not null" json:"name_pt"`
	NameFr        string        `gorm:"type: text; not null" json:"name_fr"`
	Type          string        `gorm:"type: text; not null" json:"type"`
	Description   string        `gorm:"type: text; not null" json:"description"`
	DescriptionPt string        `gorm:"type: text; not null" json:"description_pt"`
	DescriptionFr string        `gorm:"type: text; not null" json:"description_fr"`
	Image         pq.Int32Array `gorm:"type: integer array; not null" json:"image"`
	Attribute     string        `gorm:"type: text; not null" json:"attribute"`
	Race          string        `gorm:"type: text; not null" json:"race"`
	Archetype     string        `gorm:"type: text; not null" json:"archetype"`
	Price         string        `gorm:"type: float; not null" json:"price"`
	Atk           int           `gorm:"type: integer; not null" json:"atk"`
	Def           int           `gorm:"type: integer; not null" json:"def"`
	Level         int           `gorm:"type: integer; not null" json:"level"`
}

type Portfolios struct {
	ID          uuid.UUID `gorm:"type:uuid; primary key; not null; default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type: text; not null" json:"name"`
	Description string    `gorm:"type: text;" json:"description"`
	Cover       string    `gorm:"type: text;" json:"cover"`
	CreatedAt   time.Time `gorm:"type:timestamp; not null; default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp; not null; default:now()" json:"updated_at"`
}

type PortfolioCards struct {
	Card      Cards      `gorm:"foreignKey: ID; type: integer; not null" json:"card_id"`
	Portfolio Portfolios `gorm:"foreignKey: ID; type: uuid; not null" json:"portfolio_id"`
}
