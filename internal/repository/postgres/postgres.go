package postgres

import (
	"database/sql"
	"errors"

	"github.com/SmoothWay/metrics/internal/model"
)

type PostgreDB struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgreDB {
	return &PostgreDB{
		db: db,
	}
}

func (p *PostgreDB) SetCounterMetric(key string, value int64) error {
	var id string
	var prevDelta sql.NullInt64
	stmtGetDelta := `SELECT id, delta FROM metrics WHERE name = $1 and type = 'counter'`
	stmtUpdateDelta := `UPDATE metrics SET delta = $1 WHERE id = $2`
	stmtInsertDelta := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3)`

	getDelta := p.db.QueryRow(stmtGetDelta, key)
	err := getDelta.Scan(&id, &prevDelta)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = p.db.Exec(stmtInsertDelta, key, model.MetricTypeCounter, value)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	_, err = p.db.Exec(stmtUpdateDelta, prevDelta.Int64+value, id)
	return err
}

func (p *PostgreDB) SetGaugeMetric(key string, value float64) error {
	var id string
	stmtGetDelta := `SELECT id FROM metrics WHERE name = $1 and type = 'gauge'`
	stmtUpdateDelta := `UPDATE metrics SET value = $1 WHERE id = $2`
	stmtInsertDelta := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3)`

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
	stmtSelect := `SELECT delta FROM metrics WHERE key = $1, type = 'counter'`
	var counter int64

	row := p.db.QueryRow(stmtSelect, key)

	err := row.Scan(&counter)
	if err != nil {
		return 0, err
	}

	return counter, nil
}

func (p *PostgreDB) GetGaugeMetric(key string) (float64, error) {
	stmtSelect := `SELECT value FROM metrics WHERE key = $1, type = 'gauge'`
	var value float64

	row := p.db.QueryRow(stmtSelect, key)

	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
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
