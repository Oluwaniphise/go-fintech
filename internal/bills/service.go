package bills

import (
	"net/http"

	"gorm.io/gorm"
)

type Dependencies struct {
	DB         *gorm.DB
	HTTPClient *http.Client
}
