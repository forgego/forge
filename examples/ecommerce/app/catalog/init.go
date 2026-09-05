package catalog

import "github.com/forgego/forge/db"

// Init initializes the catalog package with database connection
func Init(database *db.DB) {
	if CategoryObjects != nil {
		CategoryObjects.SetDB(database)
	}
	if BrandObjects != nil {
		BrandObjects.SetDB(database)
	}
	if ProductObjects != nil {
		ProductObjects.SetDB(database)
	}
	if ProductVariantObjects != nil {
		ProductVariantObjects.SetDB(database)
	}
	if ProductImageObjects != nil {
		ProductImageObjects.SetDB(database)
	}
	if ProductAttributeObjects != nil {
		ProductAttributeObjects.SetDB(database)
	}
	if ProductAttributeValueObjects != nil {
		ProductAttributeValueObjects.SetDB(database)
	}
}
