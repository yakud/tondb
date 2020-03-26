package timeseries

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/cache"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	createTsVolumeByGrams = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_ts_VolumeByGrams
	ENGINE = SummingMergeTree() 
	PARTITION BY tuple()
	ORDER BY (Time, WorkchainId)
	POPULATE 
	AS
	SELECT
		toStartOfInterval(Time, INTERVAL 10 MINUTE) as Time,
		WorkchainId,
	    sum(Messages.ValueNanograms) as VolumeNanograms
	FROM transactions
	ARRAY JOIN Messages
	GROUP BY Time, WorkchainId
`
	dropTsVolumeByGrams = `DROP TABLE _view_ts_VolumeByGrams`

	selectVolumeByGrams = `
	SELECT 
       WorkchainId,
	   groupArray(Time),
	   groupArray(toString(VolumeGrams))
    FROM (
		SELECT 
			toUInt64(Time) as Time, 
		    WorkchainId,
		    toDecimal128(sum(VolumeNanograms), 10) * toDecimal128(0.000000001, 10) as VolumeGrams
		FROM _view_ts_VolumeByGrams 
		WHERE Time <= now() AND Time >= now()-?
		GROUP BY Time, WorkchainId
		ORDER BY Time, WorkchainId
	) GROUP BY WorkchainId
`
)

type VolumeByGrams struct {
	conn        *sql.DB
	resultCache *cache.WithTimer
}

func (t *VolumeByGrams) CreateTable() error {
	_, err := t.conn.Exec(createTsVolumeByGrams)
	return err
}

func (t *VolumeByGrams) DropTable() error {
	_, err := t.conn.Exec(dropTsVolumeByGrams)
	return err
}

func (t *VolumeByGrams) GetVolumeByGrams() (*tonapi.VolumeByGramsResult, error) {
	if res, ok := t.resultCache.Get(); ok {
		switch res.(type) {
		case *tonapi.VolumeByGramsResult:
			return res.(*tonapi.VolumeByGramsResult), nil
		}
	}

	rows, err := t.conn.Query(selectVolumeByGrams, []byte("INTERVAL 2 DAY"))
	if err != nil {
		return nil, err
	}

	var resp = &tonapi.VolumeByGramsResult{
		Rows: make([]tonapi.VolumeByGramsTimeseries, 0),
	}

	for rows.Next() {
		times := make([]uint64, 0)

		row := tonapi.VolumeByGramsTimeseries{
			Time:        make([]tonapi.Uint64, 0),
			VolumeGrams: make([]string, 0),
		}
		if err := rows.Scan(
			&row.WorkchainId,
			&times,
			&row.VolumeGrams,
		); err != nil {
			rows.Close()
			return nil, err
		}

		for _, v := range times {
			row.Time = append(row.Time, tonapi.Uint64(v))
		}
		for i, v := range row.VolumeGrams {
			row.VolumeGrams[i] = utils.TruncateRightZeros(v)
		}

		resp.Rows = append(resp.Rows, row)
	}

	rows.Close()

	t.resultCache.Set(resp)

	return resp, nil
}

func NewVolumeByGrams(conn *sql.DB) *VolumeByGrams {
	return &VolumeByGrams{
		conn:        conn,
		resultCache: cache.NewWithTimer(time.Second),
	}
}
