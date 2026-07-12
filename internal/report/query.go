package report

const (
	wasteSummaryQuery = `
		SELECT type, status, COUNT(*) AS count
		FROM waste_pickups
		GROUP BY type, status
		ORDER BY type, status`

	paymentSummaryQuery = `
		SELECT status, COUNT(*) AS count, COALESCE(SUM(amount), 0) AS total_amount
		FROM payments
		GROUP BY status
		ORDER BY status`

	householdInfoQuery = `SELECT id, owner_name, address FROM households WHERE id = $1`

	pickupsByHouseholdQuery = `
		SELECT id, type, status, pickup_date, created_at
		FROM waste_pickups WHERE household_id = $1
		ORDER BY created_at DESC`

	paymentsByHouseholdQuery = `
		SELECT id, waste_id, amount, status, payment_date, created_at
		FROM payments WHERE household_id = $1
		ORDER BY created_at DESC`
)
