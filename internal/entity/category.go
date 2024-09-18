package entity

// type Category struct {
// 	ID            string        `bson:"id" json:"id"`
// 	Name          string        `bson:"name" json:"name"`
// 	Description   string        `bson:"description" json:"description"`
// 	SubCategories []SubCategory `bson:"sub_categories" json:"sub_categories"`
// }

// type SubCategory struct {
// 	ID   string `bson:"id" json:"id"`
// 	Name string `bson:"name" json:"name"`
// }

type Category struct {
	ID     string `bson:"id" json:"id"`
	NameUz string `bson:"name_uz" json:"name_uz"`
	NameRu string `bson:"name_ru" json:"name_ru"`
}

type SubCategory struct {
	ID         string `bson:"id" json:"id"`
	CategoryID string `bson:"category_id" json:"category_id"`
	NameUz     string `bson:"name_uz" json:"name_uz"`
	NameRu     string `bson:"name_ru" json:"name_ru"`
	Name       string `json:"name"`
}

type CategoryWithSubCategories struct {
	ID            string        `bson:"id" json:"id"`
	NameUz        string        `bson:"name_uz" json:"name_uz"`
	NameRu        string        `bson:"name_ru" json:"name_ru"`
	Name          string        `json:"name"`
	SubCategories []SubCategory `bson:"subcategories" json:"subcategories"`
}
