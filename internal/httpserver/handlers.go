package httpserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/zkhrg/go_day03/internal/elasticsearch"
)

type PlacesPageContent struct {
	Total    int
	Items    []elasticsearch.PlacesHit
	Page     int
	NextPage int
	PrevPage int
	LastPage int
}

type PlacesAPISchema struct {
	Name     string        `json:"name"`
	Total    int           `json:"total"`
	Places   []PlaceSchema `json:"places"`
	PrevPage int           `json:"prev_page"`
	NextPage int           `json:"next_page"`
	LastPage int           `json:"last_page"`
}

type PlaceSchema struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
}

// Структура заголовка JWT
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Структура полезной нагрузки JWT (claims)
type JWTClaims struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

// Секретный ключ для подписи JWT
var jwtKey = []byte(os.Getenv("SECRET_KEY"))

func base64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// я знаю что это не лучшее решение и можно иначе кэшировать страницы
// мне на первых порах кажется что это лучше чем воротить зверинец из
// технологий на 4 дне бассейна и сделать хардкодом чтобы у меня не падал
// по ООМу эластик
var pageCacheMap map[int]PlacesPageContent

const pageSize = 11

// StartPageHandler обрабатывает только GET-запросы на /
func StartPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}
	var pgc PlacesPageContent
	total := elasticsearch.CountIndexRecords("places")
	pages := total / pageSize
	if total%pageSize != 0 {
		pages += 1
	}
	if len(pageCacheMap) == 0 {
		pageCacheMap = make(map[int]PlacesPageContent)
	}
	tmpl, _ := template.ParseFiles("templates/index.html")

	page, err := validatePageParam(r.URL.Query().Get("page"), pages, w)
	if err != nil {
		return
	}

	if val, ok := pageCacheMap[page]; ok {
		tmpl.Execute(w, val)
		return
	}
	items := elasticsearch.GetPageData(page, pageSize, "places")

	pgc = PlacesPageContent{
		Total:    total,
		Items:    items,
		Page:     page,
		PrevPage: page - 1,
		NextPage: page + 1,
		LastPage: pages,
	}

	if err := tmpl.Execute(w, pgc); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Template execution error:", err)
		return
	}
	pageCacheMap[page] = pgc
}

func SimpleAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}

	total := elasticsearch.CountIndexRecords("places")
	pages := total / pageSize
	if total%pageSize != 0 {
		pages += 1
	}

	page, err := validatePageParam(r.URL.Query().Get("page"), pages, w)
	if err != nil {
		return
	}
	pas := PlacesAPISchema{
		Total:    total,
		LastPage: pages,
		PrevPage: page - 1,
		NextPage: page + 1,
		Name:     "palces",
		Places:   make([]PlaceSchema, 0),
	}
	data := elasticsearch.GetPageData(page, pageSize, "places")
	for _, v := range data {
		ps := PlaceSchema{
			ID:      v.Source.ID,
			Name:    v.Source.Name,
			Address: v.Source.Address,
			Phone:   v.Source.Phone,
		}
		ps.Location.Lat = v.Source.Location.Lat
		ps.Location.Lon = v.Source.Location.Lon
		pas.Places = append(pas.Places, ps)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pas)
}

func validatePageParam(page string, pages int, w http.ResponseWriter) (int, error) {
	if page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt > pages || pageInt < 1 {
			w.Write([]byte(fmt.Sprintf("Invalid 'page' value: '%v'", page)))
			fmt.Println("as")
			return 0, errors.New("invalid page value")
		}
		return pageInt, nil
	}
	w.Write([]byte("no page argument provided"))
	return 0, errors.New("no page argument provided")
}

func APIRecommendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
		return
	}
	lat, lon, err := validateLatLonParams(r.URL.Query().Get("lat"), r.URL.Query().Get("lon"), w)
	if err != nil {
		return
	}
	response := elasticsearch.GetNearestPlaces(lat, lon, "places")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func validateLatLonParams(lat string, lon string, w http.ResponseWriter) (float64, float64, error) {
	if lat != "" && lon != "" {
		latFloat, latErr := strconv.ParseFloat(lat, 64)
		lonFloat, lonErr := strconv.ParseFloat(lon, 64)
		if latErr != nil || lonErr != nil {
			w.Write([]byte("Invalid value"))
			return 0, 0, errors.New("invalid page value")
		}
		return latFloat, lonFloat, nil
	}
	w.Write([]byte("lat or lon argument no provided"))
	return 0, 0, errors.New("lat or lon argument no provided")
}

func createJWT(username string) (string, error) {
	// Создание заголовка JWT
	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64Encode(headerBytes)

	// Создание полезной нагрузки (claims)
	claims := JWTClaims{
		Username: username,
		Exp:      time.Now().Add(5 * time.Minute).Unix(),
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64Encode(claimsBytes)

	// Формирование unsigned-токена
	unsignedToken := fmt.Sprintf("%s.%s", headerEncoded, claimsEncoded)

	// Создание подписи с использованием HMAC-SHA256
	h := hmac.New(sha256.New, jwtKey)
	h.Write([]byte(unsignedToken))
	signature := base64Encode(h.Sum(nil))

	// Формирование итогового токена
	token := fmt.Sprintf("%s.%s", unsignedToken, signature)
	return token, nil
}

func generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	token, err := createJWT(username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Извлечение токена из заголовка Authorization
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}
	token := parts[1]

	// Проверка токена
	claims, err := validateJWT(token)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Если токен валиден, предоставляем доступ к защищенному ресурсу
	w.Write([]byte(fmt.Sprintf("Hello, %s! You have access to the protected endpoint.", claims.Username)))
}

func validateJWT(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerAndClaims := strings.Join(parts[:2], ".")
	signatureProvided := parts[2]

	// Верификация подписи
	h := hmac.New(sha256.New, jwtKey)
	h.Write([]byte(headerAndClaims))
	signatureExpected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if signatureProvided != signatureExpected {
		return nil, errors.New("invalid token signature")
	}

	// Декодирование claims
	claimsBytes, err := base64Decode(parts[1])
	if err != nil {
		return nil, err
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	// Проверка срока действия токена
	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func base64Decode(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}
