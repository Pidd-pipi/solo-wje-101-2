package testsupport

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync/atomic"

	"gorm.io/gorm"
)

// ErrInjected is the sentinel returned when a fault switch is ON.
var ErrInjected = errors.New("testsupport: injected storage read failure (recoverable)")

// Profile statistic fault kinds. Each targets one Profile read so tests can
// prove the whole profile fails (rather than zeroing that statistic) per source.
const (
	ProfileFaultNone    int32 = 0
	ProfileFaultNotes   int32 = 1 // ListByUser: tasting note history
	ProfileFaultAvg     int32 = 2 // AvgScore: average overall score
	ProfileFaultOrigins int32 = 3 // TopOrigins: favorite origins grouping
	ProfileFaultFollow  int32 = 4 // follower/following counts
	ProfileFaultLikes   int32 = 5 // likes received
)

// Fault is a runtime, concurrently-safe, recoverable switchboard that intercepts
// executed SQL at the connection-pool layer. Intercepting there (instead of a
// GORM query callback) ensures .Scan()/.Find()/.Count() paths are ALL covered,
// including GROUP BY aggregations that bypass query callbacks.
type Fault struct {
	// failFavoriteState targets the per-user favorite-id read used to stamp
	// is_favored on the logged-in bean list (IDSetByUser: SELECT bean_id FROM
	// bean_favorites WHERE user_id=?, without a JOIN).
	failFavoriteState atomic.Bool

	// failProfilePreference targets the roast-level grouping, the first
	// aggregation of FavoriteService.Preference: ... GROUP BY roast_level.
	failProfilePreference atomic.Bool

	// profileStat selects one profile-statistic read to fail (ProfileFault*).
	profileStat atomic.Int32
}

// NewFault returns an initially-off switchboard.
func NewFault() *Fault { return &Fault{} }

// Install swaps BOTH the root pool and the statement pool for a faulting
// wrapper. New per-request sessions inherit db.Statement.ConnPool (GORM
// getInstance), so replacing only db.ConnPool has no effect on real queries.
// Transactions still begin on the underlying pool (faults target read queries,
// not the admin delete transaction).
func (f *Fault) Install(db *gorm.DB) {
	wrapped := &faultPool{Fault: f, next: db.ConnPool}
	db.ConnPool = wrapped
	if db.Statement != nil {
		db.Statement.ConnPool = wrapped
	}
}

// SetFavoriteStateFault flips the favorite-state read fault on/off at runtime.
func (f *Fault) SetFavoriteStateFault(on bool) { f.failFavoriteState.Store(on) }

// SetProfilePreferenceFault flips the profile preference read fault on/off.
func (f *Fault) SetProfilePreferenceFault(on bool) { f.failProfilePreference.Store(on) }

// SetProfileStatFault makes one profile-statistic read fail (see ProfileFault*).
// Pass ProfileFaultNone to recover.
func (f *Fault) SetProfileStatFault(kind int32) { f.profileStat.Store(kind) }

// shouldFail inspects the final SQL for an active fault signature. Identifier
// quotes (backticks from SQLite, double quotes elsewhere) are stripped so the
// match is dialect-independent.
func (f *Fault) shouldFail(query string) bool {
	q := strings.NewReplacer("`", "", "\"", "").Replace(strings.ToLower(query))
	if f.failFavoriteState.Load() {
		// Plain bean_favorites read selecting bean_id for one user with no join
		// (the IDSetByUser pluck). Excludes joined recent-favorites / process
		// aggregation queries and COUNT(*) totals.
		if strings.Contains(q, "from bean_favorites") &&
			strings.Contains(q, "bean_id") &&
			!strings.Contains(q, "join") &&
			!strings.Contains(q, "count(") {
			return true
		}
	}
	if f.failProfilePreference.Load() {
		// RoastCountsByUser: ... FROM tasting_notes ... GROUP BY roast_level.
		if strings.Contains(q, "from tasting_notes") &&
			strings.Contains(q, "group by") &&
			strings.Contains(q, "roast_level") {
			return true
		}
	}
	switch f.profileStat.Load() {
	case ProfileFaultNotes:
		// Tasting history: full-row note read for one user (Find, not an
		// aggregate/pluck which select explicit columns).
		if strings.HasPrefix(strings.TrimSpace(q), "select * from tasting_notes") &&
			strings.Contains(q, "where user_id") {
			return true
		}
	case ProfileFaultAvg:
		if strings.Contains(q, "avg(overall_score)") &&
			strings.Contains(q, "from tasting_notes") {
			return true
		}
	case ProfileFaultOrigins:
		// TopOrigins plucks origin with GROUP BY origin.
		if strings.Contains(q, "select origin from tasting_notes") &&
			strings.Contains(q, "group by") &&
			strings.Contains(q, "origin") {
			return true
		}
	case ProfileFaultFollow:
		// FollowService.Counts: count(*) over user_follows (either direction).
		if strings.Contains(q, "from user_follows") &&
			strings.Contains(q, "count(") {
			return true
		}
	case ProfileFaultLikes:
		// LikeRepository.CountByUserNotes joins likes -> tasting_notes.
		if strings.Contains(q, "from likes") &&
			strings.Contains(q, "join tasting_notes") {
			return true
		}
	}
	return false
}

// faultPool wraps a gorm.ConnPool and fails matching read queries.
type faultPool struct {
	Fault *Fault
	next  gorm.ConnPool
}

func (p *faultPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.next.PrepareContext(ctx, query)
}

func (p *faultPool) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.next.ExecContext(ctx, query, args...)
}

func (p *faultPool) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if p.Fault.shouldFail(query) {
		return nil, ErrInjected
	}
	return p.next.QueryContext(ctx, query, args...)
}

func (p *faultPool) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return p.next.QueryRowContext(ctx, query, args...)
}

// BeginTx keeps gorm transactions working by delegating to the underlying pool.
// The returned *sql.Tx bypasses read faults (the injected outages target
// list/profile reads, not the admin de-listing write transaction).
func (p *faultPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if b, ok := p.next.(gorm.TxBeginner); ok {
		return b.BeginTx(ctx, opts)
	}
	return nil, errors.New("testsupport: underlying pool does not support transactions")
}
