# 1) Как запустить:
- Скачивайте все зип архивом или подобным образом к себе в VS Code например.
- Далее проверьте импорты. Если у вас нет библиотеки uuid, то скачайте ее:
в терминал проекта: go get github.com/google/uuid
-В одном терминале запускаем оркестратора:
 ![image](https://github.com/user-attachments/assets/a5228596-d743-4f28-982d-ff9005e55cf9)

-Во втором терминале запускаем агента:
 ![image](https://github.com/user-attachments/assets/e40e8443-bf81-4366-aa82-2e605f1d6939)


Итог: все запустилось, видим по логам, как отработала тестовая задача, которая была добавлена вручную.

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

# 2)	Работа ОРКЕСТРАТОРА:
Тестирование с помощью Postman 
# - Добавление вычисления арифметического выражения 'localhost/api/v1/calculate'
{
  "expression": <строка с выражение>
}'
Тело ответа
{
    "id": <уникальный идентификатор выражения>
}

 ![image](https://github.com/user-attachments/assets/c3307a42-2c07-4d3f-a55b-a8a0713099c0)

Код ответа 201 - выражение принято для вычисления.

 ![image](https://github.com/user-attachments/assets/f0e36664-6c46-4e59-bfb0-d0fbee56b5d0)

Код ответа 422 – невалидные данные.
 ![image](https://github.com/user-attachments/assets/a3b17c29-de39-411e-bebe-a385abbf110a)

Код ответа 500 - что-то пошло не так.
# Логи в оркестраторе:
	 ![image](https://github.com/user-attachments/assets/8ea3e4f1-9a5f-4f2b-bdd1-ba5e5f4b45a6)

# Логи в агенте:
 ![image](https://github.com/user-attachments/assets/be77114d-4e61-402f-8156-9021a4c9dc1a)

# -Получение списка выражений 'localhost/api/v1/expressions' (+ было добавлено 3 * 4)
 ![image](https://github.com/user-attachments/assets/adfeecd7-5b5b-410a-9674-414fc34c3762)

![image](https://github.com/user-attachments/assets/4c89ec56-3911-471d-a30f-ae3b7e74922f)

200	- успешно получен список выражений

# -	Получение выражения по его идентификатору  'localhost/api/v1/expressions/:id'
 ![image](https://github.com/user-attachments/assets/b16cb3b1-9bb0-4519-88b0-1b00be17c58f)


# -	Получение задачи для выполнения curl --location 'localhost/internal/task'
![image](https://github.com/user-attachments/assets/36d3c1b9-52bb-4873-a353-32030e1f4e09)



# 3)	Также был начат фронт, но кнопки не подружены
![image](https://github.com/user-attachments/assets/fdd439dd-a0bb-4830-978f-537e7c2d9349)
 

