package database

import (
	"database/sql"
	"fmt"
)

type PortCodeMapping struct {
	Id          int64
	DeviceId    int64
	PortCode    string
	Destination int
}

type PortCodeMappingRepository interface {
	FindByPortCode(portCode string) ([]PortCodeMapping, error)
}

type repository struct {
	db *sql.DB
}

func NewPortCodeMappingRepository(db *sql.DB) PortCodeMappingRepository {
	return repository{db}
}

func (r repository) FindByPortCode(portCode string) ([]PortCodeMapping, error) {
	var res []PortCodeMapping

	rows, err := r.db.Query("SELECT id, device_id, port_code, destination FROM port_code_mapping WHERE port_code = $1", portCode)
	if err != nil {
		return nil, fmt.Errorf("port_code_mapping %q: %v", portCode, err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var rec PortCodeMapping
		if err := rows.Scan(&rec.Id, &rec.DeviceId, &rec.PortCode, &rec.Destination); err != nil {
			return nil, fmt.Errorf("port_code_mapping %q: %v", portCode, err)
		}
		res = append(res, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("port_code_mapping %q: %v", portCode, err)
	}

	return res, nil
}
