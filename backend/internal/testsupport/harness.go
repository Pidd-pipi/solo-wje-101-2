package testsupport

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/config"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/router"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/util"
)

// Envelope mirrors the unified API response for assertions.
type Envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Harness is a repeatable, on-disk integration environment: a pooled, WAL SQLite
// database (never :memory:, never single-connection serialized), the real router,
// and an httptest server. Faults are injected through a runtime, recoverable plugin.
type Harness struct {
	t      *testing.T
	dir    string
	dbPath string

	DB     *gorm.DB
	sqlDB  *sql.DB
	Cfg    *config.Config
	Fault  *Fault
	Router *gin.Engine
	Server *httptest.Server

	Users []*model.User // 0: admin, 1: barista, 2: roaster
	Beans []model.CoffeeBean
}

func openPooledDB(t *testing.T, path string) (*gorm.DB, *sql.DB) {
	t.Helper()
	// Real concurrent pool (multiple open connections). WAL + busy_timeout allow
	// concurrent readers/writers against the on-disk file without serializing.
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open db %s: %v", path, err)
	}
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		t.Fatalf("set wal: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	sqlDB.SetConnMaxLifetime(0)
	return db, sqlDB
}

func migrateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(
		&model.User{}, &model.TastingNote{}, &model.BrewRecipe{}, &model.CoffeeBean{},
		&model.Comment{}, &model.Like{}, &model.UserFollow{}, &model.BeanFavorite{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
}

func seedUsers(t *testing.T, db *gorm.DB) []*model.User {
	t.Helper()
	mk := func(username, role string) *model.User {
		h, _ := bcrypt.GenerateFromPassword([]byte("pass1234"), bcrypt.DefaultCost)
		return &model.User{Username: username, Email: username + "@coffeetaste.local", PasswordHash: string(h), Bio: "tester", Role: role}
	}
	users := []*model.User{
		mk("admin_t", "admin"), mk("barista_t", "user"), mk("roaster_t", "user"),
	}
	for _, u := range users {
		if err := db.Create(u).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	return users
}

func seedBeans(t *testing.T, db *gorm.DB) []model.CoffeeBean {
	t.Helper()
	beans := []model.CoffeeBean{
		{Name: "耶加雪菲", Origin: "埃塞俄比亚", ProcessMethod: "washed", FlavorTags: `["柑橘","茉莉"]`, Description: "明亮柑橘酸质"},
		{Name: "哥伦比亚慧兰", Origin: "哥伦比亚", ProcessMethod: "washed", FlavorTags: `["坚果","焦糖"]`, Description: "甜感平衡"},
		{Name: "哥斯达黎加蜜处理", Origin: "哥斯达黎加", ProcessMethod: "honey", FlavorTags: `["莓果","红糖"]`, Description: "醇厚甜感"},
		{Name: "印尼曼特宁", Origin: "印度尼西亚", ProcessMethod: "natural", FlavorTags: `["草本","黑巧克力"]`, Description: "醇厚浓郁"},
		// Literal %, _ in name and in a flavor tag (stored inside JSON text).
		{Name: "100%阿拉比卡拼配", Origin: "巴西", ProcessMethod: "natural", FlavorTags: `["100%_特调"]`, Description: "百分百纯阿拉比卡"},
		{Name: "AA_top微批次", Origin: "埃塞俄比亚", ProcessMethod: "washed", FlavorTags: `["花香"]`, Description: "微批次精选"},
		{Name: `日晒\批次标注`, Origin: "哥伦比亚", ProcessMethod: "natural", FlavorTags: `["发酵"]`, Description: `批次 a_b\c 反斜杠`},
		{Name: "测试豆08", Origin: "巴西", ProcessMethod: "washed", FlavorTags: `["榛果"]`, Description: "填充豆"},
		{Name: "测试豆09", Origin: "巴西", ProcessMethod: "washed", FlavorTags: `["可可"]`, Description: "填充豆"},
		{Name: "测试豆10", Origin: "哥伦比亚", ProcessMethod: "honey", FlavorTags: `["橙花"]`, Description: "填充豆"},
		{Name: "测试豆11", Origin: "埃塞俄比亚", ProcessMethod: "natural", FlavorTags: `["柠檬"]`, Description: "填充豆"},
		{Name: "测试豆12", Origin: "印度尼西亚", ProcessMethod: "natural", FlavorTags: `["雪松"]`, Description: "填充豆"},
		{Name: "测试豆13", Origin: "哥斯达黎加", ProcessMethod: "honey", FlavorTags: `["蜂蜜"]`, Description: "填充豆"},
	}
	if err := db.Create(&beans).Error; err != nil {
		t.Fatalf("seed beans: %v", err)
	}
	return beans
}

// NewHarness provisions a fresh temp directory, opens the pooled on-disk DB,
// installs the fault plugin, runs migrations, seeds data and starts the real router.
func NewHarness(t *testing.T) *Harness {
	t.Helper()
	dir := t.TempDir()
	h := &Harness{t: t, dir: dir, dbPath: filepath.Join(dir, "app.db")}
	h.open(seedUsers, seedBeans)
	return h
}

func (h *Harness) open(seedUsersFn func(*testing.T, *gorm.DB) []*model.User, seedBeansFn func(*testing.T, *gorm.DB) []model.CoffeeBean) {
	var sqlDB *sql.DB
	h.DB, sqlDB = openPooledDB(h.t, h.dbPath)
	h.sqlDB = sqlDB
	h.Fault = NewFault()
	h.Fault.Install(h.DB)
	migrateAll(h.t, h.DB)

	h.Cfg = &config.Config{
		JWTSecret: "test-jwt-secret", JWTExpire: 2 * time.Hour,
		RateLimitReq: 100000, RateLimitWin: time.Minute,
		UploadDir: h.dir,
	}

	if seedUsersFn != nil {
		h.Users = seedUsersFn(h.t, h.DB)
	}
	if seedBeansFn != nil {
		h.Beans = seedBeansFn(h.t, h.DB)
	}
	if h.Users == nil {
		// Reload existing users (restart path keeps ids 1..3).
		h.Users = []*model.User{{ID: 1}, {ID: 2}, {ID: 3}}
	}

	gin.SetMode(gin.TestMode)
	h.Router = router.Setup(h.Cfg, h.DB, slog.Default())
	h.Server = httptest.NewServer(h.Router)
}

// Restart closes the server/DB and reopens the SAME on-disk file without seeding,
// proving persistence across a real process-style restart (new pool, new router).
func (h *Harness) Restart() *Harness {
	h.Server.Close()
	if h.sqlDB != nil {
		_ = h.sqlDB.Close()
	}
	nh := &Harness{t: h.t, dir: h.dir, dbPath: h.dbPath}
	nh.open(nil, nil)
	if err := nh.DB.Order("id ASC").Find(&nh.Beans).Error; err != nil {
		h.t.Fatalf("restart load beans: %v", err)
	}
	return nh
}

// Close stops the server, closes the pool and removes the temp directory.
func (h *Harness) Close() {
	if h.Server != nil {
		h.Server.Close()
	}
	if h.sqlDB != nil {
		_ = h.sqlDB.Close()
	}
	_ = os.RemoveAll(h.dir)
}

// Token issues a JWT for a seeded user index (0 admin, 1 barista, 2 roaster).
func (h *Harness) Token(userIdx int) string {
	u := h.Users[userIdx]
	tok, err := util.GenerateToken(u.ID, u.Username, u.Role, h.Cfg.JWTSecret, h.Cfg.JWTExpire)
	if err != nil {
		h.t.Fatalf("token: %v", err)
	}
	return tok
}

// Do performs a request against the real server and returns status + envelope.
func (h *Harness) Do(method, path, token string, body any) (int, *Envelope) {
	h.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.Router.ServeHTTP(rec, req)

	env := &Envelope{}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), env); err != nil {
			h.t.Fatalf("decode response %s: %v; body=%s", path, err, rec.Body.String())
		}
	}
	return rec.Code, env
}

// Decode unmarshals the envelope data into v.
func (h *Harness) Decode(env *Envelope, v any) {
	h.t.Helper()
	if err := json.Unmarshal(env.Data, v); err != nil {
		h.t.Fatalf("decode data: %v; raw=%s", err, string(env.Data))
	}
}
