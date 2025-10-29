package notes

import (
	"context"
	"path/filepath"
    "runtime"
	"os"
	"testing"
	"time"
	"log"

	"example.com/pz8-mongo/internal/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/joho/godotenv"
)


func init() {

	_, filename, _, _ := runtime.Caller(0)
    dir := filepath.Join(filepath.Dir(filename), "../..")
    envPath := filepath.Join(dir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Println("Warning: .env file not found")
	}
}

func getTestMongoURI() string {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	return uri
}

func TestCreateAndGet(t *testing.T) {
	ctx := context.Background()
	
	dbName := "pz8_test_" + primitive.NewObjectID().Hex()
	deps, err := db.ConnectMongo(ctx, getTestMongoURI(), dbName)
	if err != nil {
		t.Fatal("connect:", err)
	}
	
	t.Cleanup(func() {
		deps.Database.Drop(ctx)
		deps.Client.Disconnect(ctx)
	})

	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("new repo:", err)
	}

	// Test Create
	created, err := r.Create(ctx, "Test Title", "Test Content")
	if err != nil {
		t.Fatal("create:", err)
	}

	// Проверяем что ID установлен
	if created.ID.IsZero() {
		t.Fatal("expected ID to be set")
	}

	// Проверяем поля
	if created.Title != "Test Title" {
		t.Fatalf("want title 'Test Title', got '%s'", created.Title)
	}
	if created.Content != "Test Content" {
		t.Fatalf("want content 'Test Content', got '%s'", created.Content)
	}

	// Проверяем временные метки (не сравниваем точное время)
	if created.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
	if created.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}

	// Test ByID
	got, err := r.ByID(ctx, created.ID.Hex())
	if err != nil {
		t.Fatal("by id:", err)
	}

	// Проверяем что получили ту же запись
	if got.ID != created.ID {
		t.Fatalf("want ID %s, got %s", created.ID.Hex(), got.ID.Hex())
	}
	if got.Title != "Test Title" {
		t.Fatalf("want title 'Test Title', got '%s'", got.Title)
	}
	if got.Content != "Test Content" {
		t.Fatalf("want content 'Test Content', got '%s'", got.Content)
	}
	
	// Сравниваем время с допуском (временные зоны могут отличаться)
	timeDiff := got.CreatedAt.Sub(created.CreatedAt)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Fatalf("CreatedAt differs too much: want %v, got %v", created.CreatedAt, got.CreatedAt)
	}
}

func TestByID_NotFound(t *testing.T) {
	ctx := context.Background()
	
	dbName := "pz8_test_" + primitive.NewObjectID().Hex()
	deps, err := db.ConnectMongo(ctx, getTestMongoURI(), dbName)
	if err != nil {
		t.Fatal("connect:", err)
	}
	
	t.Cleanup(func() {
		deps.Database.Drop(ctx)
		deps.Client.Disconnect(ctx)
	})

	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("new repo:", err)
	}

	// Пытаемся получить несуществующую запись
	_, err = r.ByID(ctx, primitive.NewObjectID().Hex())
	if err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestCreate_DuplicateTitle(t *testing.T) {
	ctx := context.Background()
	
	dbName := "pz8_test_" + primitive.NewObjectID().Hex()
	deps, err := db.ConnectMongo(ctx, getTestMongoURI(), dbName)
	if err != nil {
		t.Fatal("connect:", err)
	}
	
	t.Cleanup(func() {
		deps.Database.Drop(ctx)
		deps.Client.Disconnect(ctx)
	})

	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("new repo:", err)
	}

	// Создаём первую запись
	_, err = r.Create(ctx, "Duplicate Title", "Content 1")
	if err != nil {
		t.Fatal("create first:", err)
	}

	// Пытаемся создать запись с тем же заголовком
	_, err = r.Create(ctx, "Duplicate Title", "Content 2")
	if err == nil {
		t.Fatal("expected error for duplicate title")
	}
}

func TestList(t *testing.T) {
	ctx := context.Background()
	
	dbName := "pz8_test_" + primitive.NewObjectID().Hex()
	deps, err := db.ConnectMongo(ctx, getTestMongoURI(), dbName)
	if err != nil {
		t.Fatal("connect:", err)
	}
	
	t.Cleanup(func() {
		deps.Database.Drop(ctx)
		deps.Client.Disconnect(ctx)
	})

	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("new repo:", err)
	}

	// Создаём несколько записей с разными паттернами названий
	notes := []struct {
		title   string
		content string
	}{
		{"Note One", "Content 1"},
		{"Note Two", "Content 2"},
		{"Another Task", "Content 3"}, // Это НЕ должно находиться по поиску "Note"
	}

	for _, n := range notes {
		_, err := r.Create(ctx, n.title, n.content)
		if err != nil {
			t.Fatal("create:", err)
		}
		// Небольшая задержка для разных временных меток
		time.Sleep(10 * time.Millisecond)
	}

	// Test List all
	all, err := r.List(ctx, "", 10, 0)
	if err != nil {
		t.Fatal("list all:", err)
	}
	if len(all) != 3 {
		t.Fatalf("want 3 notes, got %d", len(all))
	}

	// Проверяем сортировку по времени (последняя созданная - первая)
	if all[0].Title != "Another Task" {
		t.Fatalf("want latest note first, got '%s'", all[0].Title)
	}

	// Test List with search - должен найти только "Note One" и "Note Two"
	search, err := r.List(ctx, "Note", 10, 0)
	if err != nil {
		t.Fatal("list search:", err)
	}
	if len(search) != 2 {
		t.Fatalf("want 2 notes with 'Note', got %d. Found: %v", len(search), getTitles(search))
	}
	
	// Проверяем что нашли правильные записи
	foundTitles := getTitles(search)
	expectedTitles := []string{"Note Two", "Note One"} // в порядке создания (новые first)
	
	for i, expected := range expectedTitles {
		if foundTitles[i] != expected {
			t.Fatalf("want title '%s' at position %d, got '%s'", expected, i, foundTitles[i])
		}
	}
}

// Вспомогательная функция для отладки
func getTitles(notes []Note) []string {
	titles := make([]string, len(notes))
	for i, n := range notes {
		titles[i] = n.Title
	}
	return titles
}