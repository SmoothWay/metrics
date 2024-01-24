package postgres

import (
	"database/sql"
	"errors"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"go.uber.org/zap"
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
	stmtGetCounter := `SELECT delta FROM metrics WHERE name = $1 and type = 'counter'`
	stmtUpdateCounter := `UPDATE metrics SET delta = $1 WHERE name = $2`
	stmtInsertCounter := `INSERT INTO metrics(name, type, delta) VALUES($1, $2, $3)`

	getDelta := p.db.QueryRow(stmtGetCounter, key)
	err := getDelta.Scan(&id, &prevDelta)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = p.db.Exec(stmtInsertCounter, key, model.MetricTypeCounter, value)
			if err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}

	_, err = p.db.Exec(stmtUpdateCounter, prevDelta.Int64+value, id)
	return err
}

func (p *PostgreDB) SetGaugeMetric(key string, value float64) error {
	var name string
	stmtGetGauge := `SELECT name FROM metrics WHERE name = $1 and type = 'gauge'`
	stmtUpdateGauge := `UPDATE metrics SET value = $1 WHERE name = $2`
	stmtInsertGauge := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET value = $3 WHERE metrics.name = $1`

	getDelta := p.db.QueryRow(stmtGetGauge, key)
	err := getDelta.Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = p.db.Exec(stmtInsertGauge, key, model.MetricTypeGauge, value)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	_, err = p.db.Exec(stmtUpdateGauge, value, name)
	return err
}

func (p *PostgreDB) SetAllMetrics(metrics []model.Metrics) error {
	stmtGetCounter := `SELECT delta FROM metrics WHERE name = $1 and type = 'counter'`

	upsertGaugeStmt := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET value = $3 WHERE metrics.name = $1 `

	upsertCounterStmt := `INSERT INTO metrics(name, type, delta) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET delta = $3 WHERE metrics.name = $1`

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	for _, v := range metrics {
		if v.Mtype == model.MetricTypeCounter {
			row := tx.QueryRow(stmtGetCounter, v.ID)
			var delta sql.NullInt64

			err = row.Scan(&delta)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			_, err = tx.Exec(upsertCounterStmt, v.ID, v.Mtype, *v.Delta+delta.Int64)
			if err != nil {
				logger.Log.Info("counter error tx", zap.Error(err))
				err = tx.Rollback()
				if err != nil {
					return err
				}
				return err
			}

		} else if v.Mtype == model.MetricTypeGauge {
			_, err = tx.Exec(upsertGaugeStmt, v.ID, v.Mtype, v.Value)
			if err != nil {
				logger.Log.Info("gauge error tx", zap.Error(err))
				err = tx.Rollback()
				if err != nil {
					return err
				}
				return err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
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
