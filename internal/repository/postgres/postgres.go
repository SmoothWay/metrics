package postgres

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/SmoothWay/metrics/internal/logger"
	"github.com/SmoothWay/metrics/internal/model"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("value not found")

type PostgreDB struct {
	db *sql.DB
}

func New(dsn string) (*PostgreDB, error) {
	var counts int
	var connection *sql.DB
	var err error
	for {
		connection, err = openDB(dsn)
		if err != nil {
			log.Println("Database not ready...")
			counts++
		} else {
			log.Println("Connected to database")
			break
		}
		if counts > 2 {
			return nil, err
		}
		log.Printf("Retrying to connect after %d seconds\n", counts+2)
		time.Sleep(time.Duration(2+counts) * time.Second)
	}

	stmtCreateTable, err := connection.Prepare(`
	CREATE TABLE IF NOT EXISTS metrics (
		name TEXT PRIMARY KEY,
		type VARCHAR(50),
		value DOUBLE PRECISION,
		delta BIGINT);`)
	if err != nil {
		return nil, err
	}

	_, err = stmtCreateTable.Exec()
	if err != nil {
		return nil, err
	}

	return &PostgreDB{
		db: connection,
	}, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (p *PostgreDB) SetCounterMetric(key string, value int64) error {
	var name string
	var prevDelta sql.NullInt64
	stmtGetCounter := `SELECT name, delta FROM metrics WHERE name = $1 and type = 'counter'`
	stmtUpdateCounter := `UPDATE metrics SET delta = $1 WHERE name = $2`
	stmtInsertCounter := `INSERT INTO metrics(name, type, delta) VALUES($1, $2, $3)`

	tx, err := p.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	getDelta := tx.QueryRow(stmtGetCounter, key)

	err = getDelta.Scan(&name, &prevDelta)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.Exec(stmtInsertCounter, key, model.MetricTypeCounter, value)
			if err != nil {
				tx.Rollback()
				return err
			}
			tx.Commit()
			return nil
		} else {
			tx.Rollback()
			return err
		}
	}
	updateValue := prevDelta.Int64 + value
	log.Println(updateValue)
	_, err = tx.Exec(stmtUpdateCounter, updateValue, name)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (p *PostgreDB) SetGaugeMetric(key string, value float64) error {
	var name string
	stmtGetGauge := `SELECT name FROM metrics WHERE name = $1 and type = 'gauge'`
	stmtUpdateGauge := `UPDATE metrics SET value = $1 WHERE name = $2`
	stmtInsertGauge := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET value = $3 WHERE metrics.name = $1`

	tx, err := p.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}
	getDelta := p.db.QueryRow(stmtGetGauge, key)
	err = getDelta.Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = p.db.Exec(stmtInsertGauge, key, model.MetricTypeGauge, value)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	} else if err != nil {
		tx.Rollback()
		return err
	}

	_, err = p.db.Exec(stmtUpdateGauge, value, name)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (p *PostgreDB) SetAllMetrics(metrics []model.Metrics) error {
	stmtGetCounter := `SELECT delta FROM metrics WHERE name = $1 and type = 'counter'`

	upsertGaugeStmt := `INSERT INTO metrics(name, type, value) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET value = $3 WHERE metrics.name = $1 `

	upsertCounterStmt := `INSERT INTO metrics(name, type, delta) VALUES($1, $2, $3) 
	ON CONFLICT (name) DO UPDATE SET delta = $3 WHERE metrics.name = $1`

	tx, err := p.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, v := range metrics {
		v := v
		if v.Mtype == model.MetricTypeCounter {
			row := tx.QueryRow(stmtGetCounter, v.ID)
			var delta sql.NullInt64

			err = row.Scan(&delta)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
				return err
			}
			_, err = tx.Exec(upsertCounterStmt, v.ID, v.Mtype, *v.Delta+delta.Int64)
			if err != nil {
				logger.Log().Info("counter error tx", zap.Error(err))
				tx.Rollback()
				return err
			}

		} else if v.Mtype == model.MetricTypeGauge {
			_, err = tx.Exec(upsertGaugeStmt, v.ID, v.Mtype, v.Value)
			if err != nil {
				logger.Log().Info("gauge error tx", zap.Error(err))
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
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
	tx, err := p.db.Begin()
	if err != nil {
		tx.Rollback()
		return nil
	}
	rows, err := p.db.Query(selectStmt)
	if err != nil {
		tx.Rollback()
		return nil
	}

	var metrics []model.Metrics

	for rows.Next() {
		var metric model.Metrics

		err = rows.Scan(metric.ID, metric.Mtype, metric.Delta, metric.Value)
		if err != nil {
			tx.Rollback()
			return nil
		}

		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		tx.Rollback()
		return nil
	}
	tx.Commit()
	return metrics
}

func (p *PostgreDB) PingStorage() error {
	err := p.db.Ping()
	return err
}
