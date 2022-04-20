package db

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/lcmps/ExodiaLibrary/model"
)

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

func (conn *Connection) AddPortfolio(name, desc, cover string) model.Portfolios {
	var res model.Portfolios
	var pfl = model.Portfolios{
		ID:          uuid.New(),
		Name:        name,
		Description: desc,
		Cover:       cover,
	}

	conn.DB.Table("portfolios").Create(&pfl).Scan(&res)
	return res
}

func (conn *Connection) AddUser(name, email, picURL, locale, accType string) model.Users {
	var res model.Users
	var user = model.Users{
		ID:          uuid.New(),
		Name:        name,
		Email:       email,
		Picture:     picURL,
		Locale:      locale,
		AccountType: accType,
	}

	conn.DB.Table("users").Create(&user).Scan(&res)
	return res
}

func (conn *Connection) GetUserByID(id string) model.Users {
	var user model.Users

	conn.DB.Table("users").Where("id = ?", id).Find(&user)

	return user
}

func (conn *Connection) GetUserByEmail(email string) model.Users {
	var user model.Users

	conn.DB.Table("users").Where("email = ?", email).Find(&user)

	return user
}
