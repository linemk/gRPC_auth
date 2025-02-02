package jwt

import (
	"github.com/golang-jwt/jwt/v5"                       // Подключение библиотеки для работы с JWT токенами
	"github.com/linemk/gRPC_auth/internal/domain/models" // Подключение моделей приложения
	"time"                                               // Подключение пакета для работы с временем
)

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) // Создание нового JWT токена с алгоритмом подписи HS256
	claims := token.Claims.(jwt.MapClaims)   // Инициализация claims (данных, содержащихся в токене) как MapClaims

	claims["uid"] = user.ID                         // Установка идентификатора пользователя в claims
	claims["email"] = user.Email                    // Установка электронной почты пользователя в claims
	claims["exp"] = time.Now().Add(duration).Unix() // Установка времени истечения срока действия токена
	claims["app_id"] = app.ID                       // Установка идентификатора приложения в claims

	// Подписание токена с использованием секрета приложения
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil { // Обработка ошибки, если подпись токена не удалась
		return "", err
	}
	return tokenString, nil // Возвращение подписанного токена
}
