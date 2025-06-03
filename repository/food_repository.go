package repository

import (
	"amur/models"
	"database/sql"
	"log"
)

type FoodRepository struct {
	db *sql.DB
}

func NewFoodRepository(db *sql.DB) *FoodRepository {
	return &FoodRepository{db: db}
}

func (r *FoodRepository) Create(food *models.Food) error {
	stmt, err := r.db.Prepare(`
        INSERT INTO foods (food_name, food_category, food_price, food_image)
        VALUES ($1, $2, $3, $4)
        RETURNING food_id
    `)
	if err != nil {
		log.Printf("Food Create prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(food.FoodName, food.FoodCategory, food.FoodPrice, food.FoodImage).Scan(&food.FoodID)
	if err != nil {
		log.Printf("Food Create exec xatolik: %v", err)
		return err
	}

	log.Printf("✅ Yangi ovqat qo'shildi: %s (ID: %d)", food.FoodName, food.FoodID)
	return nil
}

func (r *FoodRepository) GetAll() ([]*models.Food, error) {
	rows, err := r.db.Query(`
        SELECT food_id, food_name, food_category, food_price, food_image, created_at, updated_at
        FROM foods ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foods []*models.Food
	for rows.Next() {
		var food models.Food
		err := rows.Scan(&food.FoodID, &food.FoodName, &food.FoodCategory,
			&food.FoodPrice, &food.FoodImage, &food.CreatedAt, &food.UpdatedAt)
		if err != nil {
			log.Printf("Food GetAll scan xatolik: %v", err)
			continue
		}
		foods = append(foods, &food)
	}

	return foods, nil
}

func (r *FoodRepository) GetByID(id int) (*models.Food, error) {
	row := r.db.QueryRow(`
        SELECT food_id, food_name, food_category, food_price, food_image, created_at, updated_at
        FROM foods WHERE food_id = $1
    `, id)

	var food models.Food
	err := row.Scan(&food.FoodID, &food.FoodName, &food.FoodCategory,
		&food.FoodPrice, &food.FoodImage, &food.CreatedAt, &food.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &food, nil
}

func (r *FoodRepository) Update(id int, food *models.Food) error {
	stmt, err := r.db.Prepare(`
        UPDATE foods SET
            food_name = $1,
            food_category = $2,
            food_price = $3,
            food_image = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE food_id = $5
    `)
	if err != nil {
		log.Printf("Food Update prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(food.FoodName, food.FoodCategory, food.FoodPrice, food.FoodImage, id)
	if err != nil {
		log.Printf("Food Update exec xatolik: %v", err)
		return err
	}

	log.Printf("🔄 Ovqat yangilandi: %s (ID: %d)", food.FoodName, id)
	return nil
}

func (r *FoodRepository) Delete(id int) error {
	stmt, err := r.db.Prepare("DELETE FROM foods WHERE food_id = $1")
	if err != nil {
		log.Printf("Food Delete prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		log.Printf("Food Delete exec xatolik: %v", err)
		return err
	}

	log.Printf("🗑️ Ovqat o'chirildi (ID: %d)", id)
	return nil
}

func (r *FoodRepository) GetByCategory(category string) ([]*models.Food, error) {
	rows, err := r.db.Query(`
        SELECT food_id, food_name, food_category, food_price, food_image, created_at, updated_at
        FROM foods WHERE food_category = $1 ORDER BY created_at DESC
    `, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foods []*models.Food
	for rows.Next() {
		var food models.Food
		err := rows.Scan(&food.FoodID, &food.FoodName, &food.FoodCategory,
			&food.FoodPrice, &food.FoodImage, &food.CreatedAt, &food.UpdatedAt)
		if err != nil {
			log.Printf("Food GetByCategory scan xatolik: %v", err)
			continue
		}
		foods = append(foods, &food)
	}

	return foods, nil
}

func (r *FoodRepository) Count() int {
	row := r.db.QueryRow("SELECT COUNT(*) FROM foods")
	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Printf("Food Count xatolik: %v", err)
		return 0
	}
	return count
}
