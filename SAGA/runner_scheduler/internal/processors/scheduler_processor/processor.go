package scheduler_processor

type Processor struct {
	repo Repository
}

func NewProcessor(repo Repository) *Processor {
	return &Processor{
		repo: repo,
	}
}
