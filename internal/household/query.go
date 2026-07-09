package household

const (
	insertQuery = `
		INSERT INTO households (id, owner_name, address, created_at, updated_at)
		VALUES (:id, :owner_name, :address, :created_at, :updated_at)`

	listQuery   = `SELECT * FROM households ORDER BY id DESC LIMIT $1 OFFSET $2`
	countQuery  = `SELECT COUNT(*) FROM households`
	getQuery    = `SELECT * FROM households WHERE id = $1`
	deleteQuery = `DELETE FROM households WHERE id = $1`
)
