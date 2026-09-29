package models

type CategorySpend struct {
	CategoryName     string  `json:"category_name"`
	TotalAmount      float64 `json:"total_amount"`
	AverageAmount    float64 `json:"average_amount"`
	TransactionCount int     `json:"transaction_count"`
}

type DailySummary struct {
	Date             string  `json:"date"`
	Income           float64 `json:"income"`
	Expense          float64 `json:"expense"`
	Net              float64 `json:"net"`
	TransactionCount int     `json:"transaction_count"`
}

type DailyAmount struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type Summary struct {
	TotalIncome             float64         `json:"total_income"`
	TotalExpense            float64         `json:"total_expense"`
	NetBalance              float64         `json:"net_balance"`
	TransactionCount        int             `json:"transaction_count"`
	IncomeTransactionCount  int             `json:"income_transaction_count"`
	ExpenseTransactionCount int             `json:"expense_transaction_count"`
	SavingsRate             float64         `json:"savings_rate"`
	AverageIncome           float64         `json:"average_income"`
	AverageExpense          float64         `json:"average_expense"`
	MedianIncome            float64         `json:"median_income"`
	MedianExpense           float64         `json:"median_expense"`
	HighestIncomeDay        *DailyAmount    `json:"highest_income_day"`
	LowestIncomeDay         *DailyAmount    `json:"lowest_income_day"`
	HighestExpenseDay       *DailyAmount    `json:"highest_expense_day"`
	LowestExpenseDay        *DailyAmount    `json:"lowest_expense_day"`
	DailyTotals             []DailySummary  `json:"daily_totals"`
	SpendingByCategory      []CategorySpend `json:"spending_by_category"`
	IncomeByCategory        []CategorySpend `json:"income_by_category"`
}
