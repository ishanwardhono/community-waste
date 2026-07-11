package payment

const (
	insertQuery = `
		INSERT INTO payments (id, household_id, waste_id, amount, status, created_at, updated_at)
		VALUES (:id, :household_id, :waste_id, :amount, :status, :created_at, :updated_at)`

	hasPendingQuery = `SELECT EXISTS (SELECT 1 FROM payments WHERE household_id = $1 AND status = 'pending')`

	baseSelect = `SELECT * FROM payments WHERE 1=1`
	baseCount  = `SELECT COUNT(*) FROM payments WHERE 1=1`
)
