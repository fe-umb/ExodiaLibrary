package db

import (
	"fmt"

	"github.com/lcmps/ExodiaLibrary/app"
	"github.com/lcmps/ExodiaLibrary/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const Selection = `id, name, name_pt, name_fr, "type", description, description_pt, description_fr, image,	"attribute", race, archetype, price, atk, def, "level"`

type Connection struct {
	DB     *gorm.DB
	Config *app.Config
}

func InitConnection() (Connection, error) {
	var conn Connection
	appData, err := app.InitConfig()
	if err != nil {
		return conn, err
	}
	conn.Config = appData
	connString := fmt.Sprintf(`host=%s user=%s password=%s dbname=%s sslmode=disable`,
		conn.Config.DB_Host,
		conn.Config.DB_User,
		conn.Config.DB_Pass,
		conn.Config.DB_Name)
	conn.DB, err = gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		return conn, err
	}
	return conn, nil
}

func (conn *Connection) CreateTables() {
	conn.DB.AutoMigrate(model.Cards{})
	conn.DB.AutoMigrate(model.Portfolios{})
	conn.DB.AutoMigrate(model.PortfolioCards{})
	conn.DB.AutoMigrate(model.Users{})
	conn.DB.AutoMigrate(model.UserPortfolios{})

	conn.createForeignKeys()
	conn.enableUUID()
}

func (conn *Connection) createForeignKeys() {

	cardsKey := conn.DB.Migrator().HasConstraint(model.PortfolioCards{}, "portfolio_cards_fk")
	portfKey := conn.DB.Migrator().HasConstraint(model.PortfolioCards{}, "portfolio_portfolios_fk")
	userfKey := conn.DB.Migrator().HasConstraint(model.UserPortfolios{}, "user_fk")
	userpfKey := conn.DB.Migrator().HasConstraint(model.UserPortfolios{}, "portfolios_fk")

	if !cardsKey {
		conn.DB.Exec(`
		ALTER TABLE
		public.portfolio_cards
	ADD CONSTRAINT 
		portfolio_cards_fk FOREIGN KEY (card) 
	REFERENCES
		public.cards(id) ON DELETE CASCADE;`)
	}

	if !portfKey {
		conn.DB.Exec(`
		ALTER TABLE
		public.portfolio_cards 
	ADD CONSTRAINT 
		portfolio_portfolios_fk FOREIGN KEY (portfolio) 
	REFERENCES 
	public.portfolios(id) ON DELETE CASCADE;`)
	}

	if !userfKey {
		conn.DB.Exec(`
		ALTER TABLE
		public.user_portfolios 
	ADD CONSTRAINT
		user_fk FOREIGN KEY ("user") 
	REFERENCES
		public.users(id);`)
	}

	if !userpfKey {
		conn.DB.Exec(`
		ALTER TABLE
		public.user_portfolios
	ADD CONSTRAINT
		portfolios_fk FOREIGN KEY (portfolio) 
	REFERENCES 
		public.portfolios(id);`)
	}

}

func (conn *Connection) enableUUID() {
	conn.DB.Exec(`
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	`)
}

func (conn *Connection) ImportCards() {

	en, fr, pt := app.GetAllCardsLanguages()

	for _, card := range en.Data {
		conn.DB.Exec(`
		INSERT INTO
			cards
		VALUES
		(?, ?, 'name_pt', 'name_fr', ?, ?, 'desc_pt', 'desc_fr', ?, ?, ?, ?, ?, ?, ?, ?)`, card.ID, card.Name,
			card.Type, card.Desc, card.CardImages[0].ID, card.Attribute, card.Race, card.Archetype,
			card.CardPrices[0].TcgplayerPrice, card.Atk, card.Def, card.Level)

		for _, img := range card.CardImages {
			conn.DB.Exec(`UPDATE cards SET image = array_append(image, ?) where id = ?`, img.ID, card.ID)
		}
	}

	for _, card := range pt.Data {
		conn.DB.Exec(`
		UPDATE cards SET name_pt = ?, description_pt = ? WHERE id = ? OR name = ?;`, card.Name, card.Desc,
			card.ID, card.NameEn)
	}

	for _, card := range fr.Data {
		conn.DB.Exec(`
		UPDATE cards SET name_fr = ?, description_fr = ? WHERE id = ? OR name = ?;`, card.Name, card.Desc,
			card.ID, card.NameEn)
	}
}
