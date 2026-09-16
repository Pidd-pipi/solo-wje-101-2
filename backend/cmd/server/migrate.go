package main

import (
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.TastingNote{},
		&model.BrewRecipe{},
		&model.CoffeeBean{},
		&model.Comment{},
		&model.Like{},
		&model.UserFollow{},
	)
}

func seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	logger := slog.Default()

	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	userHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	admin := &model.User{Username: "admin", Email: "admin@coffeetaste.local", PasswordHash: string(adminHash), Bio: "咖啡平台管理员", Role: "admin"}
	user := &model.User{Username: "barista", Email: "barista@coffeetaste.local", PasswordHash: string(userHash), Bio: "精品咖啡爱好者", Role: "user"}
	user2 := &model.User{Username: "roaster", Email: "roaster@coffeetaste.local", PasswordHash: string(userHash), Bio: "烘焙师", Role: "user"}
	if err := db.Create(admin).Error; err != nil {
		return err
	}
	if err := db.Create(user).Error; err != nil {
		return err
	}
	if err := db.Create(user2).Error; err != nil {
		return err
	}

	beans := []model.CoffeeBean{
		{Name: "埃塞俄比亚耶加雪菲", Origin: "埃塞俄比亚", ProcessMethod: "washed", FlavorTags: `["柑橘","茉莉","蜂蜜"]`, Description: "经典水洗耶加雪菲，明亮柑橘酸质。"},
		{Name: "哥伦比亚慧兰", Origin: "哥伦比亚", ProcessMethod: "washed", FlavorTags: `["坚果","焦糖","红苹果"]`, Description: "平衡甜感，坚果与焦糖尾韵。"},
		{Name: "哥斯达黎加蜜处理", Origin: "哥斯达黎加", ProcessMethod: "honey", FlavorTags: `["莓果","红糖","葡萄干"]`, Description: "蜜处理带来醇厚甜感与莓果香气。"},
		{Name: "印尼曼特宁", Origin: "印度尼西亚", ProcessMethod: "washed", FlavorTags: `["草本","黑巧克力","香料"]`, Description: "醇厚浓郁，草本与黑巧风味。"},
	}
	if err := db.Create(&beans).Error; err != nil {
		return err
	}

	recipes := []model.BrewRecipe{
		{UserID: user.ID, Name: "手冲三段式", Device: "手冲壶", WaterTemp: 92, GrindSize: "中细", Ratio: "1:15", Steps: `[{"step_number":1,"description":"闷蒸30秒","duration_seconds":30},{"step_number":2,"description":"第一段注水至150ml","duration_seconds":20},{"step_number":3,"description":"第二段注水至300ml","duration_seconds":30}]`},
		{UserID: user2.ID, Name: "法压壶经典", Device: "法压壶", WaterTemp: 94, GrindSize: "中粗", Ratio: "1:14", Steps: `[{"step_number":1,"description":"注水并搅拌","duration_seconds":10},{"step_number":2,"description":"浸泡4分钟","duration_seconds":240},{"step_number":3,"description":"缓慢压杆","duration_seconds":15}]`},
	}
	if err := db.Create(&recipes).Error; err != nil {
		return err
	}

	notes := []model.TastingNote{
		{UserID: user.ID, CoffeeName: "埃塞俄比亚耶加雪菲", Origin: "埃塞俄比亚", RoastLevel: "light", FlavorTags: `["柑橘","茉莉"]`, AromaScore: 8.5, AcidityScore: 8.0, BodyScore: 7.0, OverallScore: 8.3, BrewMethod: "手冲", BrewRecipeID: recipes[0].ID, NotesText: "花香明显，柑橘酸质明亮，回甘持久。", ImageURL: "https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=600"},
		{UserID: user2.ID, CoffeeName: "哥伦比亚慧兰", Origin: "哥伦比亚", RoastLevel: "medium", FlavorTags: `["坚果","焦糖"]`, AromaScore: 7.5, AcidityScore: 6.8, BodyScore: 7.8, OverallScore: 7.6, BrewMethod: "法压", BrewRecipeID: recipes[1].ID, NotesText: "甜感平衡，坚果香气浓郁。", ImageURL: "https://images.unsplash.com/photo-1447933601403-0c6688de566e?w=600"},
		{UserID: user.ID, CoffeeName: "哥斯达黎加蜜处理", Origin: "哥斯达黎加", RoastLevel: "medium", FlavorTags: `["莓果","红糖"]`, AromaScore: 8.0, AcidityScore: 7.2, BodyScore: 8.0, OverallScore: 7.9, BrewMethod: "手冲", NotesText: "莓果酸甜与红糖甜感交织。", ImageURL: "https://images.unsplash.com/photo-1517701604599-bb29b565090c?w=600"},
	}
	if err := db.Create(&notes).Error; err != nil {
		return err
	}

	comments := []model.Comment{
		{NoteID: notes[0].ID, UserID: user2.ID, Content: "我也很喜欢这只耶加雪菲，柑橘调太棒了！"},
		{NoteID: notes[1].ID, UserID: user.ID, Content: "慧兰做奶咖也很合适。"},
	}
	if err := db.Create(&comments).Error; err != nil {
		return err
	}

	likes := []model.Like{
		{UserID: user2.ID, NoteID: notes[0].ID},
		{UserID: user.ID, NoteID: notes[1].ID},
		{UserID: admin.ID, NoteID: notes[0].ID},
	}
	if err := db.Create(&likes).Error; err != nil {
		return err
	}

	follows := []model.UserFollow{
		{FollowerID: user.ID, FollowingID: user2.ID},
		{FollowerID: user2.ID, FollowingID: user.ID},
	}
	if err := db.Create(&follows).Error; err != nil {
		return err
	}

	logger.Info("wjecoffeetaste seed data created",
		"users", 3, "beans", len(beans), "recipes", len(recipes), "notes", len(notes), "comments", len(comments))
	return nil
}
