package postgres

import (
	"database/sql"
	"errors"

	"github.com/SmoothWay/metrics/internal/model"
)

var ErrNotFound = errors.New("value not found")

type PostgreDB struct {
	db *sql.DB
}

func New(db *sql.DB) (*PostgreDB, error) {
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS metrics (
		name TEXT PRIMARY KEY,
		type VARCHAR(50),
		value DOUBLE PRECISION,
		delta INTEGER);`)
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return &PostgreDB{
		db: db,
	}, nil
}

func (p *PostgreDB) SetCounterMetric(key string, value int64) error {
	var id string
	var prevDelta sql.NullInt64
	stmtGetDelta := `SELECT name, delta FROM metrics WHERE name = $1 and type = 'counter'`
	stmtUpdateDelta := `UPDATE metrics SET delta = $1 WHERE name = $2`
	stmtInsertDelta := `INSERT INTO metrics(name, type, delta) VALUES($1, $2, $3)`

	getDelta := p.db.QueryRow(stmtGetDelta, key)
	err := getDelta.Scan(&id, &prevDelta)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = p.db.Exec(stmtInsertDelta, key, model.MetricTypeCounter, value)
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	_, err = p.db.Exec(stmtUpdateDelta, prevDelta.Int64+value, id)
	return err
}

func (p *PostgreDB) SetGaugeMetric(key string, value float64) error {
	var id string
	stmtGetDelta := `SELECT name FROM metrics WHERE name = $1 and type = 'gauge'`
	stmtUpdateDelta := `UPDATE metrics SET value = $1 WHERE id = $2`
	stmtInsertDelta := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET value = $3 WHERE name = $1`

	getDelta := p.db.QueryRow(stmtGetDelta, key)
	err := getDelta.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = p.db.Exec(stmtInsertDelta, key, model.MetricTypeGauge, value)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	_, err = p.db.Exec(stmtUpdateDelta, value, id)
	return err
}

func (p *PostgreDB) GetCounterMetric(key string) (int64, error) {
	stmtSelect := `SELECT delta FROM metrics WHERE name = $1 AND type = 'counter'`
	var counter sql.NullInt64

	row := p.db.QueryRow(stmtSelect, key)

	err := row.Scan(&counter)
	if err != nil {
		return 0, err
	}
	if !counter.Valid {
		return 0, sql.ErrNoRows
	}
	return counter.Int64, nil
}

func (p *PostgreDB) GetGaugeMetric(key string) (float64, error) {
	stmtSelect := `SELECT value FROM metrics WHERE name = $1 AND type = 'gauge'`
	var value sql.NullFloat64

	row := p.db.QueryRow(stmtSelect, key)

	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	if !value.Valid {
		return 0, sql.ErrNoRows
	}
	return value.Float64, nil
}

func (p *PostgreDB) GetAllMetric() []model.Metrics {
	selectStmt := `SELECT name, type, delta, value FROM metrics`

	rows, err := p.db.Query(selectStmt)
	if err != nil {
		return nil
	}

	var metrics []model.Metrics

	for rows.Next() {
		var metric model.Metrics

		err = rows.Scan(metric.ID, metric.Mtype, metric.Delta, metric.Value)
		if err != nil {
			return nil
		}

		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		return nil
	}

	return metrics
}
