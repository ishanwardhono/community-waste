package pickup

const (
	insertQuery = `
		INSERT INTO waste_pickups (id, household_id, type, status, safety_check, created_at, updated_at)
		VALUES (:id, :household_id, :type, :status, :safety_check, :created_at, :updated_at)`

	baseSelect = `SELECT * FROM waste_pickups WHERE 1=1`
	baseCount  = `SELECT COUNT(*) FROM waste_pickups WHERE 1=1`

	getQuery = `SELECT * FROM waste_pickups WHERE id = $1`

	scheduleQuery = `
		UPDATE waste_pickups
		SET status = 'scheduled', pickup_date = $2, updated_at = now()
		WHERE id = $1 AND status = 'pending'
		RETURNING *`

	completeQuery = `
		UPDATE waste_pickups
		SET status = 'completed', updated_at = now()
		WHERE id = $1 AND status = 'scheduled'
		RETURNING *`

	cancelQuery = `
		UPDATE waste_pickups
		SET status = 'canceled', updated_at = now()
		WHERE id = $1 AND status IN ('pending', 'scheduled')
		RETURNING *`
)
