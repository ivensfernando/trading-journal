module vsC1Y2025V01

go 1.23

toolchain go1.23.9

require (
	github.com/Kucoin/kucoin-universal-sdk v0.0.0
	github.com/Kucoin/kucoin-universal-sdk/sdk/golang v1.3.1
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/linstohu/nexapi v1.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.2
	golang.org/x/crypto v0.17.0
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.26.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator v9.31.0+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/leodido/go-urn v1.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Kucoin/kucoin-universal-sdk => ./internal/kucoinuniversal
