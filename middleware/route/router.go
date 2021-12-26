package route

import (
	"darvik80/go-network/database"
	"database/sql"
)

type SortingResult struct {
	Destination int
	MachineCode int
}

type SortingTask struct {
	Barcodes []string
}

type SortingRouter interface {
	Advice(task SortingTask) []SortingResult
}

type portCodeSortingRouter struct {
	repository database.PortCodeMappingRepository
}

func NewPortCodeSortingRouter(db *sql.DB) *portCodeSortingRouter {
	return &portCodeSortingRouter{
		repository: database.NewPortCodeMappingRepository(db),
	}
}

func (r *portCodeSortingRouter) Advice(task SortingTask) []SortingResult {
	var result []SortingResult
	for _, b := range task.Barcodes {
		if rows, err := r.repository.FindByPortCode(b); err == nil {
			for _, row := range rows {
				result = append(result, SortingResult{
					Destination: row.Destination,
					MachineCode: 0,
				})
			}
		}
	}

	if len(result) == 0 {
		result = append(result, SortingResult{
			Destination: 0,
			MachineCode: 1,
		})
	}

	return result
}
