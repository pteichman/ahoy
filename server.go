package ahoy

import "database/sql"

type Server struct {
	db    *sql.DB
	clock Clock
}
