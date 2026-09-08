package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"k-cms/config"
	"k-cms/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestArticleResponsesExposeOnlyIntendedSEOFields(t *testing.T) {
	article := models.Article{
		ID:             uuid.Must(uuid.NewV7()),
		Title:          "Visible title",
		Excerpt:        "Visible excerpt",
		SeoTitle:       "Search title",
		SeoDescription: "Search description",
		PrimaryKeyword: "editor only keyword",
	}

	publicJSON, err := json.Marshal(newArticleResponse(article))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(publicJSON), "primary_keyword") {
		t.Error("public article response exposes the editor-only primary keyword")
	}
	if !strings.Contains(string(publicJSON), `"seo_title":"Search title"`) ||
		!strings.Contains(string(publicJSON), `"seo_description":"Search description"`) {
		t.Errorf("public article response omits SEO metadata: %s", publicJSON)
	}

	editorJSON, err := json.Marshal(ArticleEditorResponse{
		ArticleResponse: newArticleResponse(article),
		PrimaryKeyword:  article.PrimaryKeyword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(editorJSON), `"primary_keyword":"editor only keyword"`) {
		t.Errorf("editor article response omits primary keyword: %s", editorJSON)
	}
}

func TestGetArticleDistinguishesMissingRecordFromDatabaseFailure(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"missing article", fmt.Errorf("query: %w", gorm.ErrRecordNotFound), http.StatusNotFound},
		{"database unavailable", errors.New("private database connection error"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{
				DSN: "unused:unused@tcp(127.0.0.1:1)/unused", SkipInitializeWithVersion: true,
			}), &gorm.Config{DisableAutomaticPing: true})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := sqlDB.Close(); err != nil {
					t.Error(err)
				}
			})
			if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
				tx.AddError(tc.err)
			}); err != nil {
				t.Fatal(err)
			}
			previousDB := config.DB
			config.DB = db
			t.Cleanup(func() { config.DB = previousDB })

			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Params = gin.Params{{Key: "id", Value: "missing-article"}}
			context.Request = httptest.NewRequest(http.MethodGet, "/api/articles/missing-article", nil)
			GetArticle(context)
			if recorder.Code != tc.status {
				t.Errorf("HTTP status = %d, want %d", recorder.Code, tc.status)
			}
			if strings.Contains(recorder.Body.String(), "private database connection") {
				t.Error("response exposes a private database error")
			}
		})
	}
}
