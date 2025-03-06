package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// Expression представляет арифметическое выражение
type Expression struct {
	ID     string  `json:"id"`     // Уникальный идентификатор
	Status string  `json:"status"` // Статус: "pending", "processing", "done"
	Result float64 `json:"result"` // Результат вычисления
}

// Task представляет задачу для вычисления
type Task struct {
	ID            string  `json:"id"`             // Уникальный идентификатор задачи
	Arg1          float64 `json:"arg1"`           // Первый аргумент
	Arg2          float64 `json:"arg2"`           // Второй аргумент
	Operation     string  `json:"operation"`      // Операция: "+", "-", "*", "/"
	OperationTime int     `json:"operation_time"` // Время выполнения операции (в мс)
	Result        float64 `json:"result"`         // Результат выполнения задачи
}

var (
	expressions = make(map[string]*Expression) // Хранилище выражений
	tasks       = make(map[string]*Task)       // Хранилище задач
	mu          sync.RWMutex                   // Мьютекс для потокобезопасности
)

// AddExpression добавляет новое выражение в хранилище
func AddExpression(expr *Expression) string {
	mu.Lock()
	defer mu.Unlock()
	id := uuid.New().String() // Генерация уникального ID
	expr.ID = id
	expressions[id] = expr
	return id
}

// GetExpression возвращает выражение по ID
func GetExpression(id string) (*Expression, bool) {
	mu.RLock()
	defer mu.RUnlock()
	expr, exists := expressions[id]
	return expr, exists
}

// GetAllExpressions возвращает все выражения
func GetAllExpressions() []*Expression {
	mu.RLock()
	defer mu.RUnlock()
	var result []*Expression
	for _, expr := range expressions {
		result = append(result, expr)
	}
	return result
}

// AddTask добавляет новую задачу в хранилище
func AddTask(task *Task) string {
	mu.Lock()
	defer mu.Unlock()
	id := uuid.New().String() // Генерация уникального ID
	task.ID = id
	tasks[id] = task
	return id
}

// GetTask возвращает задачу по ID
func GetTask(id string) (*Task, bool) {
	mu.RLock()
	defer mu.RUnlock()
	task, exists := tasks[id]
	return task, exists
}

// DeleteTask удаляет задачу по ID
func DeleteTask(id string) {
	mu.Lock()
	defer mu.Unlock()
	log.Printf("Deleting task ID: %s\n", id)
	if _, exists := tasks[id]; exists {
		delete(tasks, id)
		log.Printf("Task ID: %s deleted successfully\n", id)
	} else {
		log.Printf("Task ID: %s not found\n", id)
	}
}

// precedence возвращает приоритет операции
func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func parseExpression(expr string) ([]*Task, error) {
	var tasks []*Task
	var stack []string
	var output []string
	expr = strings.ReplaceAll(expr, " ", "")
	addTask := func(op string) {
		if len(output) < 2 {
			log.Println("Invalid expression: not enough operands")
			return
		}
		arg2 := output[len(output)-1]
		arg1 := output[len(output)-2]
		output = output[:len(output)-2]

		task := &Task{
			ID:            uuid.New().String(),
			Arg1:          parseOperand(arg1),
			Arg2:          parseOperand(arg2),
			Operation:     op,
			OperationTime: 1000,
			Result:        0,
		}
		tasks = append(tasks, task)
		output = append(output, fmt.Sprintf("task%d", len(tasks)-1))
	}

	for i := 0; i < len(expr); i++ {
		char := string(expr[i])
		switch {
		case char >= "0" && char <= "9" || char == ".":
			j := i
			for j < len(expr) && (expr[j] >= '0' && expr[j] <= '9' || expr[j] == '.') {
				j++
			}
			output = append(output, expr[i:j])
			i = j - 1
		case char == "(":
			stack = append(stack, char)
		case char == ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				addTask(stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			stack = stack[:len(stack)-1]
		default:
			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(char) {
				addTask(stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, char)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" || stack[len(stack)-1] == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		addTask(stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	if len(output) != 1 {
		return nil, fmt.Errorf("invalid expression")
	}

	return tasks, nil
}

// parseOperand преобразует операнд в число или ссылку на результат задачи
func parseOperand(operand string) float64 {
	if strings.HasPrefix(operand, "task") {
		// Если это ссылка на задачу, возвращаем 0 (значение будет заменено позже)
		return 0
	}
	// Если это число, парсим его
	val, _ := strconv.ParseFloat(operand, 64)
	return val
}

func isValidExpression(expr string) bool {
	// Проверка на допустимые символы
	for _, char := range expr {
		if !strings.ContainsRune("0123456789.+-*/() ", char) {
			return false
		}
	}
	return true
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	// Проверка на пустое выражение
	if req.Expression == "" {
		http.Error(w, "Expression is required", http.StatusUnprocessableEntity)
		return
	}

	// Проверка на валидность выражения
	if !isValidExpression(req.Expression) {
		http.Error(w, "Invalid characters in expression", http.StatusUnprocessableEntity)
		return
	}

	// Проверка на деление на ноль
	if strings.Contains(req.Expression, "/ 0") {
		http.Error(w, "Division by zero is not allowed", http.StatusUnprocessableEntity)
		return
	}

	// Создаем новое выражение
	expr := &Expression{
		Status: "pending", // По умолчанию статус "ожидает"
		Result: 0,         // Результат пока неизвестен
	}
	id := AddExpression(expr)

	// Разбиваем выражение на задачи
	tasks, err := parseExpression(req.Expression)
	if err != nil {
		// Внутренняя ошибка сервера
		log.Printf("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, task := range tasks {
		AddTask(task)
		log.Printf("Task added: %+v\n", task)
	}

	// Возвращаем ID выражения
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// ExpressionsHandler возвращает список всех выражений
func ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	exprs := GetAllExpressions()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]*Expression{"expressions": exprs})
}

// ExpressionByIDHandler возвращает выражение по его ID
func ExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	expr, exists := GetExpression(id)
	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]*Expression{"expression": expr})
}

// TaskHandler возвращает задачу для агента
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	// Поиск задачи:
	for _, task := range tasks {
		// Возвращаем первую найденную задачу
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]*Task{"task": task})
		return
	}

	// Если задач нет, возвращаем 404
	http.Error(w, "No tasks available", http.StatusNotFound)
}

// TaskResultHandler принимает результат выполнения задачи от агента
func TaskResultHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}
	log.Printf("Received result for task ID: %s, result: %f\n", req.ID, req.Result)
	//Удаление задачи после выполнения
	//Когда агент выполняет задачу и отправляет результат, задачу нужно удалить из хранилища
	DeleteTask(req.ID)

	// Возвращаем 200 OK
	w.WriteHeader(http.StatusOK)
}

func init() {
	// Добавляем тестовую задачу
	AddTask(&Task{
		Arg1:          2,
		Arg2:          3,
		Operation:     "+",
		OperationTime: 1000, // 1 секунда
	})
	log.Println("Test task added!")
}

// Регистрируем обработчики
func main() {
	//для добавления нового выражения
	http.HandleFunc("/api/v1/calculate", CalculateHandler)
	//для получения списка всех выражений
	http.HandleFunc("/api/v1/expressions", ExpressionsHandler)
	//для получения выражения по его ID
	http.HandleFunc("/api/v1/expressions/", ExpressionByIDHandler)

	//для получения задачи агентом
	http.HandleFunc("/internal/task", TaskHandler)
	//для отправки результата задачи агентом
	http.HandleFunc("/internal/task/result", TaskResultHandler)

	// Запускаем сервер
	log.Println("Starting orchestrator on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
