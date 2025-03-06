package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Task представляет задачу для вычисления
type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

// compute выполняет вычисление задачи
func compute(task Task) (result float64, err error) {

	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 != 0 {
			result = task.Arg1 / task.Arg2
		} else {
			result = 0 // Обработка деления на ноль
			log.Printf("Произошло деление на ноль упс")
			return 0, errors.New("Division by zero")
		}
	default:
		result = 0
	}

	// Имитируем задержку выполнения операции
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	return result, nil
}

// worker представляет собой вычислителя, который получает задачи и выполняет их
func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Получаем задачу от оркестратора
		resp, err := http.Get("http://localhost:8080/internal/task")
		if err != nil {
			log.Printf("Worker %d: failed to fetch task: %v\n", id, err)
			time.Sleep(5 * time.Second) // Пауза перед повторной попыткой
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Worker %d: no tasks available\n", id)
			time.Sleep(5 * time.Second) // Пауза перед повторной попыткой
			continue
		}

		var taskResponse struct {
			Task Task `json:"task"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&taskResponse); err != nil {
			log.Printf("Worker %d: failed to decode task: %v\n", id, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		task := taskResponse.Task
		log.Printf("Worker %d: received task %s (%f %s %f)\n", id, task.ID, task.Arg1, task.Operation, task.Arg2)
		log.Printf("Worker %d: full task response: %+v\n", id, taskResponse)
		// Выполняем вычисление
		result, err := compute(task)
		if err != nil {
			log.Printf("Error")
			continue
		}
		log.Printf("Worker %d: computed result for task %s: %f\n", id, task.ID, result)

		// Отправляем результат обратно оркестратору
		resultData := map[string]interface{}{
			"id":     task.ID,
			"result": result,
		}
		resultJSON, _ := json.Marshal(resultData)

		//Потом удалить этот лог перебор уже ->
		log.Printf("Worker %d: sending result: %s\n", id, string(resultJSON))

		resp, err = http.Post("http://localhost:8080/internal/task/result", "application/json", strings.NewReader(string(resultJSON)))
		if err != nil {
			log.Printf("Worker %d: failed to send result: %v\n", id, err)
			continue
		}
		resp.Body.Close()

		log.Printf("Worker %d: result for task %s sent successfully\n", id, task.ID)
	}
}

func main() {
	// Получаем количество горутин из переменной окружения
	computingPowerStr := os.Getenv("COMPUTING_POWER")
	if computingPowerStr == "" {
		computingPowerStr = "1" // По умолчанию 1 горутина  МОЖЕТ 3 ЗАДАТЬ И НЕ МУЧАТЬСЯ?????
	}
	computingPower, err := strconv.Atoi(computingPowerStr)
	if err != nil {
		log.Fatalf("Invalid COMPUTING_POWER value: %v\n", err)
	}

	log.Printf("Starting agent with %d workers\n", computingPower)

	var wg sync.WaitGroup
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	wg.Wait() // Бесконечный цикл, так как горутины не завершаются
}
