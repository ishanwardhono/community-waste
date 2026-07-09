package pickup

const (
	insertQuery = `
		INSERT INTO waste_pickups (id, household_id, type, status, safety_check, created_at, updated_at)
		VALUES (:id, :household_id, :type, :status, :safety_check, :created_at, :updated_at)`

	baseSelect = `SELECT * FROM waste_pickups WHERE 1=1`
	baseCount  = `SELECT COUNT(*) FROM waste_pickups WHERE 1=1`
)
