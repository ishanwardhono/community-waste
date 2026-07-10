package payment

const (
	insertQuery = `
		INSERT INTO payments (id, household_id, waste_id, amount, status, created_at, updated_at)
		VALUES (:id, :household_id, :waste_id, :amount, :status, :created_at, :updated_at)`

	baseSelect = `SELECT * FROM payments WHERE 1=1`
	baseCount  = `SELECT COUNT(*) FROM payments WHERE 1=1`
)
