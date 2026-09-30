package storage

import (
	"backend/models"
	"context"
	"time"
)

type dailyUpdate struct {
	models.DailySummary
	incomeCount  int
	expenseCount int
}

func (s *Storage) GetSummary(ctx context.Context, userID int, startDate, endDate time.Time) (models.Summary, error) {
	endExclusive := endDate.AddDate(0, 0, 1)
	summary := models.Summary{
		DailyTotals:        []models.DailySummary{},
		SpendingByCategory: []models.CategorySpend{},
		IncomeByCategory:   []models.CategorySpend{},
	}

	const summaryQuery = `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE kind = 'income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE kind = 'expense'), 0),
			COUNT(*) FILTER (WHERE kind = 'income'),
			COUNT(*) FILTER (WHERE kind = 'expense'),
			COALESCE(AVG(amount) FILTER (WHERE kind = 'income'), 0),
			COALESCE(AVG(amount) FILTER (WHERE kind = 'expense'), 0),
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) FILTER (WHERE kind = 'income'), 0),
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) FILTER (WHERE kind = 'expense'), 0)
		FROM transactions
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3`
	err := s.db.QueryRowContext(ctx, summaryQuery, userID, startDate, endExclusive).Scan(
		&summary.TotalIncome,
		&summary.TotalExpense,
		&summary.IncomeTransactionCount,
		&summary.ExpenseTransactionCount,
		&summary.AverageIncome,
		&summary.AverageExpense,
		&summary.MedianIncome,
		&summary.MedianExpense,
	)
	if err != nil {
		return models.Summary{}, err
	}
	summary.NetBalance = summary.TotalIncome - summary.TotalExpense
	summary.TransactionCount = summary.IncomeTransactionCount + summary.ExpenseTransactionCount
	if summary.TotalIncome > 0 {
		summary.SavingsRate = summary.NetBalance / summary.TotalIncome * 100
	}

	dailyUpdates, err := s.queryDailyUpdates(ctx, userID, startDate, endDate)
	if err != nil {
		return models.Summary{}, err
	}
	for _, daily := range dailyUpdates {
		summary.DailyTotals = append(summary.DailyTotals, daily.DailySummary)
		if daily.incomeCount > 0 {
			updateDailyExtremes(&summary.HighestIncomeDay, &summary.LowestIncomeDay, daily.Date, daily.Income)
		}
		if daily.expenseCount > 0 {
			updateDailyExtremes(&summary.HighestExpenseDay, &summary.LowestExpenseDay, daily.Date, daily.Expense)
		}
	}

	const categoryQuery = `
		SELECT t.kind, COALESCE(c.name, 'Uncategorized'), SUM(t.amount), AVG(t.amount), COUNT(*)
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id AND c.user_id = t.user_id
		WHERE t.user_id = $1 AND t.created_at >= $2 AND t.created_at < $3
		GROUP BY t.kind, c.name
		ORDER BY t.kind, COALESCE(c.name, 'Uncategorized')`
	categoryRows, err := s.db.QueryContext(ctx, categoryQuery, userID, startDate, endExclusive)
	if err != nil {
		return models.Summary{}, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var kind string
		var category models.CategorySpend
		if err := categoryRows.Scan(&kind, &category.CategoryName, &category.TotalAmount, &category.AverageAmount, &category.TransactionCount); err != nil {
			return models.Summary{}, err
		}
		switch kind {
		case "income":
			summary.IncomeByCategory = append(summary.IncomeByCategory, category)
		case "expense":
			summary.SpendingByCategory = append(summary.SpendingByCategory, category)
		}
	}
	if err := categoryRows.Err(); err != nil {
		return models.Summary{}, err
	}

	return summary, nil
}

func (s *Storage) GetDailyUpdates(ctx context.Context, userID int, startDate, endDate time.Time) ([]models.DailySummary, error) {
	updates, err := s.queryDailyUpdates(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	dailyTotals := make([]models.DailySummary, 0, len(updates))
	for _, update := range updates {
		dailyTotals = append(dailyTotals, update.DailySummary)
	}
	return dailyTotals, nil
}

func (s *Storage) queryDailyUpdates(ctx context.Context, userID int, startDate, endDate time.Time) ([]dailyUpdate, error) {
	endExclusive := endDate.AddDate(0, 0, 1)
	const dailyQuery = `
		WITH date_range AS (
			SELECT generate_series(
				($2 AT TIME ZONE 'Asia/Kolkata')::date,
				(($3 AT TIME ZONE 'Asia/Kolkata')::date - INTERVAL '1 day')::date,
				INTERVAL '1 day'
			)::date AS day
		), daily_transactions AS (
			SELECT
				(created_at AT TIME ZONE 'Asia/Kolkata')::date AS day,
				SUM(amount) FILTER (WHERE kind = 'income') AS income,
				SUM(amount) FILTER (WHERE kind = 'expense') AS expense,
				COUNT(*) AS transaction_count,
				COUNT(*) FILTER (WHERE kind = 'income') AS income_count,
				COUNT(*) FILTER (WHERE kind = 'expense') AS expense_count
			FROM transactions
			WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
			GROUP BY (created_at AT TIME ZONE 'Asia/Kolkata')::date
		)
		SELECT
			TO_CHAR(date_range.day, 'YYYY-MM-DD'),
			COALESCE(daily_transactions.income, 0),
			COALESCE(daily_transactions.expense, 0),
			COALESCE(daily_transactions.transaction_count, 0),
			COALESCE(daily_transactions.income_count, 0),
			COALESCE(daily_transactions.expense_count, 0)
		FROM date_range
		LEFT JOIN daily_transactions ON daily_transactions.day = date_range.day
		ORDER BY date_range.day`
	rows, err := s.db.QueryContext(ctx, dailyQuery, userID, startDate, endExclusive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updates := []dailyUpdate{}
	for rows.Next() {
		var update dailyUpdate
		if err := rows.Scan(
			&update.Date,
			&update.Income,
			&update.Expense,
			&update.TransactionCount,
			&update.incomeCount,
			&update.expenseCount,
		); err != nil {
			return nil, err
		}
		update.Net = update.Income - update.Expense
		updates = append(updates, update)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return updates, nil
}

func updateDailyExtremes(highest, lowest **models.DailyAmount, date string, amount float64) {
	if *highest == nil || amount > (*highest).Amount {
		*highest = &models.DailyAmount{Date: date, Amount: amount}
	}
	if *lowest == nil || amount < (*lowest).Amount {
		*lowest = &models.DailyAmount{Date: date, Amount: amount}
	}
}
