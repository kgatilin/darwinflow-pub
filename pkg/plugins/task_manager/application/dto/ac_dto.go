package dto

// CreateACDTO represents input for creating acceptance criteria
type CreateACDTO struct {
	TaskID              string
	Description         string
	TestingInstructions string
}

// UpdateACDTO represents input for updating acceptance criteria
type UpdateACDTO struct {
	ID                  string
	Description         *string
	TestingInstructions *string
}

// VerifyACDTO represents input for verifying acceptance criteria
type VerifyACDTO struct {
	ID         string
	VerifiedBy string
	VerifiedAt string
}

// FailACDTO represents input for marking acceptance criteria as failed
type FailACDTO struct {
	ID       string
	Feedback string
}

// SkipACDTO represents input for marking acceptance criteria as skipped
type SkipACDTO struct {
	ID     string
	Reason string
}

// ACFilters represents filters for listing acceptance criteria
type ACFilters struct {
	TaskID       *string
	IterationNum *int
	Status       []string
}
