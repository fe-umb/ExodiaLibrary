package db

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
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

	conn.createForeignKeys()
	conn.enableUUID()
}

func (conn *Connection) createForeignKeys() {

	cardsKey := conn.DB.Migrator().HasConstraint(model.PortfolioCards{}, "portfolio_cards_fk")
	portfKey := conn.DB.Migrator().HasConstraint(model.PortfolioCards{}, "portfolio_portfolios_fk")

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

}

func (conn *Connection) AddPortfolio(name, desc, cover string) {
	var pfl = model.Portfolios{
		ID:          uuid.New(),
		Name:        name,
		Description: desc,
		Cover:       cover,
	}

	conn.DB.Table("portfolios").Create(&pfl)
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

func getQueryMap(mod model.CardQuery) map[string]interface{} {
	w := make(map[string]interface{})

	if mod.Name != "" {
		w["name"] = mod.Name
	}
	if mod.Ctype != "" {
		w["type"] = mod.Ctype
	}
	if mod.Attribute != "" {
		w["attribute"] = mod.Attribute
	}
	if mod.Archetype != "" {
		w["archetype"] = mod.Archetype
	}
	if mod.Race != "" {
		w["race"] = mod.Race
	}
	if mod.Level != 0 {
		w["level"] = mod.Level
	}
	if mod.Atk != 0 {
		w["atk"] = mod.Atk
	}
	if mod.Def != 0 {
		w["def"] = mod.Def
	}
	if mod.Limit != 0 {
		w["limit"] = mod.Limit
	} else {
		w["limit"] = 10
	}
	if mod.Offset != 0 {
		w["offset"] = mod.Offset
	} else {
		w["offset"] = 0
	}

	return w
}

func (conn *Connection) GetCardsByFilter(mod model.CardQuery) model.CardResponse {
	var res []model.Cards
	var queryCount int64
	w := getQueryMap(mod)

	tx := conn.DB.Select(Selection).Table(`cards`)

	for k, v := range w {

		if k == "limit" {
			tx = tx.Limit(v.(int))
			continue
		}
		if k == "offset" {
			tx = tx.Offset(v.(int))
			continue
		}

		if reflect.TypeOf(v) == reflect.TypeOf("") {

			if k == "name" {
				tx.Where("name_pt LIKE ? OR name_fr LIKE ? OR name LIKE ?", "%"+v.(string)+"%", "%"+v.(string)+"%", "%"+v.(string)+"%")
			} else {
				tx.Where("lower("+k+") LIKE lower(?)", "%"+v.(string)+"%")
			}
		} else {
			tx.Where(k+" = ?", v)
		}

	}

	tx.Count(&queryCount)
	tx.Find(&res)

	return model.CardResponse{
		Total: queryCount,
		Cards: res,
	}
}

func (conn *Connection) GetRandomCards(lim int) []model.Cards {
	var res []model.Cards

	if lim == 0 {
		lim = 1
	}

	conn.DB.Raw(`select * from cards tablesample bernoulli(1) where name_pt != 'name_pt' and name_fr != 'name_fr' order by random() limit ?;`, lim).Find(&res)
	return res
}
