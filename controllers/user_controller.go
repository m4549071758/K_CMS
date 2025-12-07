package controllers

import (
	"k-cms/config"
	"k-cms/middlewares"
	"k-cms/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetOwner はサイトのオーナー（最初のユーザー）の公開プロフィールを取得する
func GetOwner(c *gin.Context) {
	var user models.User
	// 最初のユーザーを取得
	if err := config.DB.First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}

	// メールアドレスなどの秘密情報を除外したレスポンス構造体
	type PublicProfile struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		Bio        string `json:"bio"`
		GithubUrl  string `json:"github_url"`
		TwitterUrl string `json:"twitter_url"`
		QiitaUrl   string `json:"qiita_url"`
		ZennUrl    string `json:"zenn_url"`
		// Emailは含めない
	}

	response := PublicProfile{
		ID:         user.ID.String(),
		Username:   user.Username,
		Bio:        user.Bio,
		GithubUrl:  user.GithubUrl,
		TwitterUrl: user.TwitterUrl,
		QiitaUrl:   user.QiitaUrl,
		ZennUrl:    user.ZennUrl,
	}

	c.JSON(http.StatusOK, response)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User

	// プレースホルダを使用してパラメータ化クエリを実行
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
func CreateUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Register(c)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	
	// Contextからuuid.UUIDとして取得
	currentUserUUID, err := middlewares.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	currentUserID := currentUserUUID.String()

	if currentUserID != id {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to update this user"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	type UpdateUserInput struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Bio        string `json:"bio"`
		GithubUrl  string `json:"github_url"`
		TwitterUrl string `json:"twitter_url"`
		QiitaUrl   string `json:"qiita_url"`
		ZennUrl    string `json:"zenn_url"`
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Username != "" {
		updates["username"] = input.Username
	}
	if input.Email != "" {
		updates["email"] = input.Email
	}
	// 空文字での更新も許容するため、ポインタ型にするか、あるいは常に更新するか。
	// ここでは単純にフィールドが存在すれば更新するようにしたいが、Goのゼロ値問題がある。
	// JSONのomitempty挙動と合わせて、今回は「送信されたら更新」とするのが望ましいが、
	// 簡易的に全てstringなので、フロント側で制御してもらう想定で、そのままセットする。
	// ただし、空文字で消したい場合もあるので、空チェックは外すか、別途ロジックが必要。
	// 既存が != "" チェックしているので、それに倣うと「空文字にして削除」ができない。
	// ここは「空文字でも更新できる」ように修正すべきだが、既存のUsername/Emailは必須に近いので維持。
	// プロフィール系は空もまた値なり。

	// Bioなどは空文字入力で消去したいニーズがあるため、そのまま代入を検討したいが、
	// map更新形式だとゼロ値除外が面倒。
	// ShouldBindJSONで構造体に入れた時点で、送られてこなかったのか空文字なのかの区別がつかない。
	// 厳密にやるなら map[string]interface{} で受けるべき。
	// 今回は既存実装を踏襲しつつ、プロフィール項目は一括更新(PUT)の思想で、DBの既存値に上書きする形をとる。
	
	updates["bio"] = input.Bio
	updates["github_url"] = input.GithubUrl
	updates["twitter_url"] = input.TwitterUrl
	updates["qiita_url"] = input.QiitaUrl
	updates["zenn_url"] = input.ZennUrl

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	
	// Contextからuuid.UUIDとして取得
	currentUserUUID, err := middlewares.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	currentUserID := currentUserUUID.String()

	if currentUserID != id {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to delete this user"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// パスワード変更用の入力構造体
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword はユーザーのパスワードを変更する関数
func ChangePassword(c *gin.Context) {
	// JWTトークンからユーザーを取得
	user, err := middlewares.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証されていません"})
		return
	}

	// リクエストからパスワード情報を取得
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 現在のパスワードを検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "現在のパスワードが正しくありません"})
		return
	}

	// 新しいパスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードのハッシュ化に失敗しました"})
		return
	}

	// パスワードを更新
	if err := config.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パスワードが正常に変更されました"})
}
